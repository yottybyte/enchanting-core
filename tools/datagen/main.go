package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultManifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
	defaultCacheDir    = "./versions"
	defaultOutDir      = "./internal/mcdata"
)

var wantReports = []string{"registries.json", "packets.json", "blocks.json", "minecraft"}
var javaVersionRe = regexp.MustCompile(`version "([^"]+)"`)

type Manifest struct {
	Latest   ManifestLatest    `json:"latest"`
	Versions []ManifestVersion `json:"versions"`
}

type ManifestLatest struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type ManifestVersion struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
}

type VersionMeta struct {
	ID        string `json:"id"`
	Downloads struct {
		Server struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"server"`
	} `json:"downloads"`
	JavaVersion struct {
		MajorVersion int `json:"majorVersion"`
	} `json:"javaVersion"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	manifestURL := flag.String("manifest", defaultManifestURL, "URL of the Minecraft version manifest")
	version := flag.String("version", "latest", "version to generate for (id, or latest/snapshot)")
	cacheDir := flag.String("cache", defaultCacheDir, "directory for cached server jars")
	outDir := flag.String("out", defaultOutDir, "directory for generated data (per-version subfolder)")
	flag.Parse()

	ctx := context.Background()

	log.Printf("datagen: version=%q manifest=%s", *version, *manifestURL)

	m, err := getData[Manifest](ctx, *manifestURL)
	if err != nil {
		return err
	}

	ver, err := resolveVersion(m, *version)
	if err != nil {
		return err
	}

	meta, err := getData[VersionMeta](ctx, ver.URL)
	if err != nil {
		return err
	}
	log.Printf("resolved %s (type=%s, java=%d)", meta.ID, ver.Type, meta.JavaVersion.MajorVersion)

	jarPath, err := ensureJar(ctx, meta.ID, meta.Downloads.Server.URL, meta.Downloads.Server.SHA1, *cacheDir)
	if err != nil {
		return err
	}
	log.Println("server jar:", jarPath)

	if err := checkJava(ctx, meta.JavaVersion.MajorVersion); err != nil {
		return err
	}

	genOut, err := os.MkdirTemp("", "datagen-out-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(genOut) }()

	if err := runGenerator(ctx, jarPath, genOut); err != nil {
		return err
	}

	dst := filepath.Join(*outDir, meta.ID)
	if err := collectReports(filepath.Join(genOut, "reports"), dst, wantReports); err != nil {
		return err
	}

	log.Println("done:", dst)
	return nil
}

func resolveVersion(m *Manifest, version string) (ManifestVersion, error) {
	switch version {
	case "latest":
		version = m.Latest.Release
	case "snapshot":
		version = m.Latest.Snapshot
	}
	for i := range m.Versions {
		if m.Versions[i].ID == version {
			return m.Versions[i], nil
		}
	}
	return ManifestVersion{}, fmt.Errorf("version %q not found (latest release: %s)", version, m.Latest.Release)
}

func runGenerator(ctx context.Context, jarPath, outDir string) error {
	absJar, err := filepath.Abs(jarPath)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}

	work, err := os.MkdirTemp("", "datagen-work-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(work) }()

	cmd := exec.CommandContext(ctx, "java",
		"-DbundlerMainClass=net.minecraft.data.Main",
		"-jar", absJar,
		"--reports",
		"--output", absOut,
	)
	cmd.Dir = work
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func collectReports(reportsDir, outDir string, want []string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, name := range want {
		src := filepath.Join(reportsDir, name)
		dst := filepath.Join(outDir, name)

		info, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("report %s not found: %w", name, err)
		}

		if info.IsDir() {
			if err := os.RemoveAll(dst); err != nil {
				return err
			}
			if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
				return err
			}
			continue
		}

		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func ensureJar(ctx context.Context, version, url, wantSHA1, cacheDir string) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	jarPath := filepath.Join(cacheDir, version+"_"+wantSHA1+".jar")

	if sum, err := sha1File(jarPath); err == nil && strings.EqualFold(sum, wantSHA1) {
		log.Println("using cached server jar")
		return jarPath, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp(cacheDir, "server-*.jar.tmp")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	h := sha1.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), resp.Body); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	if got := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(got, wantSHA1) {
		return "", fmt.Errorf("sha1 mismatch: got %s, want %s", got, wantSHA1)
	}

	if err := os.Rename(tmpName, jarPath); err != nil {
		return "", err
	}
	return jarPath, nil
}

func checkJava(ctx context.Context, want int) error {
	got, err := javaMajor(ctx)
	if err != nil {
		return err
	}
	if got < want {
		return fmt.Errorf("java >= %d is required, but %d is installed", want, got)
	}
	return nil
}

func javaMajor(ctx context.Context) (int, error) {
	out, err := exec.CommandContext(ctx, "java", "-version").CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("java not found/won't run: %w", err)
	}
	m := javaVersionRe.FindSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("didn't understand the Java output -version: %s", out)
	}
	v := strings.TrimPrefix(string(m[1]), "1.")
	head, _, _ := strings.Cut(v, ".")
	head, _, _ = strings.Cut(head, "_")
	n, err := strconv.Atoi(head)
	if err != nil {
		return 0, fmt.Errorf("didn't understand the Java major from %q: %w", string(m[1]), err)
	}
	return n, nil
}

func sha1File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func getData[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}

	data := new(T)
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

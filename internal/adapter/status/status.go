package status

import (
	"encoding/json"

	"github.com/yottybyte/enchanting-core/internal/domain"
)

const (
	versionName     = "1.26.2"
	protocolVersion = 776
)

type Status struct {
	Version            Version     `json:"version"`
	Players            Players     `json:"players"`
	Description        Description `json:"description"`
	Favicon            string      `json:"favicon,omitempty"`
	EnforcesSecureChat bool        `json:"enforcesSecureChat"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type Players struct {
	Max    uint64   `json:"max"`
	Online int32    `json:"online"`
	Sample []Sample `json:"sample"`
}

type Sample struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Description struct {
	Text string `json:"text"`
}

func NewStatus(s domain.ServerStatus) *Status {
	return &Status{
		Version:            Version{Name: versionName, Protocol: protocolVersion},
		Players:            Players{Max: s.MaxPlayers, Online: s.OnlinePlayers, Sample: make([]Sample, 0)},
		Description:        Description{Text: s.Description},
		EnforcesSecureChat: false,
	}
}

func (s *Status) Build() (string, error) {
	data, err := json.Marshal(s)
	return string(data), err
}

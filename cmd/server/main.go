package main

import (
	"context"
	"flag"
	"log/slog"

	_network2 "github.com/yottybyte/enchanting-core/internal/adapter/network"
	"github.com/yottybyte/enchanting-core/internal/config"
	"github.com/yottybyte/enchanting-core/internal/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "config file")
	flag.Parse()

	fx.
		New(
			fx.Supply(*configPath),
			fx.Provide(config.Load),
			fx.Provide(func(cfg *config.Config) (*slog.Logger, error) {
				return logger.New(&cfg.Logger)
			}),
			fx.WithLogger(func(l *slog.Logger) fxevent.Logger {
				return &fxevent.SlogLogger{Logger: l.With("component", "uber/fx")}
			}),

			fx.Provide(func(cfg *config.Config) _network2.ConnectionTransportFactory {
				if cfg.Server.OnlineMode {
					return _network2.NewConnOnlineTransport()
				}
				return _network2.NewConnOfflineTransport()
			}),

			fx.Provide(_network2.NewServer),

			fx.Invoke(func(lc fx.Lifecycle, s *_network2.Server) {
				ctx, cancel := context.WithCancel(context.Background())
				lc.Append(fx.Hook{

					OnStart: func(_ context.Context) error {
						go func(ctx context.Context) {
							err := s.Run(ctx)
							if err != nil {
								slog.Error("failed to start server", "error", err)
								return
							}
						}(ctx)
						return nil
					},
					OnStop: func(_ context.Context) error {
						cancel()
						return nil
					},
				})
			}),
		).
		Run()
}

//func run() error {
//
//	//config.Load(*configPath)
//	//injector := do.New()
//	//do.Provide(injector, config.Load)
//	//addr := flag.String("addr", ":25565", "server address")
//	//motd := flag.String("motd", "Enchanting", "server motd")
//	//maxP := flag.Int("max", 20, "max players")
//	//flag.Parse()
//	//
//	//config.Load(*configPath)
//	//
//	//ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
//	//defer stop()
//	//
//	//log.Printf("listening on %s", *addr)
//	//
//	//jsonStatusBuild, err := status.NewStatus("26.2", 776, *motd, *maxP).Build()
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//srv := network.NewServer(*addr, jsonStatusBuild)
//	//return srv.Run(ctx)
//}

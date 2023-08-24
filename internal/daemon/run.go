package daemon

import (
	"example.com/demo/internal/daemon/config"
	genericapiserver "example.com/demo/internal/pkg/server"
)

// Run runs the specified job server. This should never exit.
func Run(cfg *config.Config) error {
	go genericapiserver.ServeHealthCheck(cfg.HealthCheckPath, cfg.HealthCheckAddress)

	server, err := createDaemonServer(cfg)
	if err != nil {
		return err
	}

	return server.PrepareRun().Run()
}

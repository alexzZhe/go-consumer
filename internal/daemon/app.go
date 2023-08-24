package daemon

import (
	"example.com/demo/internal/daemon/config"
	"example.com/demo/internal/daemon/options"
	"example.com/demo/pkg/app"
	"example.com/demo/pkg/log"
)

const commandDesc = `demo`

func NewApp(basename string) *app.App {
	opts := options.NewOptions()

	application := app.NewApp("daemon server", basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)))

	return application
}

func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		log.Init(opts.Log)
		defer log.Flush()

		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		// stopCh := genericapiserver.SetupSignalHandler()

		return Run(cfg)
	}
}

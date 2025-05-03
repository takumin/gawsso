package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/takumin/gawsso/internal/command/completion"
	"github.com/takumin/gawsso/internal/command/viewer"
	"github.com/takumin/gawsso/internal/config"
)

var (
	AppName  string = "gawsso"
	AppDesc  string = "Golang AWS SSO Identity Store Tool"
	Version  string = "unknown"
	Revision string = "unknown"
)

func main() {
	cfg := config.NewConfig()

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"l"},
			Usage:       "log level",
			Sources:     cli.EnvVars("LOG_LEVEL"),
			Value:       cfg.LogLevel,
			Destination: &cfg.LogLevel,
		},
	}

	cmds := []*cli.Command{
		completion.NewCommands(cfg, flags),
		viewer.NewCommands(cfg, flags),
	}

	app := &cli.Command{
		Name:                  AppName,
		Usage:                 AppDesc,
		Version:               fmt.Sprintf("%s (%s)", Version, Revision),
		Flags:                 flags,
		Commands:              cmds,
		EnableShellCompletion: true,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

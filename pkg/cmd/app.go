package cmd

import (
	"fmt"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/urfave/cli/v2"
	"sort"
	"strconv"
	"time"
)

func NewApp(action cli.ActionFunc, build flux.Build) *cli.App {
	app := &cli.App{
		Name:        "flux.go gateway",
		Version:     build.Version,
		Copyright:   "(c) " + strconv.Itoa(time.Now().Year()) + " bytepowered.net",
		Description: "Flux.go is a lightweight gateway for dubbo/http/grpc.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    argNameConfigFile,
				Aliases: []string{"c"},
				Value:   "application",
				EnvVars: []string{"APP_CONF_FILE"},
				Usage:   "Load application configuration from `FILE`",
			},
			&cli.StringFlag{
				Name:    argNameLogFile,
				Aliases: []string{"log"},
				Value:   "./conf.d/log.yml",
				EnvVars: []string{"LOG_CONF_FILE"},
				Usage:   "Load logger configuration from `FILE`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "show version",
				Action: func(c *cli.Context) error {
					fmt.Println(build.Version)
					return nil
				},
			},
			{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "show build info",
				Action: func(c *cli.Context) error {
					fmt.Println(build)
					return nil
				},
			},
		},
		Action: action,
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	return app
}

package cmd

import (
	"fmt"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/urfave/cli/v2"
	"sort"
	"strconv"
	"time"
)

type App struct {
	Name        string
	Copyright   string
	Description string
}

var _app = App{
	Name:        "flux.go gateway",
	Copyright:   "(c) " + strconv.Itoa(time.Now().Year()) + " bytepowered.net",
	Description: "Flux.go is a lightweight gateway for dubbo/http/grpc.",
}

func NewApp(action cli.ActionFunc, build flux.Build) *cli.App {
	return NewAppOf(_app, action, build)
}

func NewAppOf(app App, action cli.ActionFunc, build flux.Build) *cli.App {
	inst := &cli.App{
		Name:        app.Name,
		Version:     build.Version,
		Copyright:   app.Copyright,
		Description: app.Description,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    argNameAppConfigFile,
				Aliases: []string{"c"},
				Value:   "application",
				EnvVars: []string{"APP_CONF_NAME"},
				Usage:   "application config file name, without file ext",
			},
			&cli.StringFlag{
				Name:    argNameLogFile,
				Aliases: []string{"l"},
				Value:   "./conf.d/log.yml",
				EnvVars: []string{"APP_LOG_CONF_FILE"},
				Usage:   "Load logger configuration from `FILE`",
			},
		},
		Commands: []*cli.Command{
			showBuildInfo(build),
			showHelpInfo(build),
		},
		Action: action,
	}
	sort.Sort(cli.FlagsByName(inst.Flags))
	sort.Sort(cli.CommandsByName(inst.Commands))
	return inst
}

func showHelpInfo(build flux.Build) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "show version",
		Action: func(c *cli.Context) error {
			fmt.Println(build.Version)
			return nil
		},
	}
}

func showBuildInfo(build flux.Build) *cli.Command {
	return &cli.Command{
		Name:  "info",
		Usage: "show build info",
		Action: func(c *cli.Context) error {
			fmt.Println(build)
			return nil
		},
	}
}

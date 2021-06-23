package cmd

import (
	"github.com/bytepowered/fluxgo/pkg/server"
	"github.com/urfave/cli/v2"
)

const (
	argNameLogFile    = "logfile"
	argNameConfigFile = "config"
)

func NewActions(actions ...cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		for _, action := range actions {
			if err := action(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

func InitLoggerAction(ctx *cli.Context) error {
	if err := server.InitLogger(ctx.String(argNameLogFile)); err != nil {
		return err
	}
	return nil
}

func InitConfigAction(ctx *cli.Context) error {
	if err := server.InitConfig(ctx.String(argNameConfigFile)); err != nil {
		return err
	}
	return nil
}

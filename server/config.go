package server

import (
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/viper"
	"os"
)

func InitConfig(envKey string) {
	file := "application"
	env := os.Getenv(envKey)
	if env != "" {
		file = file + "-" + env
	}
	viper.SetConfigName(file)
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/flux/conf.d")
	viper.AddConfigPath("./conf.d")
	logger.Infof("Using config, file: %s, Env: %s", file, env)
	if err := viper.ReadInConfig(); nil != err {
		logger.Panicf("Fatal config error, path: %s, err: ", file, err)
	}
}

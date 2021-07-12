#!/bin/bash

export APP_LOG_CONF_FILE=./conf.d/log.yml
export CONF_CONSUMER_FILE_PATH=./conf.d/dubbo.yml
export CONF_PROVIDER_FILE_PATH=./conf.d/dubbo.yml

go run main.go

package pkg

import "os"

const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

func IsEnv(exceptEnv string) bool {
	return exceptEnv == os.Getenv("env")
}

package config

import "os"

type Config struct {
	HSMLibPath  string
	CU_USERNAME string
	CU_PASSWORD string
	SO_PASSWORD string
}

type ENVKey string

const (
	KeyHSMLibPath ENVKey = "HSM_LIB_PATH"
	KeyCUUsername ENVKey = "CU_USERNAME"
	KeyCUPassword ENVKey = "CU_PASSWORD"
	KeySOPassword ENVKey = "SO_PASSWORD"
)

func NewConfig() Config {
	return Config{
		HSMLibPath:  os.Getenv(string(KeyHSMLibPath)),
		CU_USERNAME: os.Getenv(string(KeyCUUsername)),
		CU_PASSWORD: os.Getenv(string(KeyCUPassword)),
		SO_PASSWORD: os.Getenv(string(KeySOPassword)),
	}
}

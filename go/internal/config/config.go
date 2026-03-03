package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	HECURL                   string `toml:"hec_url"`
	HECToken                 string `toml:"hec_token"`
	HBIntervalSeconds        int    `toml:"hb_interval_seconds"`
	CPUDetailIntervalSeconds int    `toml:"cpu_detail_interval_seconds"`
	Host                     string `toml:"host"`
	Index                    string `toml:"index"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		HBIntervalSeconds:        60,
		CPUDetailIntervalSeconds: 10,
		Index:                    "heartbeat",
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

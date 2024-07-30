package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var globalConfig *Config

type Config struct {
	Logger           Logger           `yaml:"logger"`
	Socket           Socket           `yaml:"socket"`
	CloudCollectTask CloudCollectTask `yaml:"cloud_collect_task"`
}

type CloudCollectTask struct {
	WinStackTask map[string]interface{} `yaml:"win_stack"`
	VMWare       map[string]interface{} `yaml:"vm_ware"`
}

func ParseConfig(p string) (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get working directory: %v", err)
	}
	confPath := filepath.Join(dir, p)
	content, err := os.ReadFile(confPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %s, error: %v", confPath, err)
	}
	err = yaml.Unmarshal(content, &globalConfig)
	if err != nil {
		return nil, fmt.Errorf("yaml unmarshal error: %v", err)
	}
	return globalConfig, nil
}

// GetConfig 获取全局配置文件
func GetConfig() *Config {
	return globalConfig
}

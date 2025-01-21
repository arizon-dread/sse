package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var once sync.Once
var conf *Config

type Config struct {
	Cors  Cors  `yaml:"cors"`
	Cache Cache `yaml:"cache,omitempty"`
}

type Cors struct {
	Url     string   `yaml:"url"`
	Methods []string `yaml:"methods"`
	Headers []string `yaml:"headers"`
}
type Cache struct {
	Url      string `yaml:"url"`
	Password string `yaml:"password"`
}

func Get() *Config {
	once.Do(func() {
		hostEnv := os.Environ()
		var env string
		he := make(map[string]string)
		for _, v := range hostEnv {
			keyVal := strings.Split(v, "=")
			he[keyVal[0]] = keyVal[1]
		}
		for k, v := range he {
			if k == "environment" {
				env = v
			}
		}
		conf = &Config{}
		if env != "" {
			bEnv, err := os.ReadFile(fmt.Sprintf("./configs/%v.yaml", env))
			if err != nil {
				fmt.Printf("config/%v.yaml not found or readable, skipping", env)
			} else {
				err = yaml.Unmarshal(bEnv, &conf)
			}
		} else {
			b, err := os.ReadFile("./configs/config.yaml")
			if err != nil {
				log.Fatal("config/config.yaml not found or readable, quitting")
			}
			confErr := yaml.Unmarshal(b, &conf)
			if confErr != nil {
				log.Fatal("unable to parsal config into struct")
			}

		}

	})
	return conf
}

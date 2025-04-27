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
	Cors    Cors   `yaml:"cors"`
	Cache   Cache  `yaml:"cache,omitempty"`
	ApiPort string `yaml:"apiPort"`
}

type Cors struct {
	Url     string   `yaml:"url"`
	Methods []string `yaml:"methods"`
	Headers []string `yaml:"headers"`
}
type Cache struct {
	Url              string `yaml:"url"`
	Password         string `yaml:"password,omitempty"`
	UnreadCacheCount int    `yaml:"unreadCacheCount,omitempty"`
}

func Get() *Config {
	once.Do(func() {
		hostEnv := os.Environ()
		var env string
		for _, v := range hostEnv {
			keyVal := strings.Split(v, "=")
			if keyVal[0] == "environment" {
				env = keyVal[1]
			}
		}
		conf = &Config{}
		if env != "" {
			bEnv, err := os.ReadFile(fmt.Sprintf("./configs/%v.yaml", env))
			if err != nil {
				log.Printf("config/%v.yaml not found or readable, skipping", env)
			} else {
				err = yaml.Unmarshal(bEnv, &conf)
				if err != nil {
					log.Fatalf("unable to unmarshal config/%v.yaml into go struct, err: %v\n", env, err)
				}
			}
		} else {
			b, err := os.ReadFile("./configs/config.yaml")
			if err != nil {
				log.Fatal("config/config.yaml not found or readable, quitting")
			}
			confErr := yaml.Unmarshal(b, &conf)
			if confErr != nil {
				log.Fatalf("unable to unmarsal config/%v.yaml into go struct", env)
			}

		}
		if conf.ApiPort == "" {
			log.Println("No api port set, defaulting to 8080")
			conf.ApiPort = "8080"
		}
	})
	return conf
}

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

type WebAPIConfig struct {
	ConfigFile struct {
		Path string `conf:"default:conf/config.yaml"`
	}
	DB struct {
		Filename string `conf:"default:data/wasaphoto.db"`
	}
	Debug bool `conf:"default:false"`
	Web   struct {
		APIHost         string        `conf:"default:localhost:3000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
}

// loadConfig creates a WebAPIConfig starting from flags, environment variables and configuration file.
// It works by loading environment variables first, then update the config using command line flags, finally loading the
// configuration file (specified in WebAPIConfiguration.Config.Path).
// So, CLI parameters will override the environment, and configuration file will override everything.
// Note that the configuration file can be specified only via CLI or environment variable.
func loadConfig() (WebAPIConfig, error) {
	var cfg WebAPIConfig

	// Parse the config struct from the environment variables and command line arguments
	if err := conf.Parse(os.Args[1:], "WASAPHOTO", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("WASAPHOTO", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("generating config usage message: %w", err)
			}
			fmt.Println(usage)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Override values from YAML config file
	fp, err := os.Open(cfg.ConfigFile.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("can't open the config file: %w", err)
	} else if err == nil {

		defer fp.Close()

		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("can't read config file: %w", err)
		}

		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
		}
	}

	return cfg, nil
}

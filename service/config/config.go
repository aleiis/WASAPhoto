package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

var (
	once       sync.Once    // Ensure that the configuration is loaded only once
	config     WebAPIConfig // Configuration of the web API
	loadingErr error        // Error during the configuration loading
)

type WebAPIConfig struct {
	ConfigFile   string `conf:"default:config.yml" yaml:"config_file"`
	LogFile      string `conf:"default:wasaphoto/log/webapi.log" yaml:"log_file"`
	Debug        bool   `conf:"default:false" yaml:"debug"`
	ImageStorage struct {
		Path string `conf:"default:wasaphoto/images" yaml:"path"`
	} `yaml:"image_storage"`
	DB struct {
		User     string `conf:"default:root" yaml:"user"`
		Password string `conf:"default:root" yaml:"password"`
		Address  string `conf:"default:127.0.0.1:3306" yaml:"address"`
		Database string `conf:"default:wasaphoto" yaml:"database"`
	} `yaml:"db"`
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000" yaml:"api_host"`
		DebugHost       string        `conf:"default:0.0.0.0:4000" yaml:"debug_host"`
		ReadTimeout     time.Duration `conf:"default:5s" yaml:"read_timeout"`
		WriteTimeout    time.Duration `conf:"default:5s" yaml:"write_timeout"`
		ShutdownTimeout time.Duration `conf:"default:5s" yaml:"shutdown_timeout"`
	} `yaml:"web"`
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
		// Print help message
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
	fp, err := os.Open(cfg.ConfigFile)
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

func GetConfig() (WebAPIConfig, error) {
	once.Do(func() {
		// Load the configuration only once
		config, loadingErr = loadConfig()
	})
	return config, loadingErr
}

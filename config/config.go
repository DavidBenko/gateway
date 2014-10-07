// Package config implements configuration parsing for Gateway.
package config

import (
	"flag"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// Configuration specifies the complete Gateway configuration.
type Configuration struct {
	Proxy ProxyServer
}

// ProxyServer specifies configuration options that apply to the proxy.
type ProxyServer struct {
	Port  int64
	Admin ProxyAdmin
}

// ProxyAdmin specifies configuration options that apply to the admin section
// of the proxy.
type ProxyAdmin struct {
	PathPrefix string
	Host       string
}

const envPrefix = "APGATEWAY_"

var defaultConfigFile string

func init() {
	defaultConfigFile = envValueForFlag("config")
	if defaultConfigFile == "" {
		defaultConfigFile = "/etc/gateway/gateway.conf"
	}
}

// Parse all configuration.
//
// Environment variables take precendence over the configuration file,
// but command line flags take precedence over both.
func Parse(args []string) (Configuration, error) {
	config := Configuration{}

	configFile := findConfigFile(args)
	if err := parseConfigFile(&config, configFile); err != nil {
		return config, err
	}

	setupFlags(&config)
	flag.Parse()

	setUnsetFlagsFromEnv()

	return config, nil
}

// We want to parse the flags after we've read in the config file so that they
// take precedence, so we're going to extract the config file flag directly.
func findConfigFile(args []string) string {
	configRx := regexp.MustCompile("--?config=?(.*)?")
	for index, arg := range args {
		match := configRx.FindStringSubmatch(arg)
		if match == nil {
			continue
		}
		if len(match) == 2 && match[1] != "" {
			return match[1]
		}
		if len(args) > (index + 1) {
			return args[index+1]
		}
	}
	return ""
}

func parseConfigFile(config *Configuration, configFile string) error {
	if configFile == "" {
		configFile = defaultConfigFile
	}
	_, err := toml.DecodeFile(configFile, config)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func setUnsetFlagsFromEnv() {
	set := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})
	flag.VisitAll(func(f *flag.Flag) {
		if !set[f.Name] {
			if val := envValueForFlag(f.Name); val != "" {
				flag.Set(f.Name, val)
			}
		}
	})
}

func envValueForFlag(name string) string {
	key := envPrefix + strings.ToUpper(strings.Replace(name, "-", "_", -1))
	return os.Getenv(key)
}

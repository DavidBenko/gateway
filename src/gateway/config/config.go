// Package config implements configuration parsing for Gateway.
package config

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// Configuration specifies the complete Gateway configuration.
type Configuration struct {
	File string `flag:"config" default:"/etc/gateway/gateway.conf" usage:"The path to the configuration file"`

	Proxy ProxyServer
	Admin ProxyAdmin
	Raft  RaftServer
}

// ProxyServer specifies configuration options that apply to the proxy.
type ProxyServer struct {
	Host string `flag:"proxy-host" default:"localhost" usage:"The hostname of the proxy server"`
	Port int64  `flag:"proxy-port" default:"5000"      usage:"The port of the proxy server"`
}

// ProxyAdmin specifies configuration options that apply to the admin section
// of the proxy.
type ProxyAdmin struct {
	PathPrefix string `flag:"admin-path-prefix" default:"/admin/" usage:"The path prefix the administrative area is accessible under"`
	Host       string `flag:"admin-host"        default:""        usage:"The host the administrative area is accessible via"`
}

// RaftServer specifies configuration options that apply to the Raft server.
type RaftServer struct {
	DataPath string `flag:"raft-data-path" default:"/etc/gateway/data" usage:"The path to the directory where the server data should be stored"`
	Leader   string `flag:"raft-leader"    default:""                  usage:"The connection string of the Raft leader this server should join on startup"`
	Host     string `flag:"raft-host"      default:"localhost"         usage:"The hostname of the Raft server"`
	Port     int64  `flag:"raft-port"      default:"6000"              usage:"The port of the Raft server"`
}

const envPrefix = "APGATEWAY_"

// Parse all configuration.
//
// Environment variables take precendence over the configuration file,
// but command line flags take precedence over both.
func Parse(args []string) (Configuration, error) {
	config := Configuration{}

	// Parse flags
	setupFlags(reflect.ValueOf(config))
	flag.Parse()

	// Parse environment
	setUnsetFlagsFromEnv()

	// Set default in our instance
	setDefaults(reflect.ValueOf(&config).Elem())

	// Override values with config file
	if err := parseConfigFile(&config); err != nil {
		return config, err
	}
	// Override values with flags (including environment)
	setFromFlags(reflect.ValueOf(&config).Elem())

	return config, nil
}

func parseConfigFile(config *Configuration) error {
	configFile := flag.Lookup("config").Value.String()
	_, err := toml.DecodeFile(configFile, config)
	if os.IsNotExist(err) {
		log.Printf(
			"%s Config file '%s' does not exist and will not be used.\n",
			System, configFile)
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

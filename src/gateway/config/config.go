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
	Version bool   `flag:"version" default:"false"`
	File    string `flag:"config" default:"gateway.conf"`
	License string `flag:"license"`
	Server  bool   `flag:"server" default:"false"`

	Database Database
	Proxy    ProxyServer
	Admin    ProxyAdmin
	Elastic  ElasticLogging
}

// Database specifies configuration options for your database
type Database struct {
	Migrate          bool   `flag:"db-migrate"     default:"false"`
	Driver           string `flag:"db-driver"      default:"sqlite3"`
	ConnectionString string `flag:"db-conn-string" default:"gateway.db"`
	MaxConnections   int64  `flag:"db-max-connections" default:"50"`
}

// ProxyServer specifies configuration options that apply to the proxy.
type ProxyServer struct {
	Host string `flag:"proxy-host" default:"localhost"`
	Port int64  `flag:"proxy-port" default:"5000"`

	RequestIDHeader string `flag:"proxy-request-id-header" default:""`
	EnableOSEnv     bool   `flag:"proxy-enable-os-env" default:"false"`

	CacheAPIs bool `flag:"proxy-cache-apis" default:"false"`

	HTTPTimeout   int64 `flag:"proxy-http-timeout" default:"60"`
	CodeTimeout   int64 `flag:"proxy-code-timeout" default:"5"`
	NumErrorLines int64 `flag:"proxy-code-error-lines" default:"2"`
}

// ProxyAdmin specifies configuration options that apply to the admin section
// of the proxy.
type ProxyAdmin struct {
	DevMode bool

	PathPrefix string `flag:"admin-path-prefix" default:"/admin/"`
	Host       string `flag:"admin-host"        default:""`

	SessionName    string `flag:"admin-session-name" default:"__ap_gateway"`
	AuthKey        string `flag:"admin-session-auth-key" default:""`
	EncryptionKey  string `flag:"admin-session-encryption-key" default:""`
	AuthKey2       string `flag:"admin-session-auth-key-rotate" default:""`
	EncryptionKey2 string `flag:"admin-session-encryption-key-rotate" default:""`

	RequestIDHeader string `flag:"admin-request-id-header" default:"X-Gateway-Admin-Request"`

	CORSEnabled bool   `flag:"admin-cors-enabled" default:"true"`
	CORSOrigin  string `flag:"admin-cors-origin" default:"*"`

	Username string `flag:"admin-username" default:"admin"`
	Password string `flag:"admin-password" default:""`
	Realm    string `flag:"admin-realm"    default:""`

	ShowVersion bool `flag:"admin-show-version" default:"true"`

	AddDefaultEnvironment  bool   `flag:"admin-add-default-env" default:"true"`
	DefaultEnvironmentName string `flag:"admin-default-env-name" default:"Development"`
	AddLocalhost           bool   `flag:"admin-add-localhost" default:"true"`
}

type ElasticLogging struct {
	Domain   string `flag:"elastic-logging-domain" default:""`
	Username string `flag:"elastic-logging-username" default:""`
	Password string `flag:"elastic-logging-password" default:""`
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

	// Set final convenience flags
	config.Admin.DevMode = config.DevMode()

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

func (c Configuration) DevMode() bool {
	return !c.Server
}

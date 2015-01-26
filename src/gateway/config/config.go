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

	License string `flag:"license" default:"" usage:"The path to a valid Gateway license key"`

	Database Database
	Proxy    ProxyServer
	Admin    ProxyAdmin
}

// Database specifies configuration options for your database
type Database struct {
	Migrate          bool   `flag:"db-migrate"     default:"false"                   usage:"Whether or not to migrate the database on startup"`
	Driver           string `flag:"db-driver"      default:"sqlite3"                 usage:"The database driver; sqlite or postgres"`
	ConnectionString string `flag:"db-conn-string" default:"/etc/gateway/gateway.db" usage:"The connection string for your database"`
}

// ProxyServer specifies configuration options that apply to the proxy.
type ProxyServer struct {
	Host string `flag:"proxy-host" default:"localhost" usage:"The hostname of the proxy server"`
	Port int64  `flag:"proxy-port" default:"5000"      usage:"The port of the proxy server"`

	AuthKey       string `flag:"proxy-session-auth-key" default:"" usage:"The auth key to use for cookie sessions. 64 chars recommended. If unset, they're disabled."`
	EncryptionKey string `flag:"proxy-session-encryption-key" default:"" usage:"The encryption key to use for cookie sessions. 32 chars recommended. If unset, encryption is disabled."`
}

// ProxyAdmin specifies configuration options that apply to the admin section
// of the proxy.
type ProxyAdmin struct {
	PathPrefix string `flag:"admin-path-prefix" default:"/admin/" usage:"The path prefix the administrative area is accessible under"`
	Host       string `flag:"admin-host"        default:""        usage:"The host the administrative area is accessible via"`

	SessionName    string `flag:"admin-session-name" default:"__ap_gateway" usage:"The name of the cookie to use for sessions."`
	AuthKey        string `flag:"admin-session-auth-key" default:"" usage:"The auth key to use for sessions. 64 chars recommended. Required."`
	EncryptionKey  string `flag:"admin-session-encryption-key" default:"" usage:"The encryption key to use for sessions. 32 chars recommended. If unset, encryption is disabled."`
	AuthKey2       string `flag:"admin-session-auth-key-rotate" default:"" usage:"Same as admin-session-auth-key, to be used during key rotation."`
	EncryptionKey2 string `flag:"admin-session-encryption-key-rotate" default:"" usage:"Same as admin-session-encryption-key, to be used during key rotation."`

	CORSEnabled bool   `flag:"admin-cors-enabled" default:"true" usage:"Set to false to disable CORS headers from being added to admin responses."`
	CORSOrigin  string `flag:"admin-cors-origin" default:"*" usage:"The Access-Control-Allow-Origin header value to send with admin responses."`

	Username string `flag:"admin-username" default:"admin" usage:"The username to require with HTTP Basic Auth to protect the site admin functionality"`
	Password string `flag:"admin-password" default:""      usage:"The password to require with HTTP Basic Auth to protect the site admin functionality"`
	Realm    string `flag:"admin-realm"    default:""      usage:"The HTTP Basic realm to use. Optional."`
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

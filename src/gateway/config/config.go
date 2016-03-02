// Package config implements configuration parsing for Gateway.
package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gateway/logreport"

	"github.com/BurntSushi/toml"
)

var defaultDomain = "lvh.me"

// Configuration specifies the complete Gateway configuration.
type Configuration struct {
	Version        bool   `flag:"version" default:"false"`
	File           string `flag:"config" default:"gateway.conf"`
	License        string `flag:"license"`
	LicenseContent string `flag:"license-content"`
	Server         bool   `flag:"server" default:"false"`
	Jobs           bool   `flag:"jobs" default:"true"`

	Airbrake       Airbrake
	Database       Database
	Proxy          ProxyServer
	Admin          ProxyAdmin
	Elastic        ElasticLogging
	Bleve          BleveLogging
	Soap           Soap
	RemoteEndpoint RemoteEndpoint
	SMTP           SMTP
}

// Airbrake specifies configuration for error reporting with Airbrake
type Airbrake struct {
	APIKey      string `flag:"airbrake-api-key" default:""`
	ProjectID   int64  `flag:"airbrake-project-id" default:"0"`
	Environment string `flag:"airbrake-environment" default:""`
}

// Database specifies configuration options for your database
type Database struct {
	Migrate          bool   `flag:"db-migrate"     default:"false"`
	Driver           string `flag:"db-driver"      default:"sqlite3"`
	ConnectionString string `flag:"db-conn-string" default:"gateway.db"`
	MaxConnections   int64  `flag:"db-max-connections" default:"50"`
}

// Soap specifies configuration options pertaining to remote SOAP endpoints
type Soap struct {
	JdkPath        string `flag:"soap-jdk-path"    default:""`
	SoapClientHost string `flag:"soap-client-host" default:"localhost"`
	SoapClientPort int64  `flag:"soap-client-port" default:"19083"`
	ThreadPoolSize int64  `flag:"soap-thread-pool-size" default:"0"`
	JavaOpts       string `flag:"soap-java-opts" default:""`
}

// ProxyServer specifies configuration options that apply to the proxy.
type ProxyServer struct {
	Domain string `flag:"proxy-domain" default:"lvh.me"`
	Host   string `flag:"proxy-host" default:"localhost"`
	Port   int64  `flag:"proxy-port" default:"5000"`

	RequestIDHeader string `flag:"proxy-request-id-header" default:""`
	EnableOSEnv     bool   `flag:"proxy-enable-os-env" default:"false"`

	CacheAPIs bool `flag:"proxy-cache-apis" default:"false"`

	HTTPTimeout   int64 `flag:"proxy-http-timeout" default:"60"`
	CodeTimeout   int64 `flag:"proxy-code-timeout" default:"5"`
	NumErrorLines int64 `flag:"proxy-code-error-lines" default:"2"`
}

// RemoteEndpoint specifies which types of remote endpionts are available
type RemoteEndpoint struct {
	HTTPEnabled       bool `flag:"remote-endpoint-http-enabled" default:"true"`
	SQLServerEnabled  bool `flag:"remote-endpoint-sqlserver-enabled" default:"true"`
	MySQLEnabled      bool `flag:"remote-endpoint-mysql-enabled" default:"true"`
	PostgreSQLEnabled bool `flag:"remote-endpoint-postgresql-enabled" default:"true"`
	MongoDBEnabled    bool `flag:"remote-endpoint-mongodb-enabled" default:"true"`
	ScriptEnabled     bool `flag:"remote-endpoint-script-enabled" default:"true"`
	SoapEnabled       bool `flag:"remote-endpoint-soap-enabled" default:"true"`
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
	CookieDomain   string `flag:"admin-session-cookie-domain" default:""`

	RequestIDHeader string `flag:"admin-request-id-header" default:"X-Gateway-Admin-Request"`

	CORSEnabled bool   `flag:"admin-cors-enabled" default:"true"`
	CORSOrigin  string `flag:"admin-cors-origin" default:"*"`

	Username string `flag:"admin-username" default:"admin"`
	Password string `flag:"admin-password" default:""`
	Realm    string `flag:"admin-realm"    default:""`

	ShowVersion bool `flag:"admin-show-version" default:"true"`

	AddDefaultEnvironment  bool   `flag:"admin-add-default-env" default:"true"`
	DefaultEnvironmentName string `flag:"admin-default-env-name" default:"Development"`

	CreateDefaultHost bool `flag:"admin-create-default-host" default:"true"`

	EnableBroker    bool   `flag:"enable-broker" default:"true"`
	Broker          string `flag:"broker" default:"localhost"`
	BrokerPubPort   string `flag:"broker-pub-port" default:"5555"`
	BrokerSubPort   string `flag:"broker-sub-port" default:"5556"`
	BrokerTransport string `flag:"broker-transport" default:"tcp"`
	BrokerWs        string `flag:"broker-ws" default:"localhost:5000"`

	EnableRegistration bool `flag:"admin-enable-registration" default:"true"`

	DefaultAPIAccessScheme string `flag:"admin-default-api-access-scheme" default:"http://{{hosts.[0]}}:5000"`

	GoogleAnalyticsTrackingId string `flag:"admin-google-analytics-tracking-id" default:""`
}

type ElasticLogging struct {
	Url string `flag:"elastic-logging-url" default:""`
}

type BleveLogging struct {
	File        string `flag:"bleve-logging-file" default:"logs.bleve"`
	DeleteAfter int64  `flag:"bleve-logging-delete-after" default:"30"`
}

type SMTP struct {
	Server      string `flag:"smtp-server"`
	Port        int64  `flag:"smtp-port" default:"25"`
	User        string `flag:"smtp-user"`
	Password    string `flag:"smtp-password"`
	Sender      string `flag:"smtp-sender"`
	EmailScheme string `flag:"smtp-email-scheme" default:"http"`
	EmailHost   string `flag:"smtp-email-host"`
	EmailPort   int64  `flag:"smtp-email-port" default:"0"`
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

	// Verify that the configuration is valid before proceeding
	if err := verify(config); err != nil {
		return config, err
	}

	return config, nil
}

func Commands() []string {
	return flag.Args()
}

// verify the configuration
func verify(config Configuration) error {
	if config.DevMode() {
		return nil
	}
	// Verify that a domain is set (other than the default)
	if config.Proxy.Domain == defaultDomain {
		return fmt.Errorf("proxy-domain not provided.  proxy-domain must be set when running in server mode")
	}

	return nil
}

func parseConfigFile(config *Configuration) error {
	configFile := flag.Lookup("config").Value.String()
	_, err := toml.DecodeFile(configFile, config)
	if os.IsNotExist(err) {
		logreport.Printf(
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

func (config *ProxyAdmin) XPub() string {
	return config.BrokerTransport + "://" +
		config.Broker + ":" +
		config.BrokerPubPort
}

func (config *ProxyAdmin) XSub() string {
	return config.BrokerTransport + "://" +
		config.Broker + ":" +
		config.BrokerSubPort
}

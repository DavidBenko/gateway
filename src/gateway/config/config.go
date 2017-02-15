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
	Version       bool   `flag:"version" default:"false"`
	OssLicenses   string `flag:"oss-licenses" default:"false"`
	ExampleConfig string `flag:"example-config" default:"false"`
	File          string `flag:"config" default:"gateway.conf"`
	Server        bool   `flag:"server" default:"false"`
	Jobs          bool   `flag:"jobs" default:"true"`

	Airbrake        Airbrake
	Database        Database
	Proxy           ProxyServer
	Job             BackgroundJob
	Admin           ProxyAdmin
	Elastic         ElasticLogging
	Bleve           BleveLogging
	PostgresLogging PostgresLogging
	Soap            Soap
	Store           Store
	RemoteEndpoint  RemoteEndpoint
	SMTP            SMTP
	Push            Push
	Docker          Docker
	Stats           Stats
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
	ClientJar      string `flag:"soap-client-jar"        default:"gateway-soap-client.jar"`
}

// Store specifies configuration options for store remote endpoints
type Store struct {
	Migrate          bool   `flag:"store-migrate"     default:"false"`
	Type             string `flag:"store-type"        default:"boltdb"`
	ConnectionString string `flag:"store-conn-string" default:"store.db"`
	MaxConnections   int64  `flag:"store-max-connections" default:"50"`
}

// Docker specifies configuration options for docker remote endpoints
type Docker struct {
	Memory           int64  `flag:"docker-memory" default:"128"`
	CPUShares        int64  `flag:"docker-cpu-shares" default:"0"`
	Host             string `flag:"docker-host" default:""`
	Tls              bool   `flag:"docker-tls" default:"false"`
	TlsCertFile      string `flag:"docker-tls-cert" default:""`
	TlsCaCertFile    string `flag:"docker-tls-cacert" default:""`
	TlsKeyFile       string `flag:"docker-tls-key" default:""`
	TlsCertContent   string `flag:"docker-tls-cert-content" default:""`
	TlsCaCertContent string `flag:"docker-tls-cacert-content" default:""`
	TlsKeyContent    string `flag:"docker-tls-key-content" default:""`
}

// Stats database configuration.
type Stats struct {
	Collect          bool   `flag:"stats-collect"     default:"true"`
	Migrate          bool   `flag:"stats-migrate"     default:"false"`
	Driver           string `flag:"stats-driver"      default:"sqlite3"`
	ConnectionString string `flag:"stats-conn-string" default:"gateway-stats.db"`
	MaxConnections   int64  `flag:"stats-max-connections" default:"50"`
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

	HealthCheckPath string `flag:"proxy-health-check-path" default:"/__gw-health-check"`

	KeyCacheSize            int64 `flag:"proxy-key-cache-size" default:"0"`
	RemoteEndpointCacheSize int64 `flag:"proxy-remote-endpoint-cache-size" default:"0"`
}

// BackgroundJob specifies configuration options that apply to jobs.
type BackgroundJob struct {
	Enable      bool `flag:"job-enable" default:"true"`
	EnableOSEnv bool `flag:"job-enable-os-env" default:"false"`

	CodeTimeout   int64 `flag:"job-code-timeout" default:"5"`
	NumErrorLines int64 `flag:"job-code-error-lines" default:"2"`
}

// RemoteEndpoint specifies which types of remote endpionts are available
type RemoteEndpoint struct {
	HTTPEnabled           bool `flag:"remote-endpoint-http-enabled" default:"true"`
	SQLServerEnabled      bool `flag:"remote-endpoint-sqlserver-enabled" default:"true"`
	MySQLEnabled          bool `flag:"remote-endpoint-mysql-enabled" default:"true"`
	PostgreSQLEnabled     bool `flag:"remote-endpoint-postgresql-enabled" default:"true"`
	MongoDBEnabled        bool `flag:"remote-endpoint-mongodb-enabled" default:"true"`
	StoreEnabled          bool `flag:"remote-endpoint-store-enabled" default:"true"`
	LDAPEnabled           bool `flag:"remote-endpoint-ldap-enabled" default:"true"`
	HanaEnabled           bool `flag:"remote-endpoint-hana-enabled" default:"true"`
	PushEnabled           bool `flag:"remote-endpoint-push-enabled" default:"true"`
	RedisEnabled          bool `flag:"remote-endpoint-redis-enabled" default:"true"`
	SMTPEnabled           bool `flag:"remote-endpoint-smtp-enabled" default:"true"`
	JobEnabled            bool `flag:"remote-endpoint-job-enabled" default:"true"`
	ScriptEnabled         bool `flag:"remote-endpoint-script-enabled" default:"false"`
	SoapEnabled           bool `flag:"remote-endpoint-soap-enabled" default:"false"`
	DockerEnabled         bool `flag:"remote-endpoint-docker-enabled" default:"false"`
	KeyEnabled            bool `flag:"remote-endpoint-key-enabled" default:"true"`
	CustomFunctionEnabled bool `flag:"remote-endpoint-custom-function-enabled" default:"false"`
	ScrubData             bool `flag:"remote-endpoint-scrub-data"     default:"false"`
}

// ProxyAdmin specifies configuration options that apply to the admin section
// of the proxy.
type ProxyAdmin struct {
	DevMode bool

	PathPrefix   string `flag:"admin-path-prefix" default:"/admin/"`
	UiPathPrefix string `flag:"admin-ui-path-prefix" default:"/admin/"`
	Host         string `flag:"admin-host"        default:""`

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
	Password string `flag:"admin-password" default:"admin"`
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

	// Stripe related configuration
	StripeSecretKey       string `flag:"stripe-secret-key" default:""`
	StripePublishableKey  string `flag:"stripe-publishable-key" default:""`
	StripeFallbackPlan    string `flag:"stripe-fallback-plan" default:""`
	StripeMigrateAccounts bool   `flag:"stripe-migrate-accounts"     default:"false"`

	APIHost             string `flag:"admin-api-host"        default:""`
	WsHeartbeatInterval int64  `flag:"ws-heartbeat-interval" default:"60"`
	WsWriteDeadline     int64  `flag:"ws-write-deadline" default:"10"`
	WsReadDeadline      int64  `flag:"ws-read-deadline" default:"10"`

	ReplMaximumFrameSize int64 `flag:"repl-maximum-frame-size" default:"1024"`
}

type ElasticLogging struct {
	Url         string `flag:"elastic-logging-url" default:""`
	DeleteAfter int64  `flag:"elastic-logging-delete-after" default:"30"`
}

type BleveLogging struct {
	File        string `flag:"bleve-logging-file" default:"logs.bleve"`
	DeleteAfter int64  `flag:"bleve-logging-delete-after" default:"30"`
}

type PostgresLogging struct {
	Enable           bool   `flag:"postgres-logging-enable" default:"false"`
	Migrate          bool   `flag:"postgres-logging-migrate" default:"false"`
	ConnectionString string `flag:"postgres-logging-conn-string" default:"dbname=gateway_logs sslmode=disable"`
	MaxConnections   int64  `flag:"postgres-logging-max-connections" default:"50"`
	DeleteAfter      int64  `flag:"postgres-logging-delete-after" default:"30"`
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

// Push specifies the configuration for the push subsystem
type Push struct {
	EnableBroker    bool   `flag:"enable-push-broker" default:"true"`
	Broker          string `flag:"push-broker" default:"localhost"`
	BrokerPubPort   string `flag:"push-broker-pub-port" default:"5557"`
	BrokerSubPort   string `flag:"push-broker-sub-port" default:"5558"`
	BrokerTransport string `flag:"push-broker-transport" default:"tcp"`
	ConnectTimeout  int64  `flag:"push-connect-timeout" default:"2"`
	MQTTURI         string `flag:"push-mqtt-uri" default:"tcp://:1883"`
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

	// Enable store remote endpoints if store is configured
	config.RemoteEndpoint.StoreEnabled = config.Store.Type != ""

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

func (p *ProxyServer) GetEnableOSEnv() bool {
	return p.EnableOSEnv
}

func (p *ProxyServer) GetCodeTimeout() int64 {
	return p.CodeTimeout
}

func (p *ProxyServer) GetNumErrorLines() int64 {
	return p.NumErrorLines
}

func (j *BackgroundJob) GetEnableOSEnv() bool {
	return j.EnableOSEnv
}

func (j *BackgroundJob) GetCodeTimeout() int64 {
	return j.CodeTimeout
}

func (j *BackgroundJob) GetNumErrorLines() int64 {
	return j.NumErrorLines
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

func (config *Push) XPub() string {
	return config.BrokerTransport + "://" +
		config.Broker + ":" +
		config.BrokerPubPort
}

func (config *Push) XSub() string {
	return config.BrokerTransport + "://" +
		config.Broker + ":" +
		config.BrokerSubPort
}

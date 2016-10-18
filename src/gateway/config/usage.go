package config

const dbConnStringHelp = `The connection string for your database

    	Connection string is of the format "parameter=value parameter2=value2 parameter3=value3"

    	Valid parameters:

    	* dbname - The name of the database to connect to
    	* user - The user to sign in as
    	* password - The user's password
    	* host - The host to connect to. Values that start with / are for unix domain sockets. (default is localhost)
    	* port - The port to bind to. (default is 5432)
    	* sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
    	* fallback_application_name - An application_name to fall back to if one isn't provided.
    	* connect_timeout - Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
    	* sslcert - Cert file location. The file must contain PEM encoded data.
    	* sslkey - Key file location. The file must contain PEM encoded data.
    	* sslrootcert - The location of the root certificate file. The file must contain PEM encoded data.

    	Valid sslmode values:

    	* disable - No SSL
    	* require - Always SSL (skip verification)
    	* verify-ca - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
    	* verify-full - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)

    	Use single quotes for values that contain whitespace:

    	"user=pgtest password='with spaces'"

    	A backslash will escape the next character in values:

    	"user=space\ man password='it\'s valid'"`

var usageStrings = map[string]string{
	"version":         "Whether to print the version and quit",
	"oss-licenses":    "Whether to print the list of libraries and respective licenses used by this product and quit",
	"example-config":  "Whether to print an example gateway.conf file and quit",
	"config":          "The path to the configuration file",
	"license":         "The path to a valid Gateway license key",
	"license-content": "The license contents as a base64 encoded string.  If present, license option is ignored",
	"server":          "Whether or not to run in server mode",
	"jobs":            "Run background jobs",

	"airbrake-api-key":     "The API key to use for Airbrake notifications",
	"airbrake-project-id":  "The ID assigned to your Airbrake project",
	"airbrake-environment": "The environment tag under which errors are reported to Airbrake",

	"db-migrate":         "Whether or not to migrate the database on startup",
	"db-driver":          "The database driver; sqlite or postgres",
	"db-conn-string":     dbConnStringHelp,
	"db-max-connections": "The maximum number of connections to use",

	"soap-jdk-path":         "The home directory of your JDK 1.8 installation",
	"soap-client-host":      "The hostname for the soap client.  Defaults to localhost.",
	"soap-client-port":      "The port number to listen on for the soap client.  Defaults to 19083",
	"soap-thread-pool-size": "The number of worker threads in the JVM that will concurrently process soap requests.  When set to 0, pooling is disabled (i.e. a new thread per request).  Defaults to 0.",
	"soap-java-opts":        "The JVM options to pass to the JVM on startup that will be used to invoke SOAP services",

	"store-migrate":         "Whether or not to migrate the store database on startup",
	"store-type":            "The type of database to use for store remote endpoints",
	"store-conn-string":     "The database connection string for the store. See: db-conn-string",
	"store-max-connections": "The maximum number of connections to use",

	"proxy-domain": "The domain name for the proxy server. Required when running in server mode, else defaults to lvh.me",
	"proxy-host":   "The hostname of the proxy server",
	"proxy-port":   "The port of the proxy server",

	"proxy-request-id-header": "The header to send the proxy request ID back in. Not sent if blank.",
	"proxy-enable-os-env":     "Whether or not to expose the OS's ENV to proxy code.",

	"proxy-cache-apis": "Whether or not to cache API data when serving proxy calls",

	"proxy-http-timeout":     "The timeout in seconds to use for proxied HTTP requests.",
	"proxy-code-timeout":     "The timeout in seconds to use for proxy script code.",
	"proxy-code-error-lines": "The number of lines of code to show around script errors in dev mode.",

	"proxy-health-check-path": "The endpoint that responds with 200 - ok to health check requests when healthy. This applies system-wide and cannot be used by APIs as an endpoint. Defaults to /__gw-health-check. Blank will disable the feature.",

	"remote-endpoint-http-enabled":       "Whether or not http remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-sqlserver-enabled":  "Whether or not MS SQLServer remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-mysql-enabled":      "Whether or not MySQL remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-postgresql-enabled": "Whether or not PostgreSQL remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-mongodb-enabled":    "Whether or not MongoDB remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-store-enabled":      "Whether or not Store remote endpoints are enabled. Defaults to false.",
	"remote-endpoint-ldap-enabled":       "Whether or not LDAP remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-hana-enabled":       "Whether or not SAP Hana remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-push-enabled":       "Whether or not Push remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-redis-enabled":      "Whether or not Redis remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-smtp-enabled":       "Whether or not SMTP remote endpoints are enabled. Defaults to true.",
	"remote-endpoint-script-enabled":     "Whether or not script remote endpoints are enabled. Defaults to false.",
	"remote-endpoint-soap-enabled":       "Whether or not soap remote endpoints are enabled. Defaults to false.",
	"remote-endpoint-docker-enabled":     "Whether or not Docker remote endpoints are enabled. Defaults to false.",

	"admin-path-prefix": "The path prefix the administrative area is accessible under",
	"admin-host":        "The host the administrative area is accessible via",

	"admin-session-name":                  "The name of the cookie to use for sessions.",
	"admin-session-auth-key":              "The auth key to use for sessions. 64 chars recommended. Required.",
	"admin-session-encryption-key":        "The encryption key to use for sessions. 32 chars recommended. If unset, encryption is disabled.",
	"admin-session-auth-key-rotate":       "Same as admin-session-auth-key, to be used during key rotation.",
	"admin-session-encryption-key-rotate": "Same as admin-session-encryption-key, to be used during key rotation.",
	"admin-session-cookie-domain":         "The domain to set on the session cookie.",

	"admin-request-id-header": "The header to send the admin request ID back in. Not sent if blank.",

	"admin-cors-enabled": "Set to false to disable CORS headers from being added to admin responses.",
	"admin-cors-origin":  "The Access-Control-Allow-Origin header value to send with admin responses.",

	"admin-username": "The username to require with HTTP Basic Auth to protect the site admin functionality",
	"admin-password": "The password to require with HTTP Basic Auth to protect the site admin functionality",
	"admin-realm":    "The HTTP Basic realm to use. Optional.",

	"admin-show-version": "Whether or not to expose the Gateway version",

	"admin-add-default-env":  "Whether or not to add a default environment to new APIs in dev mode",
	"admin-default-env-name": "The name of the default environment to create",

	"admin-enable-registration": "Enable account registration API",

	"admin-create-default-host":       "Whether or not to create a deafult host when an API is created, based off of the configured proxy-domain",
	"admin-default-api-access-scheme": "The Base URL to use for accessing the API that will be displayed to the user in the front-end admin app after an API is created.  This value can be interpolated based on the hosts configured for an API.  Default is 'http://{{hosts[0]}}:5000' where '{{hosts[0]}}' the first value in the array of interpolated hosts for that API.",

	"admin-google-analytics-tracking-id": "A Google Analytics Tracking ID to be used by the rendered Admin UI.",

	"stripe-secret-key":       "Stripe API Secret Key",
	"stripe-publishable-key":  "Stripe API Publishable Key",
	"stripe-fallback-plan":    "Stripe plan to fallback on when subscription billing fails (this uses the internal Gateway plan name)",
	"stripe-migrate-accounts": "Stripe Migrate Accounts is whether or not to create Stripe customers for existing accounts without a Stripe Customer ID.",

	"admin-api-host": "The host the administrative api is accessible via. If left blank the Admin Host is used. Include the protocol (http or https) when using this setting.",

	"elastic-logging-url":          "The url of the elastic server",
	"elastic-logging-delete-after": "How long in days to keep logs",

	"bleve-logging-file":         "The bleve file to store logs in",
	"bleve-logging-delete-after": "How long in days to keep logs",

	"enable-broker":    "Run as messaging broker",
	"broker":           "The address or name of the broker",
	"broker-pub-port":  "The port of the broker pub",
	"broker-sub-port":  "The port of the broker sub",
	"broker-transport": "The broker transport",
	"broker-ws":        "The broker websocket location",

	"smtp-server":       "The address or name of the smtp server",
	"smtp-port":         "The port of the smtp server",
	"smtp-user":         "The user name for the smtp server",
	"smtp-password":     "The password for the smtp server",
	"smtp-sender":       "The sender of emails from gateway",
	"smtp-email-scheme": "The scheme to be used in email links",
	"smtp-email-host":   "The host to be used in email links",
	"smtp-email-port":   "The port to be used in email links",

	"enable-push-broker":    "Run as messaging push broker",
	"push-broker":           "The address or name of the push broker",
	"push-broker-pub-port":  "The port of the push broker pub",
	"push-broker-sub-port":  "The port of the push broker sub",
	"push-broker-transport": "The push broker transport",
	"push-connect-timeout":  "The connect timeout for MQTT",
	"push-mqtt-uri":         "The URI for the MQTT server",

	"docker-memory":             "The amount of memory MB to allocate to each running docker container. Default is 128MB.",
	"docker-cpu-shares":         "The CPU share weight to allocate to each running docker container. Default is 1.",
	"docker-host":               "The host to connect to when performing Docker client operations. Default is blank which then uses the env to configure the Docker client.",
	"docker-tls":                "Whether to use TLS when connecting to the provided Docker Host. Default is false.",
	"docker-tls-cert":           "The cert PEM file to use when conneting to the Docker Host using TLS. Default is blank.",
	"docker-tls-cacert":         "The CA cert PEM file to use when conneting to the Docker Host using TLS. Default is blank.",
	"docker-tls-key":            "The key PEM file to use when conneting to the Docker Host using TLS. Default is blank.",
	"docker-tls-cert-content":   "The base64 encoded cert PEM file contents to use when conneting to the Docker Host using TLS. Default is blank.",
	"docker-tls-cacert-content": "The base64 encoded CA cert PEM file contents to use when conneting to the Docker Host using TLS. Default is blank.",
	"docker-tls-key-content":    "The base64 encoded key PEM file contents to use when conneting to the Docker Host using TLS. Default is blank.",

	"stats-collect":         "Whether or not to collect stats on Proxy Endpoint usage.",
	"stats-migrate":         "Whether or not to migrate the stats database on startup.",
	"stats-driver":          "The stats database driver; sqlite or postgres.",
	"stats-conn-string":     dbConnStringHelp,
	"stats-max-connections": "The maximum number of connections to use.",
}

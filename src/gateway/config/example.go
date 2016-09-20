package config

// ExampleConfigurationFile outputs gateway.conf file contents
const ExampleConfigurationFile = `##################################################
### Example Nanoscale.io server configuration file
### Copyright (c) 2016 Nanoscale.io.
### Documentation: http://devhub.nanoscale.io
### Support: support@nanoscale.io
##################################################

## ---------- General ------------
# The path to a valid Gateway license key, default is './license'.
# license = ''

# The license contents as a base64 encoded string.  If present, license option is ignored
# licenseContent = ''

# Whether or not to run in server mode, default is false.
# server =

# Whether or not to run background jobs, default is true.
# jobs =

## ---------- Database ------------
[database]
# The connection string for your database, default is './gateway.db' for sqlite driver.
# Example: postgres connectionString = 'dbname=my_db user=user sslmode=disable host=my_host password=my_password'
# Postgres connectionString options:
#   dbname - The name of the database to connect to
#   user - The user to sign in as
#   password - The user's password
#   host - The host to connect to. Values that start with / are for unix domain sockets. (default is localhost)
#   port - The port to bind to. (default is 5432)
#   sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
#   fallback_application_name - An application_name to fall back to if one isn't provided.
#   connect_timeout - Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
#   sslcert - Cert file location. The file must contain PEM encoded data.
#   sslkey - Key file location. The file must contain PEM encoded data.
#   sslrootcert - The location of the root certificate file. The file must contain PEM encoded data.
#
#   Valid sslmode values:
#       * disable - No SSL
#       * require - Always SSL (skip verification)
#       * verify-ca - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
#       * verify-full - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)
#
#     Use single quotes for values that contain whitespace:
#       "user=pgtest password='with spaces'"
#
#     A backslash will escape the next character in values:
#       "user=space\ man password='it\'s valid'" (default "gateway.db")
#
# connectionString = 'dbname=my_db user=user sslmode=disable host=my_host password=my_password'

# The database driver; sqlite or postgres, default is 'sqlite3'. NOTE: Use sqllite for local development only.
# driver = 'postgres'

# The maximum number of connections to use, default is 50.
# maxConnections =

# Whether or not to migrate the database on startup, default is false.
# migrate =

## --------------------------------------------------------------------------


## ---------- Admin ------------
[admin]
# Whether or not to add a default environment to new APIs, default is true.
# addDefaultEnv =

# The auth key to use for sessions. 64 chars recommended. Required in server mode.
# authKey = 'CHANGE'

# Same as sessionAuthKey, to be used during key rotation.
# authKey2 = ''

# The Access-Control-Allow-Origin header value to send with admin responses, default is '*'
# corsOrigin = ''

# Set to false to disable CORS headers from being added to admin responses, default is true.
# corsEnabled =

# Whether or not to create a deafult host when an API is created, based off of the configured proxy-domain, default is true
# createDefaultHost =

# The name of the default environment to create, default is 'Development'
# defaultEnvName = ''

# Enable account enableRegistration API, default is true.
# enableRegistration =

# The encryption key to use for sessions. 32 chars recommended. If unset, encryption is disabled.
# encryptionKey = ''

# Same as sessionEncryptionKey, to be used during key rotation.
# encryptionKey2 = ''

# The host the administrative area is accessible via.
# host = ''

# The password for HTTP Basic Auth to protect the site admin functionality. Default is blank.
# password = 'CHANGE'

# The path prefix the administrative area is accessible under, default is '/admin/'
# pathPrefix = ''

# The HTTP Basic realm to use (optional).
# realm = ''

# The header to send the admin request ID back in. Not sent if blank, default is 'X-Gateway-Admin-Request'
# requestIdHeader = ''

# The domain to set on the admin session cookie, default is blank.
# cookieDomain =

# The name of the cookie to use for sessions, default is '__ap_gateway'.
# sessionName = ''

# Whether or not to expose the Gateway version to the Admin UI, default is true.
# showVersion =

# The username HTTP Basic Auth to protect the site admin functionality, default is 'admin'.
# username = ''

# Set to false to disable CORS headers from being added to admin responses.
# corsEnabled = true

# The Access-Control-Allow-Origin header value to send with admin responses.
# corsOrigin = '*'

# Run as messaging broker, default is true.
# enableBroker = true

# The address or name of the broker, default is localhost.
# broker = 'localhost'

# The port of the broker pub, default is 5555.
# brokerPubPort = '5555'

# The port of the broker sub, default is 5556.
# brokerSubPort = '5556'

# The broker transport, default is 'tcp'.
# brokerTransport = 'tcp'

# The broker websocket location, default is localhost:5000.
# brokerWs = 'localhost:5000'

# Enable account registration API, default is true. Only applies in server mode.
# enableRegistration = true

# The Base URL to use for accessing the API that will be displayed to the user in the front-end admin app after an API is created.  This value can be interpolated based on the hosts configured for an API.  Default is 'http://{{hosts[0]}}:5000' where '{{hosts[0]}}' the first value in the array of interpolated hosts for that API.
# defaultAPIAccessScheme = "http://{{hosts.[0]}}:5000"

# A Google Analytics Tracking ID to be used by the rendered Admin UI.
# googleAnalyticsTrackingId = ''

# Stripe API Secret Key
# stripeSecretKey = ''

# Stripe API Publishable Key
# stripePublishableKey = ''

# Stripe fallback plan (this uses the internal Gateway plan name).
# stripeFallbackPlan = ''

# If the gateway should attempt to migrate non-stripe accounts to Stripe on startup.
# stripeMigrateAccounts = false

# The host the administrative api is accessible via. If left blank the Admin Host is used. Include the protocol (http or https) when using this setting.
# apiHost = ''

## ---------- Proxy ------------
[proxy]
# Whether or not to cache API data when serving proxy calls, default is false.
# cacheApis =

# The number of lines of code to show around script errors, default is 2. NOTE: this option is only available when server = true.
# numErrorLines =

# The timeout in seconds to use for proxy script code, default is 5.
# codeTimeout =

# The domain name for the proxy server. Required when running in server mode, else defaults to lvh.me (default "lvh.me")
# domain = ''

# Whether or not to expose the OS's ENV to proxy code, default is false.
# enableOsEnv =

# The hostname of the proxy server, default is localhost.
# host =

# The timeout in seconds to use for proxied HTTP requests, default is 60.
# httpTimeout =

# The port of the proxy server, default is 5000.
# port =

# The header to send the proxy request ID back in. Not sent if blank.
# requestIdHeader = ''

# The endpoint that responds with 200 - ok to health check requests when healthy. This applies system-wide and cannot be used by APIs as an endpoint. Defaults to /__gw-health-check. Blank will disable the feature.
# healthCheckPath = '/__gw-health-check'

[remoteEndpoint]
# Whether or not http remote endpoints are enabled, defaults is true.
# httpEnabled =

# Whether or not MongoDB remote endpoints are enabled, defaults is true.
# mongoDbEnabled =

# Whether or not MySQL remote endpoints are enabled, defaults is true.
# mySqlEnabled =

# Whether or not PostgreSQL remote endpoints are enabled, defaults is true.
# postgreSqlEnabled =

# Whether or not script remote endpoints are enabled, defaults is true.
# scriptEnabled =

# Whether or not soap remote endpoints are enabled, defaults is true.
# soapEnabled =

# Whether or not MS SQLServer remote endpoints are enabled, defaults is true.
# sqlServerEnabled =

# Whether or not LDAP remote endpoints are enabled, default is true.
# ldapEnabled =

# Whether or not the local store remote endpoints are enabled, default is false.
# storeEnabled =

# Whether or not the SAP Hana remote endpoints are enabled, default is true.
# hanaEnabled =

# Whether or not Docker remote endpoints are enabled, default is true.
# dockerEnabled =

[soap]
# The hostname for the SOAP client, default is 'localhost'
# soapClientHost = ''

# The port number to listen on for the SOAP client, default is 19083
# soapClientPort =

# The JVM options to pass to the JVM on startup that will be used to invoke SOAP services, default is blank
# javaOpts = ''

# The home directory of your JDK 1.8 installation, default is what's in the path
# jdkPath = ''

# The number of worker threads in the JVM that will concurrently process soap requests.  When set to 0, pooling is disabled (i.e. a new thread per request), defaults to 0.
# threadPoolSize =

[docker]
# The amount of memory MB to allocate to each running docker container.
# memory = 128

# The CPU share weight to allocate to each running docker container
# cpuShares = 1024

## ---------- Logging ------------
[bleve]
# Number of days to keep the logs, default is 30
# loggingDeleteAfter =

# The bleve file to store logs, default is 'logs.bleve'
# loggingFile = ''

[elasticLogging]
# Number of days to keep the logs, default is 30
# loggingDeleteAfter =

# The url of the elastic server
# url = ''

## ---------- Error Notification (optional) ------------
[airbrake]
# The API key to use for Airbrake notifications.
# apiKey = ''

# The environment tag under which errors are reported to Airbrake.
# environment = ''

# The ID assigned to your Airbrake project.
# projectId =

## ---------- SMTP (required if registration is enabled or password reset notification is needed) ------------
[smtp]
# The host to be used in email links
# emailHost = ''

# The port to be used in email links
# emailPort =

# The scheme to be used in email links, default is 'http'
# emailScheme = ''

# The password for the smtp server
# password = ''

# The port of the smtp server, default is 25
# port = 25

# The sender of emails from gateway
# sender = ''

# The address or name of the smtp server
# server = ''

# The user name for the smtp server
# user = ''
`

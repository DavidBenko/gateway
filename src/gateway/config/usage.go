package config

var usageStrings = map[string]string{
	"version": "Whether to print the version and quit",
	"config":  "The path to the configuration file",
	"license": "The path to a valid Gateway license key",
	"server":  "Whether or not to run in server mode",

	"db-migrate":         "Whether or not to migrate the database on startup",
	"db-driver":          "The database driver; sqlite or postgres",
	"db-conn-string":     "The connection string for your database",
	"db-max-connections": "The maximum number of connections to use",

	"proxy-host": "The hostname of the proxy server",
	"proxy-port": "The port of the proxy server",

	"proxy-request-id-header": "The header to send the proxy request ID back in. Not sent if blank.",
	"proxy-enable-os-env":     "Whether or not to expose the OS's ENV to proxy code.",

	"proxy-cache-apis": "Whether or not to cache API data when serving proxy calls",

	"proxy-http-timeout":     "The timeout in seconds to use for proxied HTTP requests.",
	"proxy-code-timeout":     "The timeout in seconds to use for proxy script code.",
	"proxy-code-error-lines": "The number of lines of code to show around script errors in dev mode.",

	"admin-path-prefix": "The path prefix the administrative area is accessible under",
	"admin-host":        "The host the administrative area is accessible via",

	"admin-session-name":                  "The name of the cookie to use for sessions.",
	"admin-session-auth-key":              "The auth key to use for sessions. 64 chars recommended. Required.",
	"admin-session-encryption-key":        "The encryption key to use for sessions. 32 chars recommended. If unset, encryption is disabled.",
	"admin-session-auth-key-rotate":       "Same as admin-session-auth-key, to be used during key rotation.",
	"admin-session-encryption-key-rotate": "Same as admin-session-encryption-key, to be used during key rotation.",

	"admin-request-id-header": "The header to send the admin request ID back in. Not sent if blank.",

	"admin-cors-enabled": "Set to false to disable CORS headers from being added to admin responses.",
	"admin-cors-origin":  "The Access-Control-Allow-Origin header value to send with admin responses.",

	"admin-username": "The username to require with HTTP Basic Auth to protect the site admin functionality",
	"admin-password": "The password to require with HTTP Basic Auth to protect the site admin functionality",
	"admin-realm":    "The HTTP Basic realm to use. Optional.",

	"admin-show-version": "Whether or not to expose the Gateway version",

	"admin-add-default-env":  "Whether or not to add a default environment to new APIs in dev mode",
	"admin-default-env-name": "The name of the default environment to create",
	"admin-add-localhost":    "Whether or not to add a localhost host to a the first API in dev mode",

	"elastic-logging-domain":   "The domain name of the elastic server",
	"elastic-logging-username": "The user name for the elastic server",
	"elastic-logging-password": "The password for the elasitic server",

	"bleve-logging-file":         "The bleve file to store logs in",
	"bleve-logging-delete-after": "How long in days to keep logs",
}

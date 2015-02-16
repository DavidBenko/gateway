package config

var usageStrings = map[string]string{
	"config":  "The path to the configuration file",
	"license": "The path to a valid Gateway license key",

	"db-migrate":     "Whether or not to migrate the database on startup",
	"db-driver":      "The database driver; sqlite or postgres",
	"db-conn-string": "The connection string for your database",

	"proxy-host": "The hostname of the proxy server",
	"proxy-port": "The port of the proxy server",

	"proxy-session-auth-key":       "The auth key to use for cookie sessions. 64 chars recommended. If unset, they're disabled.",
	"proxy-session-encryption-key": "The encryption key to use for cookie sessions. 32 chars recommended. If unset, encryption is disabled.",

	"proxy-request-id-header": "The header to send the response ID back in. Not sent if blank.",

	"admin-path-prefix": "The path prefix the administrative area is accessible under",
	"admin-host":        "The host the administrative area is accessible via",

	"admin-session-name":                  "The name of the cookie to use for sessions.",
	"admin-session-auth-key":              "The auth key to use for sessions. 64 chars recommended. Required.",
	"admin-session-encryption-key":        "The encryption key to use for sessions. 32 chars recommended. If unset, encryption is disabled.",
	"admin-session-auth-key-rotate":       "Same as admin-session-auth-key, to be used during key rotation.",
	"admin-session-encryption-key-rotate": "Same as admin-session-encryption-key, to be used during key rotation.",

	"admin-cors-enabled": "Set to false to disable CORS headers from being added to admin responses.",
	"admin-cors-origin":  "The Access-Control-Allow-Origin header value to send with admin responses.",

	"admin-username": "The username to require with HTTP Basic Auth to protect the site admin functionality",
	"admin-password": "The password to require with HTTP Basic Auth to protect the site admin functionality",
	"admin-realm":    "The HTTP Basic realm to use. Optional.",
}

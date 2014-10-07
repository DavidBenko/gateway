package config

import "flag"

func setupFlags(config *Configuration) {
	_ = flag.String("config", "/etc/gateway/gateway.conf", "The path to the configuration file")

	// Proxy configuration
	flag.Int64Var(&config.Proxy.Port, "proxy-port", 5000, "The port of the proxy server")
	flag.StringVar(&config.Proxy.Admin.PathPrefix, "proxy-admin-path-prefix", "/admin/", "The path prefix the administrative area is accessible under")
	flag.StringVar(&config.Proxy.Admin.Host, "proxy-admin-host", "", "The host the administrative area is accessible via")
}

package config

import "flag"

func init() {
	_ = flag.String("config", "/etc/gateway/gateway.conf", "The path to the configuration file")

	// Proxy configuration
	flag.Int64Var(&config.Proxy.Port, "proxy-port", 5000, "The port of the proxy server")
}

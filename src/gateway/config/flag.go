package config

import "flag"

func setupFlags(config *Configuration) {
	_ = flag.String("config", "/etc/gateway/gateway.conf", "The path to the configuration file")

	// General proxy configuration
	flag.StringVar(&config.Proxy.Host, "proxy-host", "localhost", "The hostname of the proxy server")
	flag.Int64Var(&config.Proxy.Port, "proxy-port", 5000, "The port of the proxy server")

	// Proxy admin configuration
	flag.StringVar(&config.Proxy.Admin.PathPrefix, "admin-path-prefix", "/admin/", "The path prefix the administrative area is accessible under")
	flag.StringVar(&config.Proxy.Admin.Host, "admin-host", "", "The host the administrative area is accessible via")

	// Raft configuration
	flag.StringVar(&config.Raft.DataPath, "raft-data-path", "/etc/gateway/data", "The path to the directory where the server data should be stored")
	flag.StringVar(&config.Raft.Leader, "raft-leader", "", "The connection string of the Raft leader this server should join on startup")
	flag.StringVar(&config.Raft.Host, "raft-host", "localhost", "The hostname of the Raft server")
	flag.Int64Var(&config.Raft.Port, "raft-port", 6000, "The port of the Raft server")
}

# Gateway

Welcome to Gateway.

## Developer Setup

### Install Go

On OS X, simply:

    brew update
    brew install go

### Fetch, Build & Run

    git clone git@github.com:AnyPresence/gateway.git
    cd gateway
    make run

### `GOPATH`

The script `gopath.sh` will alter your `GOPATH` to include this project's dependent
paths (the working directory & `_vendor`). To include it in your shell:
    
	source gopath.sh
	
This will allow it to be picked up by your IDE and other tools (I'm using Atom with 
[`go-plus`](https://atom.io/packages/go-plus)).

## Gateway Setup

The Gateway can be configured using a configuration file, environment variables,
command line flags, or all three.

The command line flags take precedence, then the environment variables, then
finally any values set in the configuration file.

All options can be found in `config/flag.go`. Environment variables take the
same format, but upcased and prefixed with `APGATEWAY`. For instance, the
`-proxy-port` flag can be specified with the `APGATEWAY_PROXY_PORT` environment
variable.

The configuration file format is [`toml`](https://github.com/toml-lang/toml).

Run the app with the `--help` flag to see all options.

# Gateway

Welcome to Gateway.

## Developer Setup

### Install Go

On OS X, simply:

    brew update
    brew install go

### Fetch this Repo

    git clone git@github.com:AnyPresence/gateway.git
	
### Add `gateway` to your `GOPATH`

In your `~/.bash_profile` or similar:

	export GOPATH=/path/to/gateway:$GOPATH

### Build and Run

    make run

### IDE

I'm using Atom with [`go-plus`](https://atom.io/packages/go-plus).

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

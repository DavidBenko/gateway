# Gateway

Welcome to Gateway.

## Developer Setup

### Install Go

On OS X:

    brew update
    brew install go
    
Now set up a global `GOPATH`. Here we'll assume it's going to be `~/go`.

    mkdir ~/go
    
In `~/.bash_profile`, add:

    export GOPATH=~/go
    
Now source the file into your local shell and install a few Go tools:

    source ~/.bash_profile
    go get code.google.com/p/go.tools/cmd/godoc
    go get code.google.com/p/go.tools/cmd/vet

### Fetch, Build & Run

    git clone git@github.com:AnyPresence/gateway.git
    cd gateway
    make run

### `GOPATH`

For building and testing, Gateway manages its own `GOPATH` inside the `Makefile`.
Still, sometimes you want to have access to that `GOPATH` outside of `make`.

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

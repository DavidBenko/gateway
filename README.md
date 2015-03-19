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

This runs a Gateway instance using the configuration specified in
`test/gateway.conf`, and sample proxy code stored in `test/examples`.

### Static Assets

`make build` and `make run` both use the static assets stored on disk. This
means that you can edit them and the changes are instantly reflected.

However, this implementation will not find new files. You will need to rebuild
the server if you add a file to one of the static assets folders.

### `GOPATH`

For building and testing, Gateway manages its own `GOPATH` inside the
`Makefile`. Still, sometimes you want to have access to that `GOPATH` outside
of `make`.

The script `gopath.sh` will alter your `GOPATH` to include this project's
dependent paths (the working directory & `_vendor`). To include it in your
shell:

	source gopath.sh

This will allow it to be picked up by your IDE and other tools (I'm using Atom
with [`go-plus`](https://atom.io/packages/go-plus)).

### Admin Front End

I'm using Ember CLI to manage the front end application found in `admin`.

## Gateway Setup

The Gateway can be configured using a configuration file, environment
variables, command line flags, or all three.

The command line flags take precedence, then the environment variables, then
finally any values set in the configuration file.

All options can be found in `config/flag.go`. Environment variables take the
same format, but upcased and prefixed with `APGATEWAY`. For instance, the
`-proxy-port` flag can be specified with the `APGATEWAY_PROXY_PORT` environment
variable.

The configuration file format is [`toml`](https://github.com/toml-lang/toml).

Run the app with the `--help` flag to see all options.

## License Keys

License keys are generated using asymmetric key cryptography. AnyPresence signs
key data with an RSA private key, and the public key is embedded in the binary
for validation. A set of keys for development are included in the `test`
directory. To make compatible keys for production, use:

	ssh-keygen -t rsa -C "AnyPresence Gateway Keypair"

And to extract the public key in a compatible PEM format:

	openssl rsa -in <private key> -pubout -out <public key>

To generate license files, use the `keygen` application in `src/keygen`. For
example, the development license in `test/dev_license` was generated with:

    make keygen
    ./bin/keygen v1 -name="Gateway Development Team" \
	    -company="AnyPresence, Inc" -private-key=./test/license/private_key

## Examples

FIXME Out of date with filesystem code

The `test` directory has several sets of example data. The one being maintained
most frequently right now is the standalone "loopback" server, which serves as
its own backend.

Setup scripts for the "loopback" data and the other examples are run with
`rake`, which you can get with a standard Ruby installation by invoking:

    gem install rake

To seed a fresh server (run with `make run`) with the loopback data:

	cd test/examples
	rake loopback seed

And to update the proxy code after making changes:

    rake loopback update

To completely clear the default Gateway data:

    rake clean


## Packaging

`make package` will build the gateway for all available platforms and put the
resulting binaries in the `build` directory.

The command is using [`gox`](https://github.com/mitchellh/gox). To install you
will need to:

	go get github.com/mitchellh/gox
	gox -build-toolchain

### Building for Linux and Windows with Docker

If you have Docker installed locally, you can build the LinuxAmd64 Dockerfile and then compile in there. You can build the Docker image with:

    docker build --no-cache -t anypresence/gateway:cross-compilation -f dockerfiles/CrossCompilation .

Then, run the following to compile the binary with the dev public key:

    make vet admin assets generate
    docker run --rm -v "$PWD":/usr/src/justapis -w /usr/src/justapis -it anypresence/gateway:cross-compilation

Use the following for the production public:

    docker run -e "LICENSE_PUBLIC_KEY=/usr/src/justapis/public_keys/production" --rm -v "$PWD":/usr/src/justapis -w /usr/src/justapis -it anypresence/gateway:cross-compilation

Your new binary will be at ./build/gateway-{GOOS}-{GOARCH}

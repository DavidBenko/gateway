# Nanoscale.io Gateway Server

This is the source code for the nanoscale.io gateway server. It is the underlying technology behind the hosted nanoscale.io service. We recommend you sign up and use the free hosted version at http://www.nanoscale.io (and also check out the docs at http://devhub.nanoscale.io), to understand the context of how the solution is meant to be used.

You can use this open source version of the nanoscale.io gateway server to run microservices on your own infrastructure.

Before committing new code, be sure to run `scripts/make-hooks` to add gateway's
git hooks, to ensure your commits pass `go fmt` checks, etc.

## Static Assets

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

```bash
source gopath.sh
```

This will allow it to be picked up by your IDE and other tools (I'm using Atom
with [`go-plus`](https://atom.io/packages/go-plus)).

## Gateway Setup

The Gateway can be configured using a configuration file, environment
variables, command line flags, or all three.

The command line flags take precedence, then the environment variables, then
finally any values set in the configuration file.

All options can be found in `config/config.go`. Environment variables take the
same format, but upcased and prefixed with `APGATEWAY`. For instance, the
`-proxy-port` flag can be specified with the `APGATEWAY_PROXY_PORT` environment
variable.

The configuration file format is [`toml`](https://github.com/toml-lang/toml).

Run the app with the `--help` flag to see all options.

## Testing

You can run the unit tests via:

    make test

You can run the integration test suite via:

    make test_all

This will spin up Docker images to test some of the remote endpoint types. It will also run against a local Postgres instance. In order for this test suite to pass please make sure you have Docker installed and you have access to the nanoscale/gateway-ldap repo on the Docker Hub.

Also, make sure that the CA cert file located at `test/ldap/security/cacert.pem` is added to your Keychain and trusted.

If you are using docker-machine locally, make sure to set `DOCKERTEST_LEGACY_DOCKER_MACHINE=1` in your environment.

## Packaging

`make package` will build the gateway for all available platforms and put the
resulting binaries in the `build` directory.

### Building for Linux and Windows with Docker

If you have Docker installed locally, you can build for the target platform using the Makefile. To build for 64bit Linux, do the following:

    make docker_build_linux_amd64_full

If you just want to build the binary without rebuilding the UI assets, the following should suffice:

    make docker_build_linux_amd64

You can build binaries for all target platforms with:

    make docker_build_all

Your new binary will be at ./build/gateway-{GOOS}-{GOARCH}

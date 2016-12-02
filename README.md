# Gateway

Welcome to Gateway.

First, make sure you have the [DEPS](doc/DEPS.md).

To build and test Gateway, read [BUILD.md](doc/BUILD.md).

Before committing new code, be sure to run `scripts/make-hooks` to add gateway's
git hooks, to ensure your commits pass `go fmt` checks, etc.

Then, take a look over the [examples](#examples) and [Admin API doc](doc/Admin API.md).

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

All options can be found in `config/flag.go`. Environment variables take the
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

This will spin up Docker images to test some of the remote endpoint types. It will also run against a local Postgres instance. In order for this test suite to pass please make sure you have Docker installed and you have access to the anypresence/justapis-ldap repo on the Docker Hub.

Also, make sure that the CA cert file located at `test/ldap/security/cacert.pem` is added to your Keychain and trusted.

If you are using docker-machine locally, make sure to set `DOCKERTEST_LEGACY_DOCKER_MACHINE=1` in your environment.

## Packaging

`make package` will build the gateway for all available platforms and put the
resulting binaries in the `build` directory.

### Building for Linux and Windows with Docker

If you have Docker installed locally, you can build the CrossCompilation Dockerfile and then compile in there. You can build the Docker image with:

    docker build --no-cache -t anypresence/gateway:cross-compilation-5.0.0 -f dockerfiles/CrossCompilation .

If you are a collaborator on the Docker Hub repository for anypresence/gateway, you can alternatively just pull the image from there:

    docker pull anypresence/gateway:cross-compilation-5.0.0

Then, run the following to compile the binary with the dev public key:

    make package
    docker run --rm -v "$PWD":/usr/src/justapis -w /usr/src/justapis -it anypresence/gateway:cross-compilation-0.0.1

Use the following for the production public:

    docker run -e "LICENSE_PUBLIC_KEY=/usr/src/justapis/public_keys/production" --rm -v "$PWD":/usr/src/justapis -w /usr/src/justapis -it anypresence/gateway:cross-compilation-5.0.0

Your new binary will be at ./build/gateway-{GOOS}-{GOARCH}

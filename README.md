# Gateway

Welcome to Gateway.

## Developer Setup

### Install Go

On OS X, simply:

    brew update
    brew install go

Now set up your `GOPATH`. For instance, in your `~/.bash_profile`:

	  export GOPATH=~/go

You may also wish to add the `bin` directory to your `PATH`:

    export PATH=$PATH:$GOPATH/bin

### Fetch this Repo

    go get github.com/AnyPresence/gateway

### Build and Run

    go install github.com/AnyPresence/gateway
    $GOPATH/bin/gateway

### IDE

I'm using Atom with [`go-plus`](https://atom.io/packages/go-plus).

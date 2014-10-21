# Many thanks to: http://zduck.com/2014/go-project-structure-and-dependencies/

.PHONY: build doc fmt lint run test vendor_clean vendor_get vendor_update vet install_bindata

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/_vendor:${PWD}:${GOPATH}
export GOPATH

PATH := ${PWD}/_vendor/bin:${PWD}/bin:${PATH}

default: build

assets: install_bindata
	go-bindata -o src/gateway/proxy/admin/bindata.go -pkg admin -debug -prefix "src/gateway/proxy/admin/static/" src/gateway/proxy/admin/static/...
	go-bindata -o src/gateway/model/router_bindata.go -pkg model -debug -prefix "src/gateway/model/static/" src/gateway/model/static/...
	go-bindata -o src/gateway/proxy/bindata.go -pkg proxy -debug -prefix "src/gateway/proxy/static/" src/gateway/proxy/static/...

build: vet assets
	go build -o ./bin/gateway ./src/gateway/main.go
	
doc:
	godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
	goimports ./src/...

run: build
	./bin/gateway -raft-data-path=./test/node -config=./test/gateway.conf

test: assets
	go test ./src/...

vendor_clean:
	rm -dRf ./_vendor/src

# We have to set GOPATH to just the _vendor
# directory to ensure that `go get` doesn't
# update packages in our primary GOPATH instead.
# This will happen if you already have the package
# installed in GOPATH since `go get` will use
# that existing location as the destination.
vendor_get: vendor_clean
	GOPATH=${PWD}/_vendor go get -d -u -v \
	github.com/BurntSushi/toml \
	github.com/gorilla/context \
	github.com/gorilla/handlers \
	github.com/gorilla/mux \
	github.com/robertkrimen/otto \
	github.com/goraft/raft \
	code.google.com/p/go.tools/cmd/goimports \
	github.com/jteeuwen/go-bindata

vendor_update: vendor_get
	rm -rf `find ./_vendor/src -type d -name .git` \
	&& rm -rf `find ./_vendor/src -type d -name .hg` \
	&& rm -rf `find ./_vendor/src -type d -name .bzr` \
	&& rm -rf `find ./_vendor/src -type d -name .svn`

install_bindata:
	if hash go-bindata 2>/dev/null; then : ; else go install github.com/jteeuwen/go-bindata/...; fi;

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
	go vet ./src/...


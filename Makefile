# Many thanks to: http://zduck.com/2014/go-project-structure-and-dependencies/

.PHONY: admin assets build fmt godoc gateway jsdoc keygen package run test vendor_clean vendor_get vendor_update install_bindata install_goimports vet

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/_vendor:${PWD}:${GOPATH}
export GOPATH

PATH := ${PWD}/_vendor/bin:${PWD}/bin:${PATH}

ifneq ($(MAKECMDGOALS), package)
	BINDATA_DEBUG = -debug
	LICENSE_PUBLIC_KEY = test/dev_public_key_assets
endif

default: run

admin:
	# cd admin; ember build -output-path ../src/gateway/admin/static/
	# Placeholder:
	mkdir -p src/gateway/admin/static/js
	mkdir -p src/gateway/admin/static/css
	echo "<html></html>" > src/gateway/admin/static/index.html
	echo "body { color: black; }" > src/gateway/admin/static/css/style.css
	echo "function foo(){}" > src/gateway/admin/static/js/app.js

assets: install_bindata
	go-bindata -o src/gateway/admin/bindata.go -pkg admin $(BINDATA_DEBUG) -prefix "src/gateway/admin/static/" src/gateway/admin/static/...
	go-bindata -o src/gateway/proxy/vm/bindata.go -pkg vm $(BINDATA_DEBUG) -prefix "src/gateway/proxy/vm/static/" src/gateway/proxy/vm/static/...
	go-bindata -o src/gateway/sql/bindata.go -pkg sql $(BINDATA_DEBUG) -prefix "src/gateway/sql/static/" src/gateway/sql/static/...
	go-bindata -o src/gateway/license/bindata.go -pkg license -nocompress -prefix `dirname $(LICENSE_PUBLIC_KEY)/public_key` $(LICENSE_PUBLIC_KEY)
	
generate: install_goimports
	go generate gateway/...
	
build: vet admin assets generate
	go build -o ./bin/gateway ./src/gateway/main.go
	
keygen:
	go build -o ./bin/keygen keygen

package: vet assets
	gox -output="build/binaries/{{.Dir}}_{{.OS}}_{{.Arch}}" -parallel=1 gateway

jsdoc:
	jsdoc -c ./jsdoc.conf -r
	
godoc:
	godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
	goimports ./src/...

run: 
	./bin/gateway -config=./test/gateway.conf -db-migrate

runpg: 
	./bin/gateway -config=./test/gateway.conf -db-migrate -db-driver=postgres -db-conn-string="dbname=gateway_dev sslmode=disable"

test: admin assets generate
	go test ./src/...

test_api: build
	mkdir -p tmp
	./bin/gateway -config=./test/gateway.conf -db-migrate & echo "$$!" > ./tmp/server.pid
	sleep 1
	rspec test/admin-api --color ; status=$$?; kill -9 `cat ./tmp/server.pid`; exit $$status

test_all: test test_api

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
	github.com/gorilla/sessions \
	github.com/robertkrimen/otto \
	github.com/goraft/raft \
	code.google.com/p/go.tools/cmd/goimports \
	github.com/jteeuwen/go-bindata \
	gopkg.in/fsnotify.v1 \
	github.com/jmoiron/sqlx \
	github.com/mattn/go-sqlite3 \
	github.com/lib/pq \
	golang.org/x/crypto/bcrypt

vendor_update: vendor_get
	rm -rf `find ./_vendor/src -type d -name .git` \
	&& rm -rf `find ./_vendor/src -type d -name .hg` \
	&& rm -rf `find ./_vendor/src -type d -name .bzr` \
	&& rm -rf `find ./_vendor/src -type d -name .svn`

install_bindata:
	if hash go-bindata 2>/dev/null; then : ; else go install github.com/jteeuwen/go-bindata/...; fi;

install_goimports:
	if hash goimports 2>/dev/null; then : ; else go install code.google.com/p/go.tools/cmd/goimports/...; fi;

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
	go vet ./src/...


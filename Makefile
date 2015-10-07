# Many thanks to: http://zduck.com/2014/go-project-structure-and-dependencies/

.PHONY: admin assets build fmt godoc gateway jsdoc keygen package run test vendor_clean vendor_get vendor_update install_bindata install_goimports vet soapclient

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/_vendor:${PWD}:${GOPATH}
export GOPATH

PATH := ${PWD}/_vendor/bin:${PWD}/bin:${PATH}

ifndef LICENSE_PUBLIC_KEY
	LICENSE_PUBLIC_KEY = "test/dev_public_key_assets"
endif

ifneq ($(MAKECMDGOALS), package)
	BINDATA_DEBUG = -debug
endif

ifdef TDDIUM_DB_NAME
	POSTGRES_DB_NAME =  $$TDDIUM_DB_NAME
endif
ifndef POSTGRES_DB_NAME
	POSTGRES_DB_NAME = "gateway_test"
endif

default: run

soapclient:
	cd soapclient && ./gradlew shadowJar && rm -f build/libs/gateway-soap-client*.jar

admin:
	cd admin; bundle install; npm install; node_modules/ember-cli/bin/ember build -output-path ../src/gateway/admin/static/ --environment production
	./scripts/templatize-admin.rb src/gateway/admin/static/index.html

assets: install_bindata soapclient
	go-bindata -o src/gateway/admin/bindata.go -pkg admin $(BINDATA_DEBUG) -prefix "src/gateway/admin/static/" src/gateway/admin/static/...
	go-bindata -o src/gateway/proxy/vm/bindata.go -pkg vm $(BINDATA_DEBUG) -prefix "src/gateway/proxy/vm/static/" src/gateway/proxy/vm/static/...
	go-bindata -o src/gateway/sql/bindata.go -pkg sql $(BINDATA_DEBUG) -prefix "src/gateway/sql/static/" src/gateway/sql/static/...
	go-bindata -o src/gateway/soap/bindata.go -pkg soap $(BINDATA_DEBUG) -prefix "soapclient/build/libs/" soapclient/build/libs/...
	go-bindata -o src/gateway/license/bindata.go -pkg license -nocompress -prefix `dirname $(LICENSE_PUBLIC_KEY)/public_key` $(LICENSE_PUBLIC_KEY)

generate: install_goimports
	go generate gateway/...

DeveloperVersionAccounts = 1
DeveloperVersionUsers = 1
DeveloperVersionAPIs = 1
DeveloperVersionProxyEndpoints = 5

LDFLAGS = -ldflags "-X gateway/license.developerVersionAccounts $(DeveloperVersionAccounts)\
 -X gateway/license.developerVersionUsers $(DeveloperVersionUsers)\
 -X gateway/license.developerVersionAPIs $(DeveloperVersionAPIs)\
 -X gateway/license.developerVersionProxyEndpoints $(DeveloperVersionProxyEndpoints)"

build: vet assets generate
	go build $(LDFLAGS) -o ./bin/gateway ./src/gateway/main.go

build_race: vet assets generate
	go build $(LDFLAGS) -race -o ./bin/gateway ./src/gateway/main.go

debug: vet assets generate
	go build $(DEBUG_LDFLAGS) -o ./bin/gateway ./src/gateway/main.go
	dlv exec ./bin/gateway -- -config=./test/gateway.conf -db-migrate

package: vet admin assets generate
	go build -o ./build/gateway ./src/gateway/main.go

keygen:
	go build -o ./bin/keygen keygen

# Prior package for building for multiple architectures
# package: vet assets
# 	gox -output="build/binaries/{{.Dir}}_{{.OS}}_{{.Arch}}" -parallel=1 gateway

jsdoc:
	jsdoc -c ./jsdoc.conf -r

godoc:
	godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
	goimports ./src/...

run:
	./bin/gateway -config=./test/gateway.conf -db-migrate

run_developer:
	./bin/gateway -config=./test/gateway_developer.conf -db-migrate

runpg:
	./bin/gateway -config=./test/gateway.conf -db-migrate -db-driver=postgres -db-conn-string="dbname=gateway_dev sslmode=disable"

test: build
	go test ./src/...

test_api_sqlite_fast:
	mkdir -p tmp
	-rm ./tmp/gateway_test.db
	./bin/gateway -config=./test/gateway.conf -db-migrate -db-conn-string="./tmp/gateway_test.db" -server="true" > /dev/null & echo "$$!" > ./tmp/server.pid
	sleep 1
	rspec test/admin-api; status=$$?; kill -9 `cat ./tmp/server.pid`; exit $$status

test_api_sqlite: build test_api_sqlite_fast

test_api_postgres_fast:
	-dropdb $(POSTGRES_DB_NAME)
	-createdb $(POSTGRES_DB_NAME)
	./bin/gateway -config=./test/gateway.conf -db-migrate -db-driver=postgres -db-conn-string="dbname=$(POSTGRES_DB_NAME) sslmode=disable" -server="true" > /dev/null & echo "$$!" > ./tmp/server.pid
	sleep 1
	rspec test/admin-api; status=$$?; kill -9 `cat ./tmp/server.pid`; exit $$status

test_api_postgres: build test_api_postgres_fast

test_api: test_api_sqlite test_api_postgres

test_api_fast: test_api_sqlite_fast test_api_postgres_fast

test_all: admin assets test test_api

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
	golang.org/x/crypto/bcrypt \
	github.com/denisenkom/go-mssqldb \
	gopkg.in/check.v1 \
	github.com/juju/testing/checkers \
	gopkg.in/mgo.v2 \
	github.com/blevesearch/bleve \
	github.com/mattbaird/elastigo \
	github.com/jackc/pgx \
	github.com/derekparker/delve/cmd/dlv \
	github.com/go-sql-driver/mysql \
	golang.org/x/net/websocket \
	github.com/vincent-petithory/dataurl \
	github.com/gdamore/mangos

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
	./scripts/make-hooks
	./scripts/hooks/pre-commit

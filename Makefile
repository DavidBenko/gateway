# Many thanks to: http://zduck.com/2014/go-project-structure-and-dependencies/

.PHONY: admin assets build fmt godoc gateway jsdoc keygen package run test vendor_clean vendor_get vendor_update install_bindata install_goimports vet soapclient

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/_vendor:${PWD}:${GOPATH}
export GOPATH

PATH := ${PWD}/_vendor/bin:${PWD}/bin:${PATH}

# This path has to be ${HOME}/lib on El Capitan
ORACLE_INSTANT_CLIENT_DIR = ${HOME}/lib

PKG_CONFIG_PATH := $(ORACLE_INSTANT_CLIENT_DIR)
export PKG_CONFIG_PATH

# This must be done for OCI8 to work on Linux
LD_LIBRARY_PATH := ${LD_LIBRARY_PATH}:$(PKG_CONFIG_PATH)
export LD_LIBRARY_PATH

ifndef LICENSE_PUBLIC_KEY
	LICENSE_PUBLIC_KEY = "test/dev_public_key_assets"
endif

ifneq ($(MAKECMDGOALS), $(filter $(MAKECMDGOALS),package release))
	BINDATA_DEBUG = -debug
endif

ifeq ($(MAKECMDGOALS), release)
	LICENSE_PUBLIC_KEY = "public_keys/production"
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
	cd admin; bundle install; npm install; bower install; node_modules/ember-cli/bin/ember build -output-path ../src/gateway/admin/static/ --environment production
	./scripts/templatize-admin.rb src/gateway/admin/static/index.html

assets: install_bindata soapclient
	go-bindata -o src/gateway/admin/bindata.go -pkg admin $(BINDATA_DEBUG) -prefix "src/gateway/admin/static/" src/gateway/admin/static/...
	go-bindata -o src/gateway/core/bindata.go -pkg core $(BINDATA_DEBUG) -prefix "src/gateway/core/static/" src/gateway/core/static/...
	go-bindata -o src/gateway/sql/bindata.go -pkg sql $(BINDATA_DEBUG) -prefix "src/gateway/sql/static/" src/gateway/sql/static/...
	go-bindata -o src/gateway/soap/bindata.go -pkg soap $(BINDATA_DEBUG) -prefix "soapclient/build/libs/" soapclient/build/libs/...
	go-bindata -o src/gateway/license/bindata.go -pkg license -nocompress -prefix `dirname $(LICENSE_PUBLIC_KEY)/public_key` $(LICENSE_PUBLIC_KEY)
	go-bindata -o src/gateway/names/bindata.go -pkg names $(BINDATA_DEBUG) -prefix "src/gateway/names/dictionary/" src/gateway/names/dictionary/...
	go-bindata -o src/gateway/mail/bindata.go -pkg mail $(BINDATA_DEBUG) -prefix "src/gateway/mail/static/" src/gateway/mail/static/...
	go-bindata -o src/gateway/stats/sql/bindata.go -pkg sql $(BINDATA_DEBUG) -prefix "src/gateway/stats/sql/static/" src/gateway/stats/sql/static/...

generate: install_goimports install_peg
	go generate gateway/...

DeveloperVersionAccounts = 1
DeveloperVersionUsers = 1
DeveloperVersionAPIs = 1
DeveloperVersionProxyEndpoints = 20

LDFLAGS = -ldflags "-X gateway/license.developerVersionAccounts=$(DeveloperVersionAccounts)\
 -X gateway/license.developerVersionUsers=$(DeveloperVersionUsers)\
 -X gateway/license.developerVersionAPIs=$(DeveloperVersionAPIs)\
 -X gateway/license.developerVersionProxyEndpoints=$(DeveloperVersionProxyEndpoints)"

build: vet assets generate install_oracle_client
	go build $(LDFLAGS) -o ./bin/gateway ./src/gateway/main.go

build_integration_images:
	docker build -t anypresence/justapis-ldap test/ldap

build_race: vet assets generate
	go build $(LDFLAGS) -race -o ./bin/gateway ./src/gateway/main.go

build_tail:
	go build -o ./bin/tail ./src/tail/main.go

debug: vet assets generate
	go build $(DEBUG_LDFLAGS) -o ./bin/gateway ./src/gateway/main.go
	dlv exec ./bin/gateway -- -config=./test/gateway.conf -db-migrate

package: vet admin assets generate
	go build -o ./build/gateway ./src/gateway/main.go

release: vet admin assets generate
	go build -ldflags="-s -w" -o ./build/gateway ./src/gateway/main.go

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

test_api_sqlite_fast: build_tail
	mkdir -p tmp
	-rm ./tmp/gateway_test.db
	-rm ./tmp/gateway_log.txt
	./bin/gateway -config=./test/gateway.conf \
	  -db-migrate \
	  -db-conn-string="./tmp/gateway_test.db" \
	  -proxy-domain="example.com" \
	  -proxy-domain="example.com" \
	  -server="true" > ./tmp/gateway_log.txt & \
	  echo "$$!" > ./tmp/server.pid

	# Sleep until we see "Server listening" or time out
	# ./bin/tail --verbose -timeout=5 -filename="./foo/bar" "Server listening|Error"
	./bin/tail -file ./tmp/gateway_log.txt "Server listening" || kill `cat ./tmp/server.pid`

	rspec test/admin-api; status=$$?; kill `cat ./tmp/server.pid`; exit $$status

test_api_sqlite: build test_api_sqlite_fast

test_api_postgres_fast: build_tail
	mkdir -p tmp
	-rm ./tmp/gateway_log.txt
	-dropdb $(POSTGRES_DB_NAME)
	-createdb $(POSTGRES_DB_NAME)
	./bin/gateway -config=./test/gateway.conf \
	  -db-migrate \
	  -db-driver=postgres \
	  -db-conn-string="dbname=$(POSTGRES_DB_NAME) sslmode=disable" \
	  -proxy-domain="example.com" \
	  -server="true" > ./tmp/gateway_log.txt & \
	  echo "$$!" > ./tmp/server.pid

	# Sleep until we see "Server listening" or time out
	# ./bin/tail --verbose -timeout=5 -filename="./foo/bar" "Server listening|Error"
	./bin/tail -file ./tmp/gateway_log.txt "Server listening" || kill `cat ./tmp/server.pid`

	rspec test/admin-api; status=$$?; kill `cat ./tmp/server.pid`; exit $$status

test_api_postgres: build test_api_postgres_fast

test_api: test_api_sqlite test_api_postgres

test_api_fast: test_api_sqlite_fast test_api_postgres_fast

test_all: admin assets test test_api test_integration

test_integration: build test_integration_fast

test_integration_fast: build_tail
	docker run -p 389:389 -d anypresence/justapis-ldap > ./tmp/.containerid
	mkdir -p tmp
	-rm ./tmp/gateway_log.txt
	-rm ./tmp/gateway_test.db
	./bin/gateway -config=./test/gateway.conf \
	  -db-migrate \
		-db-conn-string="./tmp/gateway_test.db" > ./tmp/gateway_log.txt & \
		echo "$$!" > ./tmp/server.pid

	./bin/tail -file ./tmp/gateway_log.txt "Server listening" || (kill `cat ./tmp/server.pid`; docker kill `cat ./tmp/.containerid`)

	go test -v -ldflags "-X gateway/test/integration.IntegrationTest=true -X gateway/test/integration/ldap.ldapSetupFile=`pwd`/test/ldap/setup.ldif -X gateway/test/integration.ApiImportDirectory=`pwd`/test/integration" ./src/gateway/test/integration/...; \
		status=$$?; \
		docker kill `cat ./tmp/.containerid`; \
		kill `cat ./tmp/server.pid`; \
		exit $$status

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
	github.com/gdamore/mangos \
	github.com/xeipuuv/gojsonschema \
	gopkg.in/airbrake/gobrake.v2 \
	github.com/boltdb/bolt \
	gopkg.in/tomb.v1 \
	github.com/hpcloud/tail \
	github.com/ory-am/dockertest \
	github.com/go-ldap/ldap \
	github.com/pointlander/peg \
	github.com/SAP/go-hdb/driver \
	golang.org/x/net/http2 \
	golang.org/x/crypto/pkcs12 \
	github.com/sideshow/apns2 \
	github.com/alexjlockwood/gcm \
	github.com/mattn/go-oci8 \
	github.com/garyburd/redigo \
	github.com/davecgh/go-spew/spew

vendor_update: vendor_get
	rm -rf `find ./_vendor/src -type d -name .git` \
	&& rm -rf `find ./_vendor/src -type d -name .hg` \
	&& rm -rf `find ./_vendor/src -type d -name .bzr` \
	&& rm -rf `find ./_vendor/src -type d -name .svn`

install_bindata:
	if hash go-bindata 2>/dev/null; then : ; else go install github.com/jteeuwen/go-bindata/...; fi;

install_goimports:
	if hash goimports 2>/dev/null; then : ; else go install code.google.com/p/go.tools/cmd/goimports/...; fi;

install_peg:
	if hash peg 2>/dev/null; then : ; else go install github.com/pointlander/peg; fi;

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
	./scripts/make-hooks
	./scripts/hooks/pre-commit

install_oracle_client:
	#first argument is directory to save instant client, second is the package config file source
	#which is processed and saved as oci8.pc in the same argument directory
	./scripts/install_oracle_instant_client.rb $(ORACLE_INSTANT_CLIENT_DIR) contrib/oci8.pc

start_oracle:
	# Starts the docker container named 'orcl' running Oracle 12c on a machine named oracle
	# DB named 'ORCL' on port 1521 with login system/manager
	./scripts/start_oracle.sh

# Many thanks to: http://zduck.com/2014/go-project-structure-and-dependencies/

.PHONY: admin assets build fmt godoc gateway jsdoc keygen package run test vendor_clean vendor_get vendor_update install_bindata install_goimports vet soapclient cross_compile release docker_full_release docker_clean_bin install_vet docker_cross_compile docker_compilation_prep docker_compile_only

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/_vendor:${PWD}:${GOPATH}
export GOPATH

PATH := ${PWD}/_vendor/bin:${PWD}/bin:${PATH}

ifneq ($(MAKECMDGOALS), $(filter $(MAKECMDGOALS),package release docker_compilation_prep))
	BINDATA_DEBUG = -debug
endif

ifdef TDDIUM_DB_NAME
	POSTGRES_DB_NAME =  $$TDDIUM_DB_NAME
endif
ifndef POSTGRES_DB_NAME
	POSTGRES_DB_NAME = "gateway_test"
endif
ifndef POSTGRES_STATS_DB_NAME
	POSTGRES_STATS_DB_NAME = "gateway_stats_test"
endif

default: run

admin:
	cd admin; bundle install; npm install; bower install; node_modules/ember-cli/bin/ember build -output-path ../src/gateway/admin/static/ --environment production
	./scripts/templatize-admin.rb src/gateway/admin/static/index.html

docker_admin:
	cd admin; bundle install; npm rebuild; bower install; node_modules/ember-cli/bin/ember build -output-path ../src/gateway/admin/static/ --environment production
	./scripts/templatize-admin.rb src/gateway/admin/static/index.html

assets: install_bindata
	go-bindata -o src/gateway/admin/bindata.go -pkg admin $(BINDATA_DEBUG) -prefix "src/gateway/admin/static/" src/gateway/admin/static/...
	go-bindata -o src/gateway/core/bindata.go -pkg core $(BINDATA_DEBUG) -prefix "src/gateway/core/static/" src/gateway/core/static/...
	go-bindata -o src/gateway/sql/bindata.go -pkg sql $(BINDATA_DEBUG) -prefix "src/gateway/sql/static/" src/gateway/sql/static/...
	go-bindata -o src/gateway/names/bindata.go -pkg names $(BINDATA_DEBUG) -prefix "src/gateway/names/dictionary/" src/gateway/names/dictionary/...
	go-bindata -o src/gateway/mail/bindata.go -pkg mail $(BINDATA_DEBUG) -prefix "src/gateway/mail/static/" src/gateway/mail/static/...
	go-bindata -o src/gateway/model/bindata.go -pkg model $(BINDATA_DEBUG) -prefix "src/gateway/model/static/" src/gateway/model/static/...

generate: install_goimports install_peg
	go generate gateway/...

build: vet assets generate
	go build -o ./bin/gateway ./src/gateway/main.go

build_integration_images:
	docker build -t anypresence/justapis-ldap test/ldap

build_race: vet assets generate
	go build -race -o ./bin/gateway ./src/gateway/main.go

build_tail:
	go build -o ./bin/tail ./src/tail/main.go

debug: vet assets generate
	go build -o ./bin/gateway ./src/gateway/main.go
	dlv exec ./bin/gateway -- -config=./test/gateway.conf -db-migrate

package: vet admin assets generate
	go build -o ./build/gateway ./src/gateway/main.go

release: vet admin assets generate
	go build -ldflags="-s -w" -o ./build/gateway ./src/gateway/main.go

docker_clean_bin:
	rm -rf _vendor/bin

docker_compilation_prep: docker_clean_bin vet docker_admin assets generate

docker_binary_release: docker_compile_only

docker_compile_only:
	go build -ldflags="-s -w" -v -o ./build/gateway ./src/gateway/main.go

docker_build_admin:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && make docker_admin"

docker_build_prereqs:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && make docker_compilation_prep"

docker_build_linux_amd64_full: docker_build_prereqs docker_build_linux_amd64 docker_pack_executables

docker_build_linux_amd64:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=gcc make docker_binary_release && mv ./build/gateway ./build/gateway-linux-amd64"

docker_build_linux_386_full: docker_build_prereqs docker_build_linux_386 docker_pack_executables

docker_build_linux_386:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=linux GOARCH=386 CGO_ENABLED=1 CC=gcc make docker_binary_release && mv ./build/gateway ./build/gateway-linux-386"

docker_build_windows_amd64_full: docker_build_prereqs docker_build_windows_amd64 docker_pack_executables

docker_build_windows_amd64:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=\"x86_64-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp\" make docker_binary_release && mv ./build/gateway ./build/gateway-windows-amd64.exe"

docker_build_windows_386_full: docker_build_prereqs docker_build_windows_386 docker_pack_executables

docker_build_windows_386:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=\"i686-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp\" make docker_binary_release && mv ./build/gateway ./build/gateway-windows-386.exe"

docker_build_armv5_full: docker_build_prereqs docker_build_armv5 docker_pack_executables

docker_build_armv5:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc make docker_binary_release && mv ./build/gateway ./build/gateway-linux-armv5"

docker_build_armv6_full: docker_build_prereqs docker_build_armv6 docker_pack_executables

docker_build_armv6:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc make docker_binary_release && mv ./build/gateway ./build/gateway-linux-armv6"

docker_build_armv7_full: docker_build_prereqs docker_build_armv7 docker_pack_executables

docker_build_armv7:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc make docker_binary_release && mv ./build/gateway ./build/gateway-linux-armv7"

docker_pack_executables:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && upx -9 ./build/gateway-*"

docker_brute_pack_executables:
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:compile-5.4.0 /bin/bash -c ". /root/.bashrc && upx --brute ./build/gateway-*"

docker_build_all_full: docker_build_prereqs docker_build_all docker_clean_bin

docker_build_all: docker_build_linux_amd64 docker_build_linux_386 docker_build_windows_amd64 docker_build_windows_386 docker_build_armv5 docker_build_armv6 docker_build_armv7

docker_run:
	# Make sure docker_build_linux_amd64_full or docker_build_linux_amd64 has been run prior or there will be no binary to run within the container.
	mkdir -p ./build/docker
	docker run --rm -v $(PWD):/usr/src/justapis -w /usr/src/justapis -p 5000:5000 -it nanoscale/gateway:run-5.3.0 ./build/gateway-linux-amd64 -bleve-logging-file=/tmp/logs.bleve -store-conn-string=/tmp/store.db -db-conn-string=./build/docker/gateway.db -proxy-host=0.0.0.0

docker_test:
	- docker rm -f gateway-test-docker
	docker run --privileged --name gateway-test-docker -d docker:dind
	docker run --rm --link gateway-test-docker:docker -v $(PWD):/usr/src/justapis -w /usr/src/justapis -it nanoscale/gateway:test-5.3.0 /bin/bash -c ". /root/.bashrc && export DOCKER_HOST='tcp://docker:2375' && make test_all"

keygen:
	go build -o ./bin/keygen keygen

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
	-rm ./tmp/gateway_stats_test.db
	-rm ./tmp/gateway_log.txt
	./bin/gateway -config=./test/gateway.conf \
	  -db-migrate -stats-migrate \
	  -db-conn-string="./tmp/gateway_test.db" \
		-stats-conn-string="./tmp/gateway_stats_test.db" \
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
	-dropdb $(POSTGRES_STATS_DB_NAME)
	-createdb $(POSTGRES_STATS_DB_NAME)
	./bin/gateway -config=./test/gateway.conf \
	  -db-migrate \
	  -db-driver=postgres \
	  -db-conn-string="dbname=$(POSTGRES_DB_NAME) sslmode=disable" \
		-stats-migrate \
	  -stats-driver=postgres \
	  -stats-conn-string="dbname=$(POSTGRES_STATS_DB_NAME) sslmode=disable" \
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
	-rm ./tmp/gateway_stats_test.db
	./bin/gateway -config=./test/gateway.conf \
	  -db-migrate -stats-migrate \
		-db-conn-string="./tmp/gateway_test.db" -stats-conn-string="./tmp/gateway_stats_test.db" > ./tmp/gateway_log.txt & \
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
	github.com/go-mangos/mangos \
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
	github.com/garyburd/redigo \
	github.com/fsouza/go-dockerclient \
	github.com/nanoscaleio/surgemq \
	github.com/surge/glog \
	github.com/stripe/stripe-go \
	github.com/clbanning/mxj \
	github.com/google/go-gcm \
	github.com/edganiukov/fcm \
	github.com/Microsoft/go-winio \
	github.com/Azure/go-ansiterm \
	github.com/aymerick/raymond

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
	rm -f ./_vendor/bin/peg
	go install github.com/pointlander/peg

install_vet:
	if hash vet 2>/dev/null; then : ; else go get golang.org/x/tools/cmd/vet; fi;

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet: install_vet
	./scripts/make-hooks
	./scripts/hooks/pre-commit

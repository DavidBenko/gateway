# Building and Testing Gateway

Make sure you have all the [DEPS](DEPS.md).

## Postgres

Make sure your Postgres instance is running.

> Mac OS

```bash
pg_start
```

> Ubuntu Linux

```bash
sudo service postgresql start
```

## Project deps

```bash
cd admin
npm install
ember install
bundle install
ember build
cd ..
bundle install
```

## Build

```bash
make admin
make build
```

## Test

```bash
make <test_all | test_api_sqlite | test_api_postgres>
```

## Run

Configuration is specified in `test/gateway.conf`.

Sample proxy code is stored in `test/examples`.

```bash
make run
```

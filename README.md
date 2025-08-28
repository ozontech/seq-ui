# Seq UI Server

## First steps

### Third party dependencies

To download third party binary dependencies (e.g. proto-gen-go, grpc-gateway, vtproto) run `make deps`.

### Build

Use `make build` to build the application.
It compiles the application into the `bin` folder.

### Local run

Use `docker compose up -d` to start all necessary containers.

Use `make migrate` and `make migrate-ch` to run postgres and clickhouse migrations.

Use `make run` to run the application with default config path `seq-ui-server-config.yaml`.

SeqUI starts listening for requests:
* HTTP - `http://localhost:5555`
* gRPC - `http://localhost:5556`
* Debug - `http://localhost:5557`

SwaggerUI available at `http://localhost:5557/docs`.

## Migrations

There are two options to run migrations:
1. Change variable `MIGRATION_DSN`/`MIGRATION_DSN_CLICKHOUSE` in Makefile and use `make migrate`/`make migrate-ch`
2. Use `make migrate MIGRATION_DSN="{your_dsn}"`/`make migrate-ch MIGRATION_DSN_CLICKHOUSE="{your_dsn}"`

Similar options for rolling back the last migration, but use `make undo-last-migration`/`make undo-last-migration-ch` command.

## Lint before commit 

Run `make lint`. It will check *.proto files with [buf](https://buf.build/) and *.go files with golangci-lint.

## Configuring

The application can be configured via yaml file. See the details in [docs/config.md](docs/config.md).

Example of config file can be found [here](config/app.example.yaml).

## Tracing initialization

Run this before start to enable tracing:
```shell
export TRACING_SERVICE_NAME=seq-ui-server-local
export TRACING_SAMPLER_PARAM=1.0
export JAEGER_AGENT_PORT=6831
export JAEGER_AGENT_HOST=127.0.0.1
```

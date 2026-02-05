# IMAGE ?= ghcr.io/ozontech/seq-ui
IMAGE ?= gitlab-registry.ozon.ru/sre/images/seq-ui-server
VERSION ?= v0.48.1-7-g69cf #$(shell git describe --abbrev=4 --dirty --always --tags)
TIME := $(shell date '+%Y-%m-%d_%H:%M:%S')

MIGRATION_DSN ?= postgresql://localhost/postgres?sslmode=disable&user=postgres&password=postgres
MIGRATION_DSN_CLICKHOUSE ?= tcp://default@localhost:9000/seq_ui_server

LOCAL_BIN := $(CURDIR)/bin

GOLANGCI_LINT_VER=2.8.0
PROTOC_GEN_GO_VER=1.34.2
PROTOC_GEN_GO_GRPC_VER=1.4.0
MOCKGEN_VER=0.6.0
SWAG_VER=1.16.2

export GOBIN=$(LOCAL_BIN)

.PHONY: build-bin
build-bin:
	$(info Building binaries)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-trimpath \
		-ldflags "-X github.com/ozontech/seq-ui/buildinfo.Version=${VERSION} \
				  -X github.com/ozontech/seq-ui/buildinfo.BuildTime=${TIME}" \
		-o ${LOCAL_BIN}/seq-ui \
		./cmd/seq-ui

.PHONY: build-image
build-image:
	$(info Building image)
	docker buildx build --platform linux/amd64 \
		--build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIME=${TIME} \
		--file build/package/Dockerfile \
		-t ${IMAGE}:${VERSION} \
		.

.PHONY: push-image
push-image: build-image
	docker push ${IMAGE}:${VERSION}

.PHONY: build-migration-image
build-migration-image:
	$(info Building pg-migration image)
	docker buildx build --platform linux/amd64 \
		--build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIME=${TIME} \
		--file migration/Dockerfile \
		-t ${IMAGE}-pg-migration:${VERSION} \
		.

.PHONY: push-migration-image
push-migration-image: build-migration-image
	docker push ${IMAGE}-pg-migration:${VERSION}

.PHONY: run
run: .check-config
	go run ./cmd/seq-ui -config=config/config.local.yaml

.PHONY: .check-config
.check-config:
	@if [ ! -f config/config.local.yaml ]; then\
		cp config/config.example.yaml config/config.local.yaml;\
	fi

.PHONY: clean
clean:
	rm -rf bin

.PHONY: test
test:
	CGO_ENABLED=1 go test -count=1 -v -race ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VER) run \
		--new-from-rev=origin/main --config=.golangci.yaml --timeout=5m

.PHONY: lint-full
lint-full:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VER) run \
		--config=.golangci.yaml --timeout=5m

.PHONY: deps
deps: .protoc-plugins .install-tools

.PHONY: mock
mock:
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
		-source=internal/pkg/repository/repository.go \
		-destination=internal/pkg/repository/mock/repository.go
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
		-source=internal/pkg/repository_ch/repository.go \
		-destination=internal/pkg/repository_ch/mock/repository.go
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
		-source=internal/pkg/client/seqdb/client.go \
		-destination=internal/pkg/client/seqdb/mock/client.go
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
    	-source=internal/pkg/client/seqdb/seqproxyapi/v1/seq_proxy_api_grpc.pb.go \
    	-destination=internal/pkg/client/seqdb/seqproxyapi/v1/mock/seq_proxy_api_grpc.pb.go
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
		-source=internal/pkg/cache/cache.go \
		-destination=internal/pkg/cache/mock/cache.go
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
		-source=internal/app/auth/oidc.go \
		-destination=internal/app/auth/mock/oidc.go
	PATH="$(LOCAL_BIN):$(PATH)" mockgen \
		-source=internal/app/auth/jwt.go \
		-destination=internal/app/auth/mock/jwt.go

.PHONY: protoc
protoc:
	PATH="$(LOCAL_BIN):$(PATH)" protoc \
		-I api \
		--go_out=pkg --go_opt=paths=source_relative \
		--go-grpc_out=pkg --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
		$(shell find api -name '*.proto' | grep -v vendor)
	PATH="$(LOCAL_BIN):$(PATH)" protoc \
		-I internal/pkg/client/seqdb/seqproxyapi \
    	--go_out=internal/pkg/client/seqdb/seqproxyapi --go_opt=paths=source_relative \
    	--go-grpc_out=internal/pkg/client/seqdb/seqproxyapi --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
    	$(shell find internal/pkg/client/seqdb/seqproxyapi -name '*.proto')

.PHONY: swagger
swagger:
	PATH="$(LOCAL_BIN):$(PATH)" swag fmt \
		-g registrar.go \
		-d internal/api
	PATH="$(LOCAL_BIN):$(PATH)" swag init \
		-q \
		-g registrar.go \
		-d internal/api \
		-o swagger \
		-ot json

.PHONY: generate
generate: protoc mock swagger

.PHONY: .protoc-plugins
.protoc-plugins:
	$(info Downloading protoc plugins)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VER)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VER)

.PHONY: .install-tools
.install-tools:
	$(info Downloading tools)
	go install go.uber.org/mock/mockgen@v$(MOCKGEN_VER)
	go install github.com/swaggo/swag/cmd/swag@v$(SWAG_VER)

.PHONY: cover
cover:
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

.PHONY: migrate
migrate:
	goose -dir='migration' postgres "$(MIGRATION_DSN)" up

.PHONY: undo-last-migration
undo-last-migration:
	goose -dir='migration' postgres "$(MIGRATION_DSN)" down

.PHONY: migrate-ch
migrate-ch:
	goose -dir='migration_ch' clickhouse "$(MIGRATION_DSN_CLICKHOUSE)" up

.PHONY: undo-last-migration-ch
undo-last-migration-ch:
	goose -dir='migration_ch' clickhouse "$(MIGRATION_DSN_CLICKHOUSE)" down

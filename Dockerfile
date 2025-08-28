ARG APP_IMAGE=ubuntu:latest

# Build
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

ARG VERSION
ARG BUILD_TIME
ARG TARGETARCH

WORKDIR /seq-ui-server

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH ${TARGETARCH:-amd64}

RUN go build -trimpath \
    -ldflags "-X github.com/ozontech/seq-ui/buildinfo.Version=${VERSION} -X github.com/ozontech/seq-ui.BuildTime=${TIME}" \
    -o seq-ui-server ./cmd/seq-ui-server

# Deploy
FROM $APP_IMAGE

WORKDIR /seq-ui-server

COPY --from=build /seq-ui-server/migration /seq-ui-server/migration

COPY --from=build /seq-ui-server/seq-ui-server /seq-ui-server/seq-ui-server

ENTRYPOINT ["./seq-ui-server"]

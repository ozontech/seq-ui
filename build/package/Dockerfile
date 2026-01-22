ARG APP_IMAGE=ubuntu:latest

# Build
FROM --platform=$BUILDPLATFORM golang:1.25 AS build

ARG VERSION
ARG BUILD_TIME
ARG TARGETARCH

WORKDIR /seq-ui

COPY go.mod go.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/

RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${TARGETARCH:-amd64}

RUN go build -trimpath \
    -ldflags "-X github.com/ozontech/seq-ui/buildinfo.Version=${VERSION} -X github.com/ozontech/seq-ui/buildinfo.BuildTime=${BUILD_TIME}" \
    -o seq-ui ./cmd/seq-ui

# Deploy
FROM $APP_IMAGE

WORKDIR /seq-ui

COPY --from=build /seq-ui/seq-ui /seq-ui/seq-ui

ENTRYPOINT ["./seq-ui"]

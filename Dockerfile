# --------------------------------------------------------------------- dev ---

FROM golang:1.25-alpine AS dev

RUN apk --update add --no-cache ca-certificates git

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/vidhya-service

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ------------------------------------------------------------------- debug ---

FROM dev AS debug

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/cespare/reflex@latest

CMD reflex -R "__debug_bin" -s -- sh -c "dlv debug --headless --continue --accept-multiclient --listen :40000 --api-version=2 --log ./src"

# ------------------------------------------------------------------- build ---

FROM dev AS build

ARG VERSION_BUILD=0.0.0-dev
ENV VERSION_BUILD $VERSION_BUILD

ARG VERSION_SHA=0000000000000000000000000000000000000000
ENV VERSION_SHA $VERSION_SHA

RUN go build -trimpath -ldflags " \
    -X vidhya-service/src/version.BuildNumber=$VERSION_BUILD \
    -X vidhya-service/src/version.BuildCommitSHA=$VERSION_SHA \
    " -o vidhya-service ./src

# -------------------------------------------------------------------- test ---

FROM dev AS test

# No -race here: CGO_ENABLED=0 (see the dev stage) keeps this image
# dependency-free (no gcc needed on Alpine); the race detector requires
# cgo. Run `go test -race ./...` on your host/CI if you want that check.
RUN go test -trimpath -coverprofile=coverage.out ./...
RUN go tool cover -func=coverage.out

# ------------------------------------------------------------------- utils ---

FROM alpine:3 AS utils

RUN apk --update add --no-cache ca-certificates tzdata

# ----------------------------------------------------------------- release ---

FROM alpine:3 AS release

RUN apk upgrade --no-cache

EXPOSE 8080

COPY --from=utils /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=utils /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /opt/app/

COPY --from=build /go/src/vidhya-service/vidhya-service ./

ENTRYPOINT [ "/opt/app/vidhya-service" ]

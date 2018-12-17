FROM golang:1.11-alpine as builder

RUN apk update && apk add --no-cache openssh git jq curl gcc libc-dev build-base

ADD . /go/src/github.com/centrifuge/go-centrifuge
WORKDIR /go/src/github.com/centrifuge/go-centrifuge

RUN go install -ldflags "-X github.com/centrifuge/go-centrifuge/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...

FROM alpine:latest

RUN apk update && apk add --no-cache jq curl

WORKDIR /root/
COPY --from=builder /go/bin/centrifuge .
COPY build/scripts/docker/entrypoint.sh /root

VOLUME ["/root/config"]

ENTRYPOINT ["/root/entrypoint.sh"]



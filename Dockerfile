FROM golang:1.10-alpine as builder

RUN apk update && apk add --no-cache openssh git jq curl gcc libc-dev build-base

ADD . /go/src/github.com/CentrifugeInc/go-centrifuge
WORKDIR /go/src/github.com/CentrifugeInc/go-centrifuge

RUN go install ./centrifuge

FROM alpine:latest

RUN apk update && apk add --no-cache jq curl

WORKDIR /root/
COPY --from=builder /go/bin/centrifuge .
COPY deployments/entrypoint.sh /root

VOLUME ["/root/config"]

ENTRYPOINT ["/root/entrypoint.sh"]



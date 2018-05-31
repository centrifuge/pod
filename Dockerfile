FROM golang:1.10-alpine as builder

RUN apk update && apk add --no-cache openssh git jq curl gcc libc-dev build-base
RUN curl https://glide.sh/get | sh

ADD . /go/src/github.com/CentrifugeInc/go-centrifuge
WORKDIR /go/src/github.com/CentrifugeInc/go-centrifuge

RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts

RUN go install ./centrifuge

FROM alpine:latest

RUN apk update && apk add --no-cache git jq curl

WORKDIR /root/
COPY --from=builder /go/bin/centrifuge .
COPY deployments/entrypoint.sh /root

VOLUME ["/root/config"]

ENTRYPOINT ["/root/entrypoint.sh"]



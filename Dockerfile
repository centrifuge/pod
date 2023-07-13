FROM golang:1.20-buster as builder

RUN apt-get -y update && apt-get -y upgrade && apt-get -y install wget && apt-get install ca-certificates -y

ADD . /go/src/github.com/centrifuge/pod
WORKDIR /go/src/github.com/centrifuge/pod

RUN go install -ldflags "-X github.com/centrifuge/pod/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...

FROM debian:buster-slim
RUN apt-get -y update && apt-get -y upgrade && apt-get install ca-certificates -y

WORKDIR /root/
COPY --from=builder /go/bin/centrifuge .
COPY build/scripts/docker/entrypoint.sh /root
RUN chmod +x /root/entrypoint.sh

VOLUME ["/root/config"]

ENTRYPOINT ["/root/entrypoint.sh"]



FROM golang:1.15-buster as builder

RUN apt-get -y update && apt-get -y upgrade && apt-get -y install wget && apt-get install ca-certificates -y

ADD . /go/src/github.com/centrifuge/go-centrifuge
WORKDIR /go/src/github.com/centrifuge/go-centrifuge
RUN wget -P /go/bin/ https://storage.googleapis.com/centrifuge-dev-public/subkey-rc6 && mv /go/bin/subkey-rc6 /go/bin/subkey  && chmod +x /go/bin/subkey

RUN go install -ldflags "-X github.com/centrifuge/go-centrifuge/version.gitCommit=`git rev-parse HEAD`" ./cmd/centrifuge/...

FROM debian:buster-slim
RUN apt-get -y update && apt-get -y upgrade && apt-get install ca-certificates -y

WORKDIR /root/
COPY --from=builder /go/bin/centrifuge .
COPY build/scripts/docker/entrypoint.sh /root

COPY --from=builder /go/bin/subkey /usr/local/bin/
RUN subkey --version

VOLUME ["/root/config"]

ENTRYPOINT ["/root/entrypoint.sh"]



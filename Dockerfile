FROM golang:1.10.2 as go-build

RUN curl -L -o /tmp/dep-linux-amd64 https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && install -m 0755 /tmp/dep-linux-amd64 /usr/local/bin/dep
RUN apt-get update && apt-get install -y \
  libsystemd-dev

WORKDIR /go/src/github.com/kontena/pharos-host-upgrades

COPY Gopkg.* ./
RUN dep ensure -vendor-only

COPY . ./
RUN go install -v .



# must match with go-build base image
FROM debian:stretch

COPY --from=go-build /go/bin/pharos-host-upgrades /usr/local/bin/pharos-host-upgrades

CMD ["/usr/local/bin/pharos-host-upgrades"]
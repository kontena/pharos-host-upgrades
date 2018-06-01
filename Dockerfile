FROM golang:1.10.2 as go-build

RUN curl -L -o /tmp/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && install -m 0755 /tmp/dep /usr/local/bin/dep
RUN apt-get update && apt-get install -y \
  libsystemd-dev

WORKDIR /go/src/github.com/kontena/pharos-host-upgrades

COPY Gopkg.* ./
RUN dep ensure -vendor-only

COPY . ./
RUN go install -v .


# download kubectl
FROM buildpack-deps:stretch-curl as kube-dl

RUN curl -L -o /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.10.3/bin/linux/amd64/kubectl && install -m 0755 /tmp/kubectl /usr/local/bin/kubectl



# must match with go-build base image
FROM debian:stretch

COPY --from=kube-dl /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=go-build /go/bin/pharos-host-upgrades /usr/local/bin/pharos-host-upgrades

CMD ["/usr/local/bin/pharos-host-upgrades"]

# Cross-architecture builds
# builds without CGO => reduced featureset
ARG BUILD_ARCH=amd64
ARG ARCH=amd64

FROM golang:1.10.2 as go-build
ARG ARCH
ARG BUILD_ARCH

RUN curl -L -o /tmp/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-${BUILD_ARCH} && install -m 0755 /tmp/dep /usr/local/bin/dep

WORKDIR /go/src/github.com/kontena/pharos-host-upgrades

COPY Gopkg.* ./
RUN dep ensure -vendor-only

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go install -v .


# download kubectl
FROM buildpack-deps:stretch-curl as kube-dl
ARG ARCH

RUN curl -L -o /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.10.3/bin/linux/${ARCH}/kubectl && install -m 0755 /tmp/kubectl /usr/local/bin/kubectl



FROM scratch
ARG ARCH

COPY --from=kube-dl /usr/local/bin/kubectl /bin/kubectl
COPY --from=go-build /go/bin/linux_${ARCH}/pharos-host-upgrades /bin/pharos-host-upgrades

CMD ["/bin/pharos-host-upgrades"]

ARG VPP_VERSION=v25.10
ARG UBUNTU_VERSION=22.04

FROM ubuntu:${UBUNTU_VERSION} as vppbuild

RUN set -eux;\
    apt-get update && env DEBIAN_FRONTEND=noninteractive TZ=US/Central apt-get install -y \
    asciidoc \
    ca-certificates \
    git \
    iproute2 \
    iputils-ping \
    make \
    python3 \
    sudo \
    tcpdump \
    rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/FDio/vpp.git

WORKDIR /vpp

ARG VPP_VERSION
RUN git checkout ${VPP_VERSION}

RUN set -eux; \
    env DEBIAN_FRONTEND=noninteractive TZ=US/Central UNATTENDED=y make install-dep
RUN set -eux; \
    make pkg-deb \
    ./src/scripts/version > /vpp/VPP_VERSION

FROM vppbuild as vppinstall

RUN set -eux; \
    env VPP_INSTALL_SKIP_SYSCTL=false apt install -f -y --no-install-recommends /vpp/build-root/*.deb ; \
    rm -rf /var/lib/apt/lists/*; \
    rm -rf /pkg

FROM golang:1.25-alpine3.22 as binapi-generator

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOBIN=/bin

ARG GOVPP_VERSION

COPY . /govpp

WORKDIR /govpp

RUN go build -o /bin/binapi-generator ./cmd/binapi-generator

FROM binapi-generator as gen

COPY --from=vppinstall /usr/share/vpp/api/ /usr/share/vpp/api/
COPY --from=vppinstall /vpp/VPP_VERSION /VPP_VERSION

WORKDIR /gen/binapi

CMD VPP_VERSION=$(cat /VPP_VERSION) go generate .

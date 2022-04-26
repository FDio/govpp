ARG VPP_VERSION=v21.06
ARG UBUNTU_VERSION=20.04

FROM ubuntu:${UBUNTU_VERSION} as vppbuild
ARG VPP_VERSION
RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive TZ=US/Central apt-get install -y git make python3 sudo asciidoc
RUN git clone https://github.com/FDio/vpp.git
WORKDIR /vpp
RUN git checkout ${VPP_VERSION}
#COPY patch/ patch/
#RUN test -x "patch/patch.sh" && ./patch/patch.sh || exit 1
RUN DEBIAN_FRONTEND=noninteractive TZ=US/Central UNATTENDED=y make install-dep
RUN make pkg-deb
RUN ./src/scripts/version > /vpp/VPP_VERSION

FROM vppbuild as vppinstall
#COPY --from=vppbuild /var/lib/apt/lists/* /var/lib/apt/lists/
#COPY --from=vppbuild [ "/vpp/build-root/libvppinfra_*_amd64.deb", "/vpp/build-root/vpp_*_amd64.deb", "/vpp/build-root/vpp-plugin-core_*_amd64.deb", "/vpp/build-root/vpp-plugin-dpdk_*_amd64.deb", "/pkg/"]
#RUN VPP_INSTALL_SKIP_SYSCTL=false apt install -f -y --no-install-recommends /pkg/*.deb ca-certificates iputils-ping iproute2 tcpdump; \
#    rm -rf /var/lib/apt/lists/*; \
#    rm -rf /pkg
RUN VPP_INSTALL_SKIP_SYSCTL=false apt install -f -y --no-install-recommends /vpp/build-root/*.deb ca-certificates iputils-ping iproute2 tcpdump; \
    rm -rf /var/lib/apt/lists/*; \
    rm -rf /pkg

FROM golang:1.17.9-alpine3.15 as binapi-generator
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
CMD  VPP_VERSION=$(cat /VPP_VERSION) go generate .

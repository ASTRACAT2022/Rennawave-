FROM golang:1.26.3-alpine AS xray-build

WORKDIR /src/Xray-core

COPY Xray-core/go.mod Xray-core/go.sum ./
COPY Xray-core/third_party/aesingflow/go.mod ./third_party/aesingflow/go.mod
COPY Xray-core/third_party/aesingflow/third_party/quic-go/go.mod ./third_party/aesingflow/third_party/quic-go/go.mod
RUN go mod download

COPY Xray-core ./
RUN CGO_ENABLED=0 go build -p 1 -trimpath -ldflags="-s -w -buildid=" -o /usr/local/bin/xray ./main


FROM mcr.microsoft.com/devcontainers/base:jammy

ARG S6_OVERLAY_VERSION=3.2.0.2

RUN apt-get update && apt-get install -y \
    curl \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN S6_ARCH="$(uname -m)" \
    && curl -L -o /tmp/s6-noarch.tar.xz "https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz" \
    && curl -L -o /tmp/s6-arch.tar.xz "https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-${S6_ARCH}.tar.xz" \
    && xz -dc /tmp/s6-noarch.tar.xz | tar -C / -xpf - \
    && xz -dc /tmp/s6-arch.tar.xz | tar -C / -xpf - \
    && rm -f /tmp/s6-noarch.tar.xz /tmp/s6-arch.tar.xz


ENV NVM_DIR=/root/.nvm
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash \
    && . $NVM_DIR/nvm.sh \
    && nvm install v24.12.0 \
    && nvm alias default v24.12.0 \
    && nvm use default


ENV PATH="/root/.nvm/versions/node/v24.12.0/bin:${PATH}"

RUN curl -L https://raw.githubusercontent.com/remnawave/scripts/main/scripts/install-latest-xray.sh | bash -s -- v26.5.3 \
    && ln -s /usr/local/bin/xray /usr/local/bin/rw-core

# Keep the installer-provided geo assets, but use the locally built Xray core
# with the vendored AesingFlow transport in every development environment.
COPY --from=xray-build /usr/local/bin/xray /usr/local/bin/xray


ARG ASN_LMDB_URL=https://github.com/remnawave/asn-index/releases/latest/download/asn-prefixes-lmdb.tar.gz

RUN mkdir -p /var/log/xray /var/lib/rnode/xray /app /usr/local/share/asn \
    && echo '{}' > /var/lib/rnode/xray/xray-config.json \
    && curl -L ${ASN_LMDB_URL} -o /tmp/asn-prefixes-lmdb.tar.gz \
    && tar -xzf /tmp/asn-prefixes-lmdb.tar.gz -C /usr/local/share/asn \
    && rm -f /tmp/asn-prefixes-lmdb.tar.gz

COPY rootfs/ /
RUN chmod +x /etc/s6-overlay/scripts/init-env.sh \
    /etc/s6-overlay/s6-rc.d/xray/run \
    /etc/s6-overlay/s6-rc.d/xray-log/run

WORKDIR /app


EXPOSE 24000

# /init brings up the s6 service tree (init-env + xray[down] + xray-log),
# then runs CMD. Node is started manually by the developer (see DEV_ENV.md).
ENTRYPOINT ["/init"]

CMD ["tail", "-f", "/dev/null"]

FROM golang:1.22.5-bookworm AS builder

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o login-server

FROM debian:bookworm-slim

# update ca-certificates
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ARG user=login-server
ARG group=login-server
ARG uid=10000
ARG gid=10001

# If you bind mount a volume from the host or a data container,
# ensure you use the same uid
RUN groupadd -g ${gid} ${group} \
    && useradd -l -u ${uid} -g ${gid} -m -s /bin/bash ${user}

USER ${user}

COPY --from=builder --chown=${uid}:${gid} /src/login-server /usr/bin/login-server

EXPOSE 8080

ENTRYPOINT [ "login-server" ]

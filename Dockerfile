FROM golang:1.24.4-bookworm AS builder

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o ory-oathkeeper-login

FROM debian:bookworm-slim

# update ca-certificates
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ARG user=ory-oathkeeper-login
ARG group=ory-oathkeeper-login
ARG uid=10000
ARG gid=10001

# If you bind mount a volume from the host or a data container,
# ensure you use the same uid
RUN groupadd -g ${gid} ${group} \
    && useradd -l -u ${uid} -g ${gid} -m -s /bin/bash ${user}

USER ${user}

COPY --from=builder --chown=${uid}:${gid} /src/ory-oathkeeper-login /usr/bin/ory-oathkeeper-login
COPY --from=builder --chown=${uid}:${gid} /src/config/config.yaml /etc/ory-oathkeeper-login/config.yaml

ENV LOG_COLOR=true
ENV LOG_LEVEL=info
ENV LOG_FORMAT=console
ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT ["ory-oathkeeper-login"]

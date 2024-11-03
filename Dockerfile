FROM golang:1.22.3-bookworm AS builder

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
ARG uid=1000
ARG gid=1000

# If you bind mount a volume from the host or a data container,
# ensure you use the same uid
RUN groupadd -g ${gid} ${group} \
    && useradd -l -u ${uid} -g ${gid} -m -s /bin/bash ${user}

ENV HOME /home/login-server

RUN mkdir -p ${HOME}

COPY --from=builder /src/login-server ${HOME}/login-server

RUN chown -R ${uid}:${gid} ${HOME}

USER ${user}

WORKDIR ${HOME}

EXPOSE 8080

ENTRYPOINT [ "./login-server" ]

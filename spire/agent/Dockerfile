FROM ghcr.io/spiffe/spire-agent:1.7.0 as spire-agent

FROM debian:bullseye

RUN apt-get update && apt-get install -y \
    ca-certificates curl vim netcat iputils-ping lsof \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/spire/bin /opt/spire/conf/agent /tmp/spire-agent/public

COPY --from=spire-agent /opt/spire /opt/spire

WORKDIR /opt/spire

ENTRYPOINT ["/opt/spire/bin/spire-agent"]
CMD ["run", "-config", "/opt/spire/conf/agent/agent.conf"]

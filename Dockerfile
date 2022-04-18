FROM ubuntu:20.04
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && \
    apt-get -y install gcc mono-mcs && \
    rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

ENTRYPOINT ["/statuspage"]
COPY statuspage /

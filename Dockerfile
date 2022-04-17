FROM golang:1.15.1 AS builder

# no need to include cgo bindings
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

# add ca certificates and timezone data files
# hadolint ignore=DL3008
RUN apt-get install --yes --no-install-recommends ca-certificates tzdata

# add unprivileged user
RUN adduser --shell /bin/true --uid 1000 --disabled-login --no-create-home --gecos '' app \
  && sed -i -r "/^(app|root)/!d" /etc/group /etc/passwd \
  && sed -i -r 's#^(.*):[^:]*$#\1:/sbin/nologin#' /etc/passwd


FROM scratch

# add-in our timezone data file
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# add-in our unprivileged user
COPY --from=builder /etc/passwd /etc/group /etc/shadow /etc/

# add-in our ca certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 1000

ENTRYPOINT ["/statuspage"]
COPY statuspage /

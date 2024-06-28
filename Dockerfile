FROM golang:1.22.4 AS builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download && go mod verify

COPY . /app

# Build the binary with CGO disabled because the builder image has a newer GLIBC
# than the iRODS client image.
RUN make build CGO_ENABLED=0 && \
    make install CGO_ENABLED=0 GOBIN=/usr/local/bin/

FROM ghcr.io/wtsi-npg/ub-18.04-baton-irods-4.2.11:latest

COPY --from=builder /usr/local/bin/ /usr/local/bin/

USER appuser

ENTRYPOINT ["/usr/local/bin/valet"]

CMD ["--version"]

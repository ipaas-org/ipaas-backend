# FROM docker as docker

FROM golang:1.22.1-alpine3.19 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.22.1-alpine3.19 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build  -o /bin/app .

# FROM scratch
# use these 3 lines for debugging purposes
FROM ubuntu:24.10 
RUN apt update
RUN apt install -y git curl iputils-ping
# RUN mkdir -p /ipaas
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /home/app
CMD ["/home/app"]

EXPOSE 8082

# Metadata
LABEL org.opencontainers.image.vendor="IPaaS" \
    org.opencontainers.image.source="https://github.com/ipaas-org/ipaas-backend" \
    org.opencontainers.image.title="ipaas-backend" \
    org.opencontainers.image.description="A monolith version of ipaas" \
    org.opencontainers.image.version="v1.0.0"
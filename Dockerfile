FROM golang:1.21-bullseye AS builder
RUN apt update && apt install -y libsystemd-dev libvirt-dev
WORKDIR /tmp/src
COPY go.mod .
COPY go.sum .
RUN go get -v libvirt.org/go/libvirt && go mod download
COPY . .
ARG VERSION=unknown
RUN CGO_ENABLED=1 go build -mod=readonly -ldflags "-X main.version=$VERSION" -o coroot-node-agent .

FROM debian:bullseye
RUN apt update && apt install -yq ca-certificates g++ libvirt0 libvirt-dev && apt clean
COPY --from=builder /tmp/src/coroot-node-agent /usr/bin/coroot-node-agents
ENTRYPOINT ["coroot-node-agents"]
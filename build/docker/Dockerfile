# syntax=docker/dockerfile:1
FROM --platform=$TARGETPLATFORM golang:1.20-alpine as builder
ARG TARGETARCH
ARG TARGETOS
WORKDIR /workspace
# RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go install github.com/grpc-ecosystem/grpc-health-probe@latest
COPY go.mod go.mod
COPY go.sum go.sum
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY internal/ internal/
COPY third_party/ third_party/
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o app ./cmd/

FROM --platform=$TARGETPLATFORM alpine
WORKDIR /
COPY configs/ configs/
COPY --from=builder /workspace/app .
# COPY --from=builder /go/bin/grpc-health-probe .
EXPOSE 7070
ENTRYPOINT ["/app"]

CMD ["-c", "/configs/daemon.yaml"]
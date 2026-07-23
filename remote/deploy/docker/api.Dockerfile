FROM golang:1.23-bookworm@sha256:167053a2bb901972bf2c1611f8f52c44d5fe7e762e5cab213708d82c421614db AS builder

WORKDIR /src
COPY shared/go/go.mod ./shared/go/
COPY remote/api/go.mod remote/api/go.sum ./remote/api/
RUN cd remote/api && go mod download
COPY shared/go ./shared/go
COPY remote/api ./remote/api
RUN cd remote/api \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/remote-api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/remote-migrate ./cmd/migrate

FROM debian:bookworm-slim@sha256:7b140f374b289a7c2befc338f42ebe6441b7ea838a042bbd5acbfca6ec875818

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tzdata \
    && ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo Asia/Shanghai > /etc/timezone \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --uid 1001 --create-home --shell /usr/sbin/nologin lighthouse
WORKDIR /app
COPY --from=builder /out/remote-api /app/remote-api
COPY --from=builder /out/remote-migrate /app/remote-migrate
USER lighthouse
EXPOSE 5031
ENTRYPOINT ["/app/remote-api"]
CMD ["-f", "/app/config/server.yaml"]

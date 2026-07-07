FROM golang:1.25 AS builder

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG ALL_PROXY
ARG http_proxy
ARG https_proxy
ARG all_proxy

WORKDIR /src

COPY gc-api/go.mod gc-api/go.sum ./gc-api/
WORKDIR /src/gc-api
RUN go mod download

WORKDIR /src
COPY oldmodel ./oldmodel
COPY gc-api ./gc-api

WORKDIR /src/gc-api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gc-api ./cmd/api

FROM golang:1.25

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG ALL_PROXY
ARG http_proxy
ARG https_proxy
ARG all_proxy

WORKDIR /app

COPY --from=builder /out/gc-api /app/gc-api
RUN mkdir -p /app/config

ENV GIN_MODE=release
ENV TZ=Asia/Shanghai

EXPOSE 15013

CMD ["/app/gc-api", "-f", "/app/config/api.config.yaml"]

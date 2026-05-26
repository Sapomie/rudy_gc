FROM golang:1.25 AS build

WORKDIR /src
COPY gc-api/go.mod gc-api/go.sum ./
RUN go mod download

COPY gc-api ./
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/gc-api ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S app \
 && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/gc-api /app/gc-api
COPY deploy/gc-api.config.example.yaml /app/config.yaml

RUN chown -R app:app /app
USER app

EXPOSE 2041
ENTRYPOINT ["./gc-api","-f","/app/config.yaml","-addr",":2041"]

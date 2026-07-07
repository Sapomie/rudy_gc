FROM golang:1.25

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG ALL_PROXY
ARG http_proxy
ARG https_proxy
ARG all_proxy

RUN apt-get update \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app/gc-web

COPY gc-web/package.json ./package.json
COPY gc-web ./

RUN mkdir -p /app/z_text

ENV PORT=15012
ENV TZ=Asia/Shanghai

EXPOSE 15012

CMD ["node", "server.mjs"]

FROM node:22-bookworm-slim

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG ALL_PROXY
ARG http_proxy
ARG https_proxy
ARG all_proxy

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web ./
RUN npm run build

ENV PORT=15012
ENV TZ=Asia/Shanghai

EXPOSE 15012

CMD ["npm", "run", "dev"]

FROM node:22-bookworm-slim

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG ALL_PROXY
ARG http_proxy
ARG https_proxy
ARG all_proxy

WORKDIR /app/gc-web-next02

COPY gc-web-next02/package.json gc-web-next02/package-lock.json ./
RUN npm ci

COPY gc-web-next02 ./
RUN npm run build

ENV PORT=15012
ENV TZ=Asia/Shanghai

EXPOSE 15012

CMD ["npm", "run", "dev"]

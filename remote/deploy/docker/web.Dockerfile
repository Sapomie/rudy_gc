FROM node:22-bookworm-slim@sha256:6c74791e557ce11fc957704f6d4fe134a7bc8d6f5ca4403205b2966bd488f6b3 AS builder

WORKDIR /src
COPY remote/web/new/package.json remote/web/new/package-lock.json ./remote/web/new/
RUN cd remote/web/new && npm ci
COPY remote/web ./remote/web
RUN cd remote/web && npm run build

FROM nginx:1.27-alpine@sha256:65645c7bb6a0661892a8b03b89d0743208a18dd2f3f17a54ef4b76fb8e2f2a10

COPY remote/deploy/docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /src/remote/web/new/dist /usr/share/nginx/html
EXPOSE 5032

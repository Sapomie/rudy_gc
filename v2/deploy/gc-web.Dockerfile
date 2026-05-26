FROM node:22-alpine AS build

WORKDIR /src
COPY gc-web/package*.json ./
RUN npm ci

COPY gc-web ./
RUN npm run build

FROM node:22-alpine

WORKDIR /app
COPY --from=build /src/dist /app/dist
COPY deploy/web-server.mjs /app/web-server.mjs

ENV PORT=2040
ENV API_PREFIX=/api/gc/v2
ENV API_ORIGIN=http://gc-api:2041

EXPOSE 2040
CMD ["node","/app/web-server.mjs"]

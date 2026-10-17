# syntax=docker/dockerfile:1.7

FROM golang:1.25.12-bookworm AS api-builder
WORKDIR /src/api
COPY api/go.mod api/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY api/ ./
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ytv-api ./cmd/ytv

FROM debian:bookworm-slim AS api
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates chromium curl fonts-noto-cjk \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --uid 10001 --shell /usr/sbin/nologin ytv
COPY --from=api-builder /out/ytv-api /usr/local/bin/ytv-api
ENV YTV_API_ADDRESS=0.0.0.0:8097 \
    YTV_BROWSER_CAPTURE_ENABLED=true \
    YTV_BROWSER_EXECUTABLE=/usr/bin/chromium \
    YTV_BROWSER_NO_SANDBOX=true \
    HOME=/home/ytv
USER ytv
EXPOSE 8097
ENTRYPOINT ["/usr/local/bin/ytv-api"]

FROM node:24.20-bookworm-slim AS web-builder
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
COPY web/vendor ./vendor
RUN --mount=type=cache,target=/pnpm/store pnpm config set store-dir /pnpm/store \
    && pnpm install --frozen-lockfile
COPY web/ ./
COPY api/contracts /src/api/contracts
RUN pnpm build

FROM node:24.20-bookworm-slim AS web
WORKDIR /app
COPY --from=web-builder --chown=node:node /src/web/.output ./.output
ENV NODE_ENV=production HOST=0.0.0.0 PORT=3000
USER node
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]

FROM golang:1.25.12-alpine AS migration-builder
RUN CGO_ENABLED=0 go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

FROM alpine:3.22 AS migrate
RUN apk add --no-cache ca-certificates
COPY --from=migration-builder /go/bin/migrate /usr/local/bin/migrate
COPY api/internal/postgres/migrations /migrations
ENTRYPOINT ["/usr/local/bin/migrate", "-path=/migrations"]

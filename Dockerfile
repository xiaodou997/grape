# ---- Stage 1: Build frontend ----
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

# ---- Stage 2: Build backend ----
FROM golang:1.25-alpine AS backend-builder

# CGO is required for SQLite (go-sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Copy built frontend assets into embed path
COPY --from=frontend-builder /app/web/dist ./web/dist

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o grape ./cmd/grape

# ---- Stage 3: Runtime ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /app/grape ./grape

# Default data directory (override via volume)
RUN mkdir -p /data

EXPOSE 4873

ENV GIN_MODE=release

ENTRYPOINT ["./grape"]
CMD ["--config", "/app/configs/config.yaml"]

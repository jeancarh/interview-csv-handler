FROM node:24-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS api
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=web /src/frontend/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/shop ./cmd/api && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=api /out/shop /app/shop
COPY --from=api --chown=nonroot:nonroot /out/data /data
COPY testdata/products.csv /app/testdata/products.csv
ENV PORT=8080 \
    DB_PATH=/data/shop.db
EXPOSE 8080
VOLUME ["/data"]
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/app/shop", "healthcheck"]
ENTRYPOINT ["/app/shop"]

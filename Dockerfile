FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY go.mod go.sum ./
# retech-authkit é módulo privado: o proxy público só tem versões antigas em
# cache. O Railway passa as variáveis do serviço como build args declarados
# com ARG; com GITHUB_TOKEN (fine-grained, só leitura de conteúdo no
# retech-auth-kit) o git autentica. O token vive só neste estágio builder, que
# não vai para a imagem final. ARG em vez de --mount=type=secret: o builder
# Metal do Railway não aceita RUN --mount.
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/theretechlabs
RUN if [ -n "$GITHUB_TOKEN" ]; then \
      git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi \
    && go mod download \
    && rm -f /root/.gitconfig
COPY . .
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/retechfin/retechfin-api/internal/version.Version=${APP_VERSION}" \
    -o /out/api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H app
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY migrations /app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/app/api"]

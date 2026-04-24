# 方案 C：改代码后执行 docker compose up -d --build api 重建镜像
FROM golang:1.24-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH="/go/bin:${PATH}"
RUN swag init -g main.go

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/techmemo .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=build /out/techmemo .
COPY --from=build /src/docs ./docs

ENV APP_SERVER_HOST=0.0.0.0
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=20s \
  CMD wget -qO- http://127.0.0.1:8080/health > /dev/null || exit 1

ENTRYPOINT ["/app/techmemo"]

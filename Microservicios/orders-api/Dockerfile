FROM golang:1.22-bookworm AS build
WORKDIR /app
COPY go.mod go.sum* ./
COPY main.go .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o orders-api main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/orders-api .
EXPOSE 8002

CMD ["./orders-api"]
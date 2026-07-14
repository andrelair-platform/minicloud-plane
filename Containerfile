FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /minicloud-plane ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /minicloud-plane /minicloud-plane
EXPOSE 8080
ENTRYPOINT ["/minicloud-plane"]

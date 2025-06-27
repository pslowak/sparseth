FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o sparseth ./cmd/sparseth

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/sparseth .

ENTRYPOINT ["./sparseth"]

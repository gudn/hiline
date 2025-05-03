FROM golang:alpine AS builder
WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o app .

FROM alpine

WORKDIR /root
COPY --from=builder /app/app .
EXPOSE 8080

CMD ["./app", "-addr", ":8080"]

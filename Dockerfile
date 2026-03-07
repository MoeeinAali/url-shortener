FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o api ./cmd/api
RUN go build -o projector ./cmd/projector

CMD ["./api"]
FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/main cmd/api/main.go

EXPOSE 8080

CMD ["./main"]
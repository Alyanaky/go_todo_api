FROM golang:1.16

RUN apt-get update && apt-get install -y sqlite3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o todo-api

CMD ["./go_todo_api"]

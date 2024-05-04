FROM golang:1.22.2

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -o run-app .

CMD ["./run-app"]
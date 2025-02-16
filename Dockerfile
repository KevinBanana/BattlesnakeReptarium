FROM golang:1.24-alpine

WORKDIR /app

ENV ENV production

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /battlesnake-app

# default port
EXPOSE 80

CMD ["/battlesnake-app"]

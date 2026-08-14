FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /battlesnake-app

# default port (PORT env overrides the listen port at runtime)
EXPOSE 8080

CMD ["/battlesnake-app"]

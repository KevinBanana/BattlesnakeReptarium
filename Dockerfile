FROM golang:1.26-bookworm AS build

ARG ONNXRUNTIME_VERSION=1.29.0
ADD https://github.com/microsoft/onnxruntime/releases/download/v${ONNXRUNTIME_VERSION}/onnxruntime-linux-x64-${ONNXRUNTIME_VERSION}.tgz /tmp/onnxruntime.tgz
RUN mkdir -p /opt/onnxruntime && tar -xzf /tmp/onnxruntime.tgz -C /opt/onnxruntime --strip-components=1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=1 GOOS=linux go build -o /battlesnake-app

FROM debian:bookworm-slim

COPY --from=build /opt/onnxruntime/lib/libonnxruntime.so* /usr/local/lib/
RUN ldconfig

WORKDIR /app
COPY --from=build /battlesnake-app /battlesnake-app
COPY config/ ./config/
COPY weights/ ./weights/

ENV ONNXRUNTIME_LIB=/usr/local/lib/libonnxruntime.so

# default port (PORT env overrides the listen port at runtime)
EXPOSE 8080

CMD ["/battlesnake-app"]

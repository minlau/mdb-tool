FROM golang:1.19.3-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /binary


FROM alpine:3.16.2

WORKDIR /

COPY --from=build /binary /binary
COPY --from=build /app/development/config-docker.json /config.json

EXPOSE 8080

ENTRYPOINT ["/binary"]
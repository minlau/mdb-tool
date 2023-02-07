FROM node:19-alpine AS js-build

WORKDIR /app/code

RUN yarn set version berry

COPY /web/ui/js-app/yarn.lock .

RUN yarn install

COPY /web/ui/js-app .

RUN yarn build

FROM golang:1.20.0-alpine AS go-build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

COPY --from=js-build /app/static /app/web/ui/static

RUN go build -o binary ./cmd/mdb-tool

FROM scratch

WORKDIR /app

COPY --from=go-build /app/binary /app/binary
COPY --from=go-build /app/development/config.json /app/config.json

EXPOSE 8080

ENTRYPOINT ["/app/binary"]
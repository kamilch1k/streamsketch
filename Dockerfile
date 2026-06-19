FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -trimpath -ldflags="-s -w" -o /out/streamsketch-api ./cmd/api

FROM alpine:3.22
RUN adduser -D -H streamsketch
USER streamsketch
COPY --from=build /out/streamsketch-api /usr/local/bin/streamsketch-api
EXPOSE 8080
ENTRYPOINT ["streamsketch-api"]

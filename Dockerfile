FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/tls-terminated-proxy .

FROM scratch
WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/tls-terminated-proxy /tls-terminated-proxy

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/tls-terminated-proxy"]

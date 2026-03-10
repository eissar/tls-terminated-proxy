FROM golang:1.22-alpine AS build
WORKDIR /src

# no deps
# COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/reverse-proxy .

FROM scratch
WORKDIR /
COPY --from=build /out/tls-terminated-proxy /tls-terminated-proxy

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/reverse-proxy"]

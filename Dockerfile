ARG GO_VERSION=1.18.3

FROM golang:latest-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN cd cmd && CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o . main.go

RUN chmod +x cmd/main

FROM scratch AS final

COPY --from=builder app/cmd/main /main

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 443
EXPOSE 80

VOLUME ["/cert-cache"]

ENTRYPOINT ["/main"]

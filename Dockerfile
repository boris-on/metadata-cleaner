ARG GO_VERSION=1.18.3

FROM golang:${GO_VERSION}-alpine AS builder

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

COPY /public ./

EXPOSE 80

ENTRYPOINT ["/main"]

FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN cd cmd && go build
RUN chmod +x cmd/cmd
CMD ["cmd/cmd"]
FROM golang:alpine

WORKDIR /form3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD CGO_ENABLED=0 go test ./...
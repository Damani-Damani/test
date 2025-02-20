FROM golang:1.23-bookworm AS builder
WORKDIR /controlserver
COPY go.mod go.sum .

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./controlserver


FROM alpine
COPY --from=builder /controlserver/controlserver /controlserver
EXPOSE 50051
EXPOSE 8080
ENTRYPOINT ["/controlserver"]

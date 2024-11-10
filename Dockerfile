FROM golang:alpine

RUN apk add openssl --no-cache

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /proxy
CMD ["/proxy"]

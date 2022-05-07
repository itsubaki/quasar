FROM golang:latest

WORKDIR /go/src/app
ENV GO111MODULE=on

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/app .

CMD ["./app"]

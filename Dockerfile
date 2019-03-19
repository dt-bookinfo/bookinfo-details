FROM golang:1.11.4 as builder
WORKDIR /go/src/github.com/dirkwall/bookDetails
RUN go get -d -v github.com/gorilla/mux
COPY main.go .
RUN GOOS=linux go build -a -ldflags -linkmode=external -installsuffix cgo -o bookDetails .

FROM debian:jessie-slim
COPY --from=builder /go/src/github.com/dirkwall/bookDetails/bookDetails /
CMD ["/bookDetails"]
EXPOSE 9080

FROM golang:alpine
RUN mkdir /socket-server
COPY . /socket-server
WORKDIR /socket-server
RUN go build -o main .
CMD ["/socket-server/main"]
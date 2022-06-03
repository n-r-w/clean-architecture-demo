FROM golang:1.18.3

WORKDIR /logserver

RUN apt-get update
RUN apt-get install -y htop mc tilde nano

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ .
RUN make build

EXPOSE 8080

CMD ["./logserver"]
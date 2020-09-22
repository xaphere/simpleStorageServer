FROM golang:1


ENV PROJECT=file-server

COPY . /
WORKDIR /

RUN go build -mod=readonly -a -o /${PROJECT}

CMD ["/file-server"]
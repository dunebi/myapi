FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY . ./

RUN go mod download

# 실행파일 빌드
RUN go build -o myapi .

ENTRYPOINT [ "./myapi" ]

# docker build --tag myapi:1.0 .
# docker run -d --name myapi --network="host" myapi:1.0
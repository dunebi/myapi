FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY go.mod go.sum main.go ./
COPY JWT ./JWT

RUN go mod download

# main 실행파일 빌드
RUN go build -o main .

WORKDIR /dist

RUN cp /build/main .

FROM scratch

COPY --from=builder /dist/main .

ENTRYPOINT [ "./main" ]

# docker build --tag myapi:1.0 .
# docker run --name myapi --network="host" myapi:1.0
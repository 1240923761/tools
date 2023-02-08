FROM golang:1.18-alpine AS builder

LABEL stage=gobuild

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE auto
WORKDIR /build

COPY . .

RUN go mod download

RUN go build -ldflags="-s -w" -o tz-backend ./main.go

FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk add --no-cache curl tzdata

ENV TZ Asia/Shanghai

WORKDIR /app

COPY --from=builder /build/tz-backend ./server/tz-backend

RUN ls -alsh

CMD ["./server/tz-backend"]
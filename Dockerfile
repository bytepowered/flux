FROM golang:1.14.4 as builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /go/release/
ADD . .
RUN make build

# Build image
FROM alpine:3.12.0 as prod
LABEL AUTHOR="yongjia.chen@hotmail.com"
ENV LANG='en_US.UTF-8' LANGUAGE='en_US:en' LC_ALL='en_US.UTF-8'
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.cloud.tencent.com/g' /etc/apk/repositories
RUN apk --no-cache add ca-certificates
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /go/release/main/conf.d /app/conf.d
COPY --from=builder /go/release/build/flux /app/

WORKDIR /app
ENV APP_LOG_CONF_FILE=/app/conf.d/log.yml
ENV CONF_CONSUMER_FILE_PATH=/app/conf.d/dubbo.yml
ENV CONF_PROVIDER_FILE_PATH=/app/conf.d/dubbo.yml
EXPOSE 8080
CMD ["./flux"]


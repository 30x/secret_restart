# FROM scratch
FROM alpine:3.4

#Install ssl certs so we can connect to ssl services
RUN apk update
RUN apk add ca-certificates
RUN update-ca-certificates


COPY build/secretrestart /

RUN chmod 755 /secretrestart

CMD ["./secretrestart"]
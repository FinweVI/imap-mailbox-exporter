FROM golang:1.15-alpine as BUILD

RUN mkdir /build/
COPY . /build/
WORKDIR /build/

RUN go build .


FROM alpine

RUN mkdir /app/
COPY --from=BUILD /build/imap-mailbox-exporter /app/
WORKDIR /app/

USER 65534
CMD ["./imap-mailbox-exporter", "-listen.address", "0.0.0.0:9117"]
EXPOSE 9117

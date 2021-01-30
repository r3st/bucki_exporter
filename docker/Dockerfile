FROM golang:1.15.6 as GO

ADD . /usr/local/buckisrc
WORKDIR /usr/local/buckisrc
RUN ls -alh
RUN go get -u all \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /tmp/bucki_exporter.go cmd/bucki/main.go

FROM scratch
COPY docker/passwd /etc/passwd
COPY docker/group /etc/group

COPY --from=GO /tmp/bucki_exporter.go /opt/bucki/
COPY --from=GO  /usr/local/buckisrc/configs /opt/bucki/configs
COPY --from=GO  /etc/ssl/certs  /etc/ssl/certs

WORKDIR /opt/bucki

ENTRYPOINT ["/opt/bucki/bucki_exporter.go"]
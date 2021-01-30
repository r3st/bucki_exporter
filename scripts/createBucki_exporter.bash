#! /bin/bash

rm -Rf /tmp/bucki_exporter

git clone https://github.com/r3st/bucki_exporter.git /tmp/bucki_exporter

cd /tmp/bucki_exporter

docker build -f docker/Dockerfile -t r3st/bucki_exporter:0.0.1 .
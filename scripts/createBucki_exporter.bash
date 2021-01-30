#! /bin/bash

rm -Rf /tmp/bucki_exporter

git clone https://github.com/r3st/bucki_exporter.git

cd /tmp/bucki_exporter

docker build -f docker/Dockerfile -t docker build -f docker/Dockerfile -t test .
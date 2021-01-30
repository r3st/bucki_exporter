# bucki_exporter

This helps to create a docker image for bucki_exporter. This exporter publishs metrics about health state from microprofile health.

## Build Image
Go must no be installed.

```bash
curl -s https://raw.githubusercontent.com/r3st/bucki_exporter/main/scripts/createBucki_exporter.bash | bash
```

## Run Image
Run docker Container:
```bash
➜  ~ docker run --name bucki-exporter --rm -p 9889:9889 r3st/bucki_exporter:0.0.1
```
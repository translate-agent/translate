# Translate-Agent

## All-in-one image
Running latest all in one image
```bash
docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -v /tmp/badger:/tmp/badger \
  expectdigital/translate-agent-all-in-one:latest
```
Remove existing, pull latest and run
```bash
docker rm -f translate-all-in-one 2> /dev/null; docker pull expectdigital/translate-agent-all-in-one; docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -v /tmp/badger:/tmp/badger \
  expectdigital/translate-agent-all-in-one
```

### All-in-one image docker run arguments description
| Argument                                | Description                                            |
|-----------------------------------------|--------------------------------------------------------|
| `-p 8080:8080`                          | Translate service port                                 |
| `-p 16686:16686`                        | Jaeger UI port                                         |
| `-v path/to/badger-dir:/tmp/badger`     | Path for BadgerDB db for data persistency *(Optional)* |
| `-v path/to/envoy.yaml:/app/envoy.yaml` | Path to custom envoy.yaml *(Optional)*                 |

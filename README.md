# Translate-Agent

## All-in-one image

### Setting up environment

```bash
# 3rd party Translator Service
# Empty string for no Translator Service
export TRANSLATE_SERVICE_TRANSLATOR=GoogleTranslate # "", GoogleTranslate or AWSTranslate

# Only when TRANSLATOR == GoogleTranslate
export TRANSLATE_OTHER_GOOGLE_PROJECT_ID= # Google project id
export TRANSLATE_OTHER_GOOGLE_LOCATION= # Google location e.g. global
export TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY= # Path to Google account key JSON file

# Only when TRANSLATOR == AWSTranslate
export TRANSLATE_OTHER_AWS_ACCESS_KEY_ID= # AWS access key id
export TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY= # AWS secret access key
export TRANSLATE_OTHER_AWS_REGION= # AWS region e.g. eu-west-2

# Optional

# Persist data (on Host) when deleting container.
# Named volume or bind mount.
export TRANSLATE_DB_HOST_BADGERDB_PATH=translate_badgerDB

# Provide custom envoy.yaml
# https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples
export TRANSLATE_ENVOY_CONFIG_PATH= # Path to envoy.yaml
```

### Running latest all in one image

```bash
docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -e TRANSLATE_SERVICE_TRANSLATOR \
  -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID \
  -e TRANSLATE_OTHER_GOOGLE_LOCATION \
  -v $TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY:/app/google_account_key.json \
  -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID \
  -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY \
  -e TRANSLATE_OTHER_AWS_REGION \
  expectdigital/translate-agent-all-in-one:latest

# Add to arguments if you want to persist data on host
# -v $TRANSLATE_DB_HOST_BADGERDB_PATH:/data/badgerdb \

# Add to arguments if you want to use custom envoy.yaml
# -v $TRANSLATE_ENVOY_CONFIG_PATH:/app/envoy.yaml \
```

### Updating all in one image

Remove the old container, pull the latest image and run it again.

```bash
docker rm -f translate-all-in-one 2> /dev/null
docker pull expectdigital/translate-agent-all-in-one
docker run -d --name translate-all-in-one \
  -p 8080:8080 \
  -p 16686:16686 \
  -e TRANSLATE_SERVICE_TRANSLATOR \
  -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID \
  -e TRANSLATE_OTHER_GOOGLE_LOCATION \
  -v $TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY:/app/google_account_key.json \
  -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID \
  -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY \
  -e TRANSLATE_OTHER_AWS_REGION \
  expectdigital/translate-agent-all-in-one:latest

# Add to arguments if you want to persist data on host
# -v $TRANSLATE_DB_HOST_BADGERDB_PATH:/data/badgerdb \

# Add to arguments if you want to use custom envoy.yaml
# -v $TRANSLATE_ENVOY_CONFIG_PATH:/app/envoy.yaml \
```

## TypeScript client

### Dependencies

```bash
npm config set @buf:registry  https://buf.build/gen/npm/v1/

npm install @buf/expectdigital_translate-agent.bufbuild_es@latest
npm install @buf/expectdigital_translate-agent.connectrpc_es@latest
npm install @connectrpc/connect
npm install @connectrpc/connect-web
```

### Usage

```typescript
import { createPromiseClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";

import { Service } from "@buf/expectdigital_translate-agent.bufbuild_es/translate/v1/translate_pb.js";
import { TranslateService } from "@buf/expectdigital_translate-agent.connectrpc_es/translate/v1/translate_connect.js";

const transport = createGrpcWebTransport({
  baseUrl: "http://localhost:8080",
});

const client = createPromiseClient(TranslateService, transport);

client.createService({
  service: new Service({ name: "test1" }),
});
```

## Development

The project uses [Earthly](https://earthly.dev) to automate all development tasks that can be run locally and in CI/CD environments.

```shell
✗ earthly doc
TARGETS:
  +init [--USERARCH] [--go_version=1.23.0]
      init sets up the project for local development.
  +up [--USERARCH] [--go_version=1.23.0]
      up installs the project to local docker instance.
  +down [--USERARCH] [--go_version=1.23.0]
      down uninstalls the project from local docker instance.
  +proto [--USERARCH] [--go_version=1.23.0] [--bufbuild_version=1.40.0]
      proto generates gRPC server.
  +buf-registry [--USERARCH] [--go_version=1.23.0] [--bufbuild_version=1.40.0]
      buf-registry pushes BUF modules to the registry.
  +migrate --db --db_user --db_host --db_port --db_schema [--USERARCH] [--go_version=1.23.0] [--migrate_version=4.17.0] [--cmd=up]
      migrate runs DDL migration scripts against the given database.
  +check [--USERARCH] [--go_version=1.23.0]
      check verifies code quality by running linters and tests.
  +lint-migrate [--USERARCH] [--go_version=1.23.0] [--sqlfluff_version=3.0.3]
      lint-migrate analyses migrate scripts for stylistic issues.
  +lint-go [--USERARCH] [--go_version=1.23.0] [--golangci_lint_version=1.60.3]
      lint-go analyses golang code for errors, bugs and stylistic issues (golangci-lint).
  +lint-proto [--USERARCH] [--go_version=1.23.0] [--bufbuild_version=1.40.0]
      lint-proto analyses proto for stylistic issues.
  +lint [--USERARCH] [--go_version=1.23.0]
      lint runs all linters for golang, proto and migrate scripts.
  +test-unit [--USERARCH] [--go_version=1.23.0]
      test-unit runs unit tests.
  +test-integration [--USERARCH] [--go_version=1.23.0] [--migrate_version=4.17.0]
      test-integration runs integration tests.
  +test [--USERARCH] [--go_version=1.23.0]
      test runs unit and integration tests.
  +build [--USERARCH] [--go_version=1.23.0] [--GOARCH=$USERARCH] [--GOOS=linux]
      build compiles translate service and client and saves them to ./bin.
  +image --registry [--USERARCH] [--go_version=1.23.0] [--TARGETARCH] [--tag=latest]
      image builds translate service image.
  +image-multiplatform --registry [--USERARCH] [--go_version=1.23.0]
      image-multiplatform builds translate service image for all platforms.
  +jaeger [--USERARCH] [--go_version=1.23.0]
      jaeger is helper target for all-in-one image, it removes the need
      to download the correct jaeger image on every build
  +image-all-in-one --registry [--USERARCH] [--go_version=1.23.0] [--TARGETARCH] [--tag=latest]
      image-all-in-one builds all-in-one image.
  +image-all-in-one-multiplatform --registry [--USERARCH] [--go_version=1.23.0] [--tag=latest]
      image-all-in-one-multiplatform builds all-in-one multiplatform images.
```

VERSION 0.8
PROJECT expect.digital/translate-agent

ARG --global USERARCH # Arch of the user running the build
ARG --global go_version=1.23.3

FROM --platform=linux/$USERARCH golang:$go_version-alpine

# Local dev flow

# init sets up the project for local development.
init:
  LOCALLY
  RUN earthly secret --org expect.digital --project translate-agent get google_account_key > google_account_key.json
  RUN \
    echo "# OpenTelemetry" > .env.test && \
    echo "OTEL_SERVICE_NAME=translate" >> .env.test && \
    echo "OTEL_EXPORTER_OTLP_INSECURE=true" >> .env.test && \
    echo "OTEL_RESOURCE_ATTRIBUTES=deployment.environment=test" >> .env.test && \
    echo "" >> .env.test && \
    echo "# Translate Service" >> .env.test && \
    echo "TRANSLATE_SERVICE_PORT=8080" >> .env.test && \
    echo "TRANSLATE_SERVICE_HOST=0.0.0.0" >> .env.test && \
    echo "TRANSLATE_SERVICE_DB=badgerdb" >> .env.test && \
    echo "" >> .env.test && \
    echo "# Translator" >> .env.test && \
    echo "TRANSLATE_SERVICE_TRANSLATOR=GoogleTranslate" >> .env.test && \
    echo "" >> .env.test && \
    echo "# MySQL" >> .env.test && \
    echo "TRANSLATE_DB_MYSQL_HOST=0.0.0.0" >> .env.test && \
    echo "TRANSLATE_DB_MYSQL_PORT=3306" >> .env.test && \
    echo "TRANSLATE_DB_MYSQL_USER=root" >> .env.test && \
    echo "TRANSLATE_DB_MYSQL_PASSWORD=" >> .env.test && \
    echo "TRANSLATE_DB_MYSQL_DATABASE=translate" >> .env.test && \
    echo "" >> .env.test && \
    echo "# BadgerDB" >> .env.test && \
    echo "TRANSLATE_DB_BADGERDB_PATH=" >> .env.test && \
    echo "" >> .env.test && \
    echo "# Google Translate API" >> .env.test && \
    echo "TRANSLATE_OTHER_GOOGLE_PROJECT_ID=expect-digital" >> .env.test && \
    echo "TRANSLATE_OTHER_GOOGLE_LOCATION=global" >> .env.test && \
    echo "TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY=$(pwd)/google_account_key.json" >> .env.test && \
    echo "" >> .env.test && \
    echo "# AWS Translate API" >> .env.test && \
    echo "TRANSLATE_OTHER_AWS_ACCESS_KEY_ID=$(earthly secret --org expect.digital --project translate-agent get aws_access_key_id)" >> .env.test && \
    echo "TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY=$(earthly secret --org expect.digital --project translate-agent get aws_secret_access_key)" >> .env.test && \
    echo "TRANSLATE_OTHER_AWS_REGION=eu-west-2" >> .env.test
  RUN \
    echo "db=mysql" > .arg && \
    echo "db_host=host.docker.internal" >> .arg && \
    echo "db_port=3306" >> .arg && \
    echo "db_user=root" >> .arg && \
    echo "db_schema=translate" >> .arg
  RUN echo "db_password=" >> .secret

# up installs the project to local docker instance.
up:
  LOCALLY
  RUN docker compose --project-name=translate --project-directory=.earthly up --detach --wait --timeout 60
  RUN docker exec mysql sh -c 'mysqladmin ping -h 127.0.0.1 -u root --wait=30 --silent'
  BUILD +migrate --db=mysql

# down uninstalls the project from local docker instance.
down:
  LOCALLY
  RUN docker compose --project-name=translate --project-directory=.earthly down -v --remove-orphans

# Others

deps:
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN --mount=type=cache,id=go-mod,target=/go/pkg/mod go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

go:
  FROM +deps
  WORKDIR /translate
  COPY --dir cmd pkg .
  COPY --platform=linux/$USERARCH +proto/translate/v1/* pkg/pb/translate/v1
  SAVE ARTIFACT /translate

# proto generates gRPC server.
proto:
  ARG bufbuild_version=1.41.0
  FROM bufbuild/buf:$bufbuild_version
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir proto .
  WORKDIR proto
  RUN --mount=type=cache,target=$BUF_CACHE_DIR,mode=0700 \
    buf dep update && buf build && buf generate
  RUN sed -i'.bak' '/client.UploadTranslationFile/i \
    \\tfile, _, err := req.FormFile("file")\n\
    \tif err != nil {\n\
    \t\t\treturn nil, metadata, status.Errorf(codes.InvalidArgument, "%s", "'file' is required")\n\
    \t}\n\
    \tdefer file.Close()\n\
    \n\
    \tprotoReq.Data, err = io.ReadAll(file)\n\
    \tif err != nil {\n\
    \t\t\treturn nil, metadata, status.Errorf(codes.Internal, "%v", err)\n\
    \t}\n\
    ' gen/proto/go/translate/v1/translate.pb.gw.go
  RUN rm gen/proto/go/translate/v1/translate.pb.gw.go.bak
  SAVE ARTIFACT . proto
  SAVE ARTIFACT gen/proto/go/translate/v1 translate/v1 AS LOCAL pkg/pb/translate/v1

# buf-registry pushes BUF modules to the registry.
buf-registry:
  ARG bufbuild_version=1.41.0
  FROM bufbuild/buf:$bufbuild_version
  WORKDIR proto
  COPY +proto/proto .
  RUN --secret BUF_USERNAME --secret BUF_API_TOKEN echo $BUF_API_TOKEN | buf registry login --username $BUF_USERNAME --token-stdin
  RUN --push buf push

# migrate runs DDL migration scripts against the given database.
migrate:
  ARG migrate_version=4.18.1
  FROM migrate/migrate:v$migrate_version
  ARG --required db # supported databases: mysql
  ARG --required db_user
  ARG --required db_host
  ARG --required db_port
  ARG --required db_schema
  ARG cmd=up
  WORKDIR /migrations
  COPY migrate/$db/* /migrations
  RUN --push --secret=db_password \
    if [[ $db = "mysql" ]]; then yes | migrate -path=. -database "$db://$db_user:$db_password@tcp($db_host:$db_port)/$db_schema" $cmd; fi

# check verifies code quality by running linters and tests.
check:
  BUILD +lint
  BUILD +test

# -----------------------Linting-----------------------

# lint-migrate analyses migrate scripts for stylistic issues.
lint-migrate:
  ARG sqlfluff_version=3.1.1
  FROM sqlfluff/sqlfluff:$sqlfluff_version
  WORKDIR migrate
  COPY migrate .sqlfluff .
  RUN sqlfluff lint

# lint-go analyses golang code for errors, bugs and stylistic issues (golangci-lint).
lint-go:
  ARG golangci_lint_version=1.61.0
  FROM golangci/golangci-lint:v$golangci_lint_version-alpine
  WORKDIR translate
  COPY +go/translate .
  COPY .golangci.yml .
  RUN \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci_lint \
    golangci-lint run --timeout 3m

# lint-proto analyses proto for stylistic issues.
lint-proto:
  ARG bufbuild_version=1.41.0
  FROM bufbuild/buf:$bufbuild_version
  WORKDIR proto
  COPY +proto/proto .
  RUN buf lint

# lint runs all linters for golang, proto and migrate scripts.
lint:
  BUILD +lint-go
  BUILD +lint-proto
  BUILD +lint-migrate

# -----------------------Testing-----------------------

# test-unit runs unit tests.
test-unit:
  FROM +go
  RUN \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    go test ./...

# test-integration runs integration tests.
test-integration:
  FROM earthly/dind:alpine-3.19
  ARG migrate_version=4.18.1
  COPY .earthly/compose.yaml compose.yaml
  COPY +go/translate /translate
  COPY --dir migrate/mysql migrate
  WITH DOCKER --compose compose.yaml --service mysql --pull migrate/migrate:v$migrate_version --pull golang:$go_version-alpine
    RUN \
      --secret=aws_access_key_id \
      --secret=aws_secret_access_key \
      --mount=type=secret,target=/translate/google_account_key.json,id=google_account_key \
      --mount=type=cache,id=go-mod,target=/go/pkg/mod \
      --mount=type=cache,id=go-build,target=/root/.cache/go-build \

      # Wait for the MySQL server to start running
      docker exec mysql sh -c 'mysqladmin ping -h 127.0.0.1 -u root --wait=30 --silent' && \

      # Run database migrations using the migrate/migrate image
      docker run -v /migrate:/migrations --network=host migrate/migrate:v$migrate_version  \
        -path /migrations \
        -database "mysql://root@tcp(127.0.0.1:3306)/translate" \
        up && \

      # Run integration tests
      docker run \
        --network=host \
        -v /go/pkg/mod:/go/pkg/mod \
        -v /root/.cache/go-build:/root/.cache/go-build \
        -v /translate:/translate \
        -e TRANSLATE_DB_MYSQL_HOST=127.0.0.1 \
        -e TRANSLATE_DB_MYSQL_PORT=3306 \
        -e TRANSLATE_DB_MYSQL_DATABASE=translate \
        -e TRANSLATE_DB_MYSQL_USER=root \
        -e TRANSLATE_OTHER_AWS_ACCESS_KEY_ID=$aws_access_key_id \
        -e TRANSLATE_OTHER_AWS_SECRET_ACCESS_KEY=$aws_secret_access_key \
        -e TRANSLATE_OTHER_AWS_REGION=eu-west-2 \
        -e TRANSLATE_OTHER_GOOGLE_PROJECT_ID=expect-digital \
        -e TRANSLATE_OTHER_GOOGLE_LOCATION=global \
        -e TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY=/translate/google_account_key.json \
        golang:$go_version-alpine go test -C /translate --tags=integration -count=1 ./...
  END

# test runs unit and integration tests.
test:
  BUILD +test-unit
  BUILD +test-integration

# -----------------------Building-----------------------

# build compiles translate service and client and saves them to ./bin.
build:
  ARG GOARCH=$USERARCH
  ARG GOOS=linux
  ENV CGO_ENABLED=0
  COPY --platform=linux/$USERARCH +go/translate translate
  WORKDIR translate
  RUN \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    go build -o translate-service cmd/translate/main.go && \
    go build -o translate cmd/client/main.go
  SAVE ARTIFACT translate-service bin/translate-service AS LOCAL bin/translate-service # service
  SAVE ARTIFACT translate bin/translate AS LOCAL bin/translate # client

# image builds translate service image.
image:
  ARG TARGETARCH
  ARG --required registry
  ARG tag=latest
  FROM alpine
  COPY --platform=linux/$USERARCH (+build/bin/translate-service --GOARCH=$TARGETARCH) /translate-service
  ENTRYPOINT ["/translate-service"]
  SAVE IMAGE --push $registry/translate:$tag

# image-multiplatform builds translate service image for all platforms.
image-multiplatform:
  ARG --required registry
  BUILD \
    --platform=linux/amd64 \
    --platform=linux/arm64 \
    +image --registry=$registry

# -----------------------All-in-one image-----------------------

# jaeger is helper target for all-in-one image, it removes the need
# to download the correct jaeger image on every build
jaeger:
  FROM jaegertracing/all-in-one:1.47
  SAVE ARTIFACT /go/bin/all-in-one-linux jaeger

# image-all-in-one builds all-in-one image.
image-all-in-one:
  ARG TARGETARCH
  ARG --required registry
  ARG tag=latest
  FROM envoyproxy/envoy:v1.26-latest

  # Install supervisor to run multiple processes in a single container
  RUN apt-get update && apt-get install -y supervisor

  WORKDIR app/

  # Copy supervisord configuration and envoy configuration
  COPY .earthly/supervisord.conf supervisord.conf
  COPY .earthly/envoy.yaml envoy.yaml

  # Copy binaries
  COPY +jaeger/jaeger /usr/local/bin/jaeger
  COPY --platform=linux/$USERARCH (+build/bin/translate-service --GOARCH=$TARGETARCH) /usr/local/bin/translate-service # service
  COPY --platform=linux/$USERARCH (+build/bin/translate --GOARCH=$TARGETARCH) /usr/local/bin/translate # client

  # Set required environment variables
  ENV TRANSLATE_DB_BADGERDB_PATH=/data/badgerdb
  ENV OTEL_SERVICE_NAME=translate
  ENV OTEL_EXPORTER_OTLP_INSECURE=true
  ENV TRANSLATE_OTHER_GOOGLE_ACCOUNT_KEY=/app/google_account_key.json

  ENTRYPOINT ["supervisord","-c","/app/supervisord.conf"]
  SAVE IMAGE --push $registry/translate-agent-all-in-one:$tag

# image-all-in-one-multiplatform builds all-in-one multiplatform images.
image-all-in-one-multiplatform:
  ARG --required registry
  ARG tag=latest
  BUILD \
    --platform=linux/amd64 \
    --platform=linux/arm64 \
    +image-all-in-one --registry=$registry --tag=$tag

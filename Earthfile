VERSION 0.7

ARG --global USERARCH

ARG --global go_version=1.20.4
ARG --global golangci_lint_version=1.53.2
ARG --global bufbuild_version=1.21.0
ARG --global migrate_version=4.16.1
ARG --global sqlfluff_version=2.1.1

FROM --platform=linux/$USERARCH golang:$go_version-alpine

deps:
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN --mount=type=cache,target=/go/pkg/mod go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

go:
  FROM +deps
  COPY --dir cmd pkg .
  COPY --platform=linux/$USERARCH +proto/translate/v1/* pkg/pb/translate/v1
  SAVE ARTIFACT /translate 

proto:
  FROM bufbuild/buf:$bufbuild_version
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir proto .
  WORKDIR proto
  RUN \
    --mount=type=cache,target=$BUF_CACHE_DIR \
      buf mod update && buf build && buf generate

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

# migrate runs DDL migration scripts against the given database.
migrate:
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

# -----------------------Linting-----------------------

lint-migrate:
  FROM sqlfluff/sqlfluff:$sqlfluff_version
  WORKDIR migrate
  COPY migrate .sqlfluff .
  RUN sqlfluff lint

lint-go:
  FROM golangci/golangci-lint:v$golangci_lint_version-alpine
  WORKDIR translate
  COPY +go/translate .
  COPY .golangci.yml .
  RUN --mount=type=cache,target=/root/.cache/golangci_lint golangci-lint run

lint-proto:
  FROM bufbuild/buf:$bufbuild_version
  WORKDIR proto
  COPY +proto/proto .
  RUN buf lint

lint:
  BUILD +lint-go
  BUILD +lint-proto
  BUILD +lint-migrate

# -----------------------Testing-----------------------

test-unit:
  FROM +go
  RUN go test ./... --count 1

test-integration:
  FROM earthly/dind:alpine
  COPY .earthly/compose.yaml compose.yaml
  COPY +go/translate /translate
  COPY --dir migrate/mysql migrate
  WITH DOCKER --compose compose.yaml --service mysql --pull migrate/migrate:v$migrate_version --pull golang:$go_version-alpine
    RUN \
      --mount=type=cache,target=/go/pkg/mod \
      --mount=type=cache,target=/root/.cache/go-build \

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
        golang:$go_version-alpine go test -C /translate --tags=integration -count=1 ./...
  END

test:
  BUILD +test-unit
  BUILD +test-integration

# -----------------------Building-----------------------

build:
  ARG GOARCH=$USERARCH
  ENV CGO_ENABLED=0 
  COPY --platform=linux/$USERARCH +go/translate translate
  WORKDIR translate
  RUN \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
    go build -o translate cmd/translate/main.go
  SAVE ARTIFACT translate bin/translate

image:
  ARG TARGETARCH
  ARG --required registry
  ARG tag=latest
  FROM alpine
  COPY --platform=linux/$USERARCH (+build/bin/translate --GOARCH=$TARGETARCH) /translate
  ENTRYPOINT ["/translate"]
  SAVE IMAGE --push $registry/translate:$tag

image-multiplatform:
  ARG --required registry
  BUILD \
  --platform=linux/amd64 \
  --platform=linux/arm64 \
  +image --registry=$registry \

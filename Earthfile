VERSION 0.7

ARG --global go_version=1.20.3
ARG --global golangci_lint_version=1.52.2
ARG --global bufbuild_version=1.17.0
ARG --global migrate_version=4.15.2
ARG --global sqlfluff_version=2.0.3

FROM golang:$go_version-alpine

deps:
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

go:
  FROM +deps
  COPY --dir cmd pkg .
  COPY +proto/translate/v1/* pkg/pb/translate/v1
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
  SAVE ARTIFACT gen/proto/go/translate/v1 translate/v1 AS LOCAL pkg/pb/translate/v1

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
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir proto .
  WORKDIR proto
  RUN \
    --mount=type=cache,target=$BUF_CACHE_DIR \
      buf mod update && buf lint

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
  # RUN --no-cache ls -lah /go
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

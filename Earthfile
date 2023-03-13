VERSION 0.7
ARG --global go_version=1.20.2
FROM golang:$go_version-alpine
ENV CGO_ENABLED=0

deps:
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

go:
  FROM +deps
  COPY --dir cmd pkg .
  COPY --dir +proto/translate/v1 pkg/server/translate/v1
  SAVE ARTIFACT /translate

init:
  LOCALLY
  RUN cp .earthly/.env .env

proto:
  ARG bufbuild_version=1.15.1
  FROM bufbuild/buf:$bufbuild_version
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir api/translate .
  WORKDIR translate
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf mod update
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf build
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf generate

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
  SAVE ARTIFACT gen/proto/go/translate/v1 translate/v1 AS LOCAL pkg/server/translate/v1

lint-go:
  FROM +deps
  ARG --global golangci_lint_version=1.51.2
  RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$golangci_lint_version
  COPY --dir cmd pkg .
  COPY --dir +proto/translate/v1 pkg/server/translate/v1
  COPY .golangci.yml .
  RUN golangci-lint run

lint-proto:
  ARG bufbuild_version=1.15.1
  FROM bufbuild/buf:$bufbuild_version
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir api/translate .
  WORKDIR translate
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf mod update
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf lint

lint:
  BUILD +lint-go
  BUILD +lint-proto

test:
  BUILD +test-unit
  BUILD +test-integration

test-unit:
  FROM +deps
  COPY --dir cmd pkg .
  COPY --dir +proto/translate/v1 pkg/server/translate/v1
  RUN go test ./...

test-integration:
  FROM earthly/dind:alpine
  COPY .earthly/compose.yaml compose.yaml
  COPY +go/translate /translate
  WITH DOCKER --compose compose.yaml
    RUN \
      --mount=type=cache,target=/go/pkg/mod \
      --mount=type=cache,target=/root/.cache/go-build \

      docker exec mysql sh -c 'while ! mysqladmin ping --silent; do sleep 1; done' && \

      docker run \
        -v /go/pkg/mod:/go/pkg/mod \
        -v /root/.cache/go-build:/root/.cache/go-build \
        -v /translate:/translate \
        -e TRANSLATE_MYSQL_ADDR=translate:3306 \
        -e TRANSLATE_MYSQL_USER=root \
        golang:$go_version go test -C /translate --tags=integration -count=1 ./...
  END

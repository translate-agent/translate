VERSION 0.6
ARG go_version=1.20
ARG golangci_lint_version=1.51.1

deps:
  FROM golang:$go_version-alpine
  ENV CGO_ENABLED=0
  RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$golangci_lint_version
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

proto:
  FROM bufbuild/buf
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir api/translate .
  WORKDIR translate
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf mod update
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf build
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf generate
  
  RUN sed -i'.bu' '/client.UploadTranslationFile/i \
  \\tfile, _, err := req.FormFile("file")\n\
	\t\tif err != nil {\n\
		\t\t\treturn nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)\n\
	\t\t}\n\
	\n\
	\tdata, err := io.ReadAll(file)\n\
	\t\tif err != nil {\n\
		\t\t\treturn nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)\n\
	\t\t}\n\
	\n\
	\tprotoReq.Data = data\n' gen/proto/go/translate/v1/translate.pb.gw.go

  RUN rm gen/proto/go/translate/v1/translate.pb.gw.go.bu
  SAVE ARTIFACT gen/proto/go/translate/v1 translate/v1 AS LOCAL pkg/server/translate/v1

lint-go:
  FROM +deps
  COPY --dir cmd pkg .
  COPY --dir +proto/translate/v1 pkg/server/translate/v1
  COPY .golangci.yml .
  RUN golangci-lint run

lint-proto:
  FROM bufbuild/buf
  ENV BUF_CACHE_DIR=/.cache/buf_cache
  COPY --dir api/translate .
  WORKDIR translate
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf mod update
  RUN --mount=type=cache,target=$BUF_CACHE_DIR buf lint

lint:
  BUILD +lint-go
  BUILD +lint-proto

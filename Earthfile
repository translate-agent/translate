VERSION 0.7
ARG --global go_version=1.20.2
ARG --global golangci_lint_version=1.51.2
ARG --global bufbuild_version=1.15.1
FROM golang:$go_version-alpine


deps:
  ENV CGO_ENABLED=0
  RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$golangci_lint_version
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

proto:
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
  #TODO Save to more apropriate dir e.g. pkg/pb/translate/v1 or pkg/gen/translate/v1
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

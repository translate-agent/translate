VERSION 0.6
ARG go_version=1.20
ARG golangci_lint_version=1.51.0

deps:
  FROM golang:$go_version-alpine
  ENV CGO_ENABLED=0
  RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$golangci_lint_version
  WORKDIR /translate
  COPY go.mod go.sum .
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

lint-go:
  FROM +deps
  COPY --dir cmd .
  COPY .golangci.yml .
  RUN golangci-lint run

lint:
  BUILD +lint-go

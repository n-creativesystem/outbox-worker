FROM golang:1.25.5-trixie as builder

ARG TARGETPLATFORM
ARG VERSION=main

ENV GO111MODULE=on \
  GOPATH=/go \
  GOBIN=/go/bin

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  make build VERSION=${VERSION} \
  && chmod +x /workspace/outbox-worker

FROM gcr.io/distroless/static:nonroot
COPY --from=builder --chown=nonroot:nonroot /workspace/outbox-worker /usr/local/bin/outbox-worker
# current directory is `/home/nonroot`
COPY --chown=nonroot:nonroot config.yaml config.yaml
ENV TZ=Asia/Tokyo
USER 65532:65532

CMD ["outbox-worker"]

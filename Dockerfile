# Builder
FROM --platform=$BUILDPLATFORM golang AS builder
WORKDIR /builder
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -tags=nosqlite -o /builder/torrenttp

# Deploy
FROM gcr.io/distroless/base
COPY --from=builder /builder/torrenttp /torrenttp
EXPOSE 1010
ENTRYPOINT ["/torrenttp", "-dir", "/dl"]
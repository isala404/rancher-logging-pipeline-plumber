FROM node:alpine as uibuild
RUN apk add --no-cache yarn

WORKDIR /workspace

COPY ui/package.json .
RUN yarn install --no-lockfile --silent --cache-folder .yc

COPY ui/ .
RUN yarn build

# Build the manager binary
FROM golang:1.16-alpine as controller-builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/
COPY controllers/ controllers/

# Copy UI build
COPY --from=uibuild /workspace/build/ pkg/webserver/build/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM golang:1.16-alpine as pod-simulator-builder
WORKDIR /build
COPY pod-simulator /build
RUN go build -o pod-simulator main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=pod-simulator-builder /build/pod-simulator /usr/local/bin/pod-simulator
COPY --from=controller-builder /workspace/manager /usr/local/bin/controller
USER 65532:65532

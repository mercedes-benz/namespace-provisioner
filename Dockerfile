# SPDX-License-Identifier: MIT

# Conventions:
# WORKDIR = /workdir
# Build and test results should be in /workdir/build

###############################################################################
# SET UP BUILD-ENV
###############################################################################
FROM golang:1.13 as build-env

ARG TASK_VERSION=2.7.1

# Install Task
WORKDIR /tmp
RUN curl -sLSfo task.tgz https://github.com/go-task/task/releases/download/v${TASK_VERSION}/task_linux_amd64.tar.gz && \
    mkdir -p task && \
    tar xvf task.tgz -C task && \
    mv task/task /usr/local/bin/ && \
    rm -rf task*

WORKDIR /workdir
COPY go.mod go.mod
COPY go.sum go.sum
COPY ./tasks/BuildTasks.yml Taskfile.yml
RUN task prepare

###############################################################################
# BUILD
###############################################################################
FROM build-env as build

ARG GO_BUILD_ENV="GOOS=linux GOARCH=amd64 CGO_ENABLED=0"

COPY pkg/controllers pkg/controllers/
COPY pkg/util pkg/util/
COPY main.go main.go
RUN task build GO_BUILD_ENV="${GO_BUILD_ENV}"

###############################################################################
# TEST
###############################################################################
FROM build as test
ARG BUILD_DATE
RUN task test

###############################################################################
# FINAL IMAGE
###############################################################################
FROM alpine:latest

ARG BUILD_DATE
ARG VCS_REF
ARG BUILD_VERSION

LABEL com.daimler.namespace-provisioner.license="MIT" \
      com.daimler.namespace-provisioner.license-url="https://github.com/Daimler/namespace-provisioner/blob/master/LICENSE" \
      org.opencontainers.image.authors="Daimler TSS GmbH" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.description="A Kubernetes operator creating k8s resources by annotating namespaces." \
      org.opencontainers.image.documentation="https://github.com/Daimler/namespace-provisioner/blob/master/README.md" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/Daimler/namespace-provisioner" \
      org.opencontainers.image.title="Namespace Provisioner" \
      org.opencontainers.image.url="https://github.com/Daimler/namespace-provisioner" \
      org.opencontainers.image.vendor="Daimler TSS GmbH" \
      org.opencontainers.image.version="${BUILD_VERSION}"

USER 100

WORKDIR /app
COPY --chown=100:100 --from=build /workdir/build/bin/namespace-provisioner .
COPY LICENSE /LICENSE

ENTRYPOINT ["./namespace-provisioner"]

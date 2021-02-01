ARG DOCKER_BUILDER_SERVER_IMAGE=golang:1.14
ARG DOCKER_BASE_IMAGE=alpine:3.12

# Build Server
FROM ${DOCKER_BUILDER_SERVER_IMAGE} AS builder-server
ARG GITHUB_USERNAME
ARG GITHUB_TOKEN

WORKDIR /server

# Download dependencies
RUN echo "machine github.com login ${GITHUB_USERNAME} password ${GITHUB_TOKEN}" > ~/.netrc
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Production environment
FROM ${DOCKER_BASE_IMAGE}

ENV PILLAR=/pillar/pillar \
  USER_UID=10001 \
  USER_NAME=pillar

WORKDIR /pillar

COPY --from=builder-server /server/build/_output/bin/pillar /pillar/
COPY --from=builder-server /server/build/bin /usr/local/bin

RUN  /usr/local/bin/user_setup

EXPOSE 8078

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

ARG DOCKER_BUILDER_SERVER_IMAGE=golang:1.14
ARG DOCKER_BUILDER_WEBAPP_IMAGE=node:12.16.1-alpine
ARG DOCKER_BASE_IMAGE=alpine:3.12
# Build Webapp
FROM ${DOCKER_BUILDER_WEBAPP_IMAGE} as builder-webapp

WORKDIR /webapp

COPY webapp/package.json .
COPY webapp/package-lock.json .

RUN npm install --production
COPY webapp/ .

ARG ENV
RUN if [ "$ENV" = "cloud" ] ; then npm run build:cloud ; else npm run build; fi

# Build Server
FROM ${DOCKER_BUILDER_SERVER_IMAGE} AS builder-server
ARG GITHUB_USERNAME
ARG GITHUB_TOKEN

WORKDIR /server

# Download dependencies
RUN echo "machine github.com login ${GITHUB_USERNAME} password ${GITHUB_TOKEN}" > ~/.netrc
COPY go.mod go.sum ./
RUN GOPRIVATE=github.com/mattermost/license-generator go mod download && go mod verify

COPY . .
RUN GOPRIVATE=github.com/mattermost/license-generator make build

# Used to generate invoice PDFs
FROM surnet/alpine-wkhtmltopdf:3.12-0.12.6-full AS wkhtmltopdf

# Production environment
FROM ${DOCKER_BASE_IMAGE}

# Install dependencies for wkhtmltopdf
RUN apk add --no-cache \
  libstdc++ \
  libx11 \
  libxrender \
  libxext \
  libssl1.1 \
  ca-certificates \
  fontconfig \
  freetype \
  ttf-dejavu \
  ttf-droid \
  ttf-freefont \
  ttf-liberation \
  ttf-ubuntu-font-family \
&& apk add --no-cache --virtual .build-deps \
  msttcorefonts-installer \
\
# Install microsoft fonts
&& update-ms-fonts \
&& fc-cache -f \
\
# Clean up when done
&& rm -rf /tmp/* \
&& apk del .build-deps

ENV CWS=/cws/cws \
  USER_UID=10001 \
  USER_NAME=cws

WORKDIR /cws

COPY --from=builder-webapp /webapp/build/ /cws/webapp/build/
COPY --from=builder-server /server/build/_output/bin/cws /cws/
COPY --from=builder-server /server/i18n /cws/i18n
COPY --from=builder-server /server/internal/templates /cws/internal/templates
COPY --from=builder-server /server/build/bin /usr/local/bin

# Copy wkhtmltopdf files from docker-wkhtmltopdf image
COPY --from=wkhtmltopdf /bin/wkhtmltopdf /bin/wkhtmltopdf
COPY --from=wkhtmltopdf /bin/wkhtmltoimage /bin/wkhtmltoimage
COPY --from=wkhtmltopdf /bin/libwkhtmltox* /bin/
RUN  /usr/local/bin/user_setup

EXPOSE 8076 8077

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

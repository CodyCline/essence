ARG ALPINE_BUILDER_VARIANT=1.22-alpine
ARG ALPINE_RELEASER_VARIANT=3.18.4


FROM golang:${ALPINE_BUILDER_VARIANT} as builder
ARG ESSENCE_VERSION="v0.0.0+unknown+docker"
ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk add build-base \
    && apk add --no-cache git 
WORKDIR /app

COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./
RUN go build -v -ldflags="-s -w -X main.Version=$ESSENCE_VERSION" -o bin/essence cmd/essence/main.go

FROM alpine:${ALPINE_RELEASER_VARIANT}
RUN apk -U upgrade --no-cache 
COPY --from=builder /app/bin/essence /usr/local/bin/essence
ENTRYPOINT [ "essence" ] 

LABEL IMAGE_NAME="essence"
LABEL AUTHOR="10ad3d"
LABEL DESCRIPTION="essence returns a list of unique domains or subdomains from a list of strings, URLs or emails"
#TODO Inject from OS environment
LABEL IMAGE_VERSION=ESSENCE_VERSION
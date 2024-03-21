package templates

const DockerLocalAPI = `FROM alpine:latest
RUN apk update && apk add --no-cache bash
WORKDIR /api
ADD ./tmp .
ENTRYPOINT ./{{.AppName}}-local`

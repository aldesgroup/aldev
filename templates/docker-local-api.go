package templates

const DockerLocalAPI = `FROM alpine:latest
RUN apk update && apk add --no-cache bash
WORKDIR /api
COPY ./{{.GetResolvedBinDir}} .
COPY ./{{.API.DataDir}} .
ENTRYPOINT ./{{.AppName}}-api-local`

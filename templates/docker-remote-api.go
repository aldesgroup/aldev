package templates

const DockerRemoteAPI = `# --- backend building --------------------------------------------------------
# Use an official Golang runtime as a parent image
FROM golang:alpine AS buildStage

# Adding Git
RUN apk update && apk add --no-cache git

# Set the GitHub token as a build argument
ARG GITHUB_TOKEN

# This is needed to use private modules
ENV GOPRIVATE=github.com/aldesgroup/goald
RUN git config --global --add url."https://${GITHUB_TOKEN}:@github.com/aldesgroup".insteadOf "https://github.com/aldesgroup"

# Set the working directory in the container
WORKDIR /build

# Copy the local package files to the container's workspace
COPY {{.API.SrcDir}} .

# Build the backend application inside the container
RUN go build -o {{.AppName}}-api ./main

# --- running -----------------------------------------------------------------
# Use a smaller base image for the final image
FROM alpine:latest

# Set the working directory in the container
WORKDIR /api

# Copy the binary from the build stage
COPY --from=buildStage /build/{{.AppName}}-api .

# Command to run the executable
ENTRYPOINT ["./{{.AppName}}-api"]`

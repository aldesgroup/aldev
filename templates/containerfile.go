package templates

const ContainerFILE = `# 2-stages = 900+ mo (golang image + our sources) -> ~145 mo final image

# Stage 1: Build the binary
FROM quay.io/projectquay/golang:%s AS builder

# Set the context to the folder containing go.mod
WORKDIR /container-api-src

# Copy the dependency files from the host's api folder (optimization)
COPY {{.API.SrcDir}}/go.mod {{.API.SrcDir}}/go.sum ./
RUN go mod download

# Copy the rest of the local api folder to the container's current WORKDIR
COPY {{.API.SrcDir}}/ .

# Build the binary
# We use '.' because we are already inside /container-api where the code lives
# -s: Omit Symbol Table, harder reverse-engineering
# -w: Omit DWARF, Removes DWARF debugging information.
# CGO_ENABLED=0: no C code, pure Go, static binary starting a bit faster
# GOOS=linux: making sure we're builing for a Linux env
RUN CGO_ENABLED=0 GOOS=linux GOAMD64=v3 go build -ldflags="-s -w" -o /bin/{{.AppNameKebab}}-api ./main

# Stage 2: Final Runtime
FROM quay.io/fedora/fedora-minimal:latest

WORKDIR /container-api-bin

# Copy the binary from the builder stage
COPY --from=builder /bin/{{.AppNameKebab}}-api bin/

# Copy your necessary config/data from the host
# Note: These paths are relative to where you run 'podman build'
COPY {{.API.SrcDir}}/{{.API.Config}} apiconf.yaml
COPY {{.API.DataDir}}/ ./{{.API.DataDir}}/

ENTRYPOINT ["bin/{{.AppNameKebab}}-api", "-config", "apiconf.yaml"]`

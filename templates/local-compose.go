package templates

const LocalCOMPOSE = `services:
  # Defines the backend Go application service
  {{.AppNameShort}}_api:
    # Uses a lightweight, glibc-compatible base to ensure host-compiled Go binaries run smoothly
    image: quay.io/fedora/fedora-minimal:latest

    # Sets the starting directory inside the container for any relative paths in your code
    working_dir: /api

    volumes:
      # Mounts the host's bin folder to the container.
      # The ':z' tells Podman to relabel the files for SELinux (essential on Fedora/RHEL)
      - ../../bin:/api/bin:z
      - ../../api:/api/api:z
      - ../../data:/api/data:z

    # Executes the binary. Running it directly (not via a shell script) helps with signal handling (SIGTERM)
    command: ./bin/{{.AppNameKebab}}-api -config api/apiconf.yaml

  # Defines the Load Balancer
  nginx:
    # Nginx micro-image maintained by the Fedora/Red Hat community on Quay
    # This is a hardened, small-footprint version of Nginx
    image: quay.io/nginx/nginx-unprivileged:latest

    ports:
      # Maps your laptop's port 8080 to the Nginx port 8080
      - "{{.API.Port}}:8080"

    volumes:
      # Mounts your custom load-balancing config as Read-Only (:ro)
      - ./nginx.conf:/etc/nginx/nginx.conf:ro

    depends_on:
      # Ensures the 'go_api' containers are created before Nginx starts looking for them
      - {{.AppNameShort}}_api`

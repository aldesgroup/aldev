package templates

const LocalNGINX = `# Moves the PID file to /tmp because the unprivileged user cannot write to the default /run/ directory.
pid /tmp/nginx.pid;

# Standard block to define worker process settings.
events {
    # Maximum number of simultaneous connections that can be opened by a worker process.
    worker_connections 1024;
}

http {
    # We don't need the NGINX access logs, it's a bit too spammy
    access_log off;

    # Defines a group of servers (load balancing) that can be referenced by the proxy_pass directive.
    upstream {{.AppNameShort}}_servers {
        # '{{.AppNameShort}}_api' is the service name defined in the compose.yaml.
        # Podman-compose creates a private network where '{{.AppNameShort}}_api' resolves to the internal IPs
        # of your Go containers on port {{.LocalPort}}.
        server {{.AppNameShort}}_api:{{.LocalPort}};
    }

    server {
        # Port the Nginx container listens on.
        # MUST be 8080 or higher because unprivileged users cannot bind to ports below 1024.
        listen 8080;

        # Configuration for all requests coming to the root path.
        location / {
            # Forwards the request to the upstream group defined above.
            proxy_pass http://{{.AppNameShort}}_servers;

            # Passes the original Host header from the client to the Go server.
            proxy_set_header Host $host;

            # Passes the real IP address of the client to the Go server (useful for logging/auth).
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}`

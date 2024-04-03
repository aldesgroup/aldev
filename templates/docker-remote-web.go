package templates

const DockerRemoteWeb = `# Use Node.js image as a base
FROM node:alpine AS builder

# Set the working directory in the container
WORKDIR /web

# Copy package.json and package-lock.json files
COPY {{.Web.Dir}}/package*.json ./

# Install dependencies
RUN npm ci

# Copy the rest of the application code
COPY {{.Web.Dir}}/. .

# Build the React app
RUN npm run build

# Use Nginx as a lightweight web server
FROM nginx:alpine

# Copy the built React app from the previous stage
COPY --from=builder /web/build /usr/share/nginx/html

# Expose port 80 to the outside world
EXPOSE 80

# Start Nginx server
CMD ["nginx", "-g", "daemon off;"]
`
package templates

const DockerLocalWEB = `FROM node:alpine
WORKDIR /web
COPY {{.Web.Dir}}/package*.json {{.Web.Dir}}/.npmrc ./
RUN npm install
COPY {{.Web.Dir}}/ ./
CMD ["npm", "run", "dev"]`

package templates

const DockerLocalWEB = `FROM node:alpine
WORKDIR /web
COPY {{.Web.SrcDir}}/package*.json {{.Web.SrcDir}}/.npmrc ./
RUN npm install --force
COPY {{.Web.SrcDir}}/ ./
CMD ["npm", "run", "dev"]`

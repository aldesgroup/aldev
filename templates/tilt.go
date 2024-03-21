package templates

const Tiltfile = `# not sending data
analytics_settings(enable=False)

# --- Preparation -------------------------------------------------------------

# custom config
config.define_string_list('use-local', usage='use this to include our own libraries in dev mode (like goaldr)')
cfg = config.parse()
localDeps = cfg.get('use-local')

# describing the deployment of all the services, and their configuration
k8s_yaml(['{{.Deploying.Dir}}/{{.AppName}}-app.yaml', '{{.Deploying.Dir}}/{{.AppName}}-cm.yaml'])

# --- API part ----------------------------------------------------------------

# building the API's code
local_resource(
    name  ='{{.AppName}}-back-compile',
    cmd   ='cd {{.API.Dir}} && go mod tidy && go build -o ../tmp/{{.AppName}}-local ./main && cd ..',
    deps  =['{{.API.Dir}}', '../goald'], # taking into account the dependencies
    ignore=['{{.API.Dir}}/go.sum', '{{.API.Config}}'],
    )

# describing the containers for the backend - cf https://docs.tilt.dev/extensions.html
load('ext://restart_process', 'docker_build_with_restart')
docker_build_with_restart(
  ref        ='{{.AppName}}-local-image',
  context    ='.',
  entrypoint =['/api/{{.AppName}}-local'],
  dockerfile ='{{.Deploying.Dir}}/{{.AppName}}-docker-local-api',
  only       =['./tmp'],
  live_update=[
    sync('./tmp', '/api'),
  ],
)

# deploying the API
k8s_resource('{{.AppName}}-back', resource_deps=['{{.AppName}}-back-compile'])

# getting the load balancer's IP & port - cf https://docs.tilt.dev/extensions.html
apiHost = str(local(echo_off=True, command="kubectl get services --namespace kube-system "+
  "-o jsonpath='{.items[?(@.spec.type==\"LoadBalancer\")].status.loadBalancer.ingress[0].ip}'"))

# --- WEB part ----------------------------------------------------------------

# the command to make sure we have the right set of dependencies
webDepsCmd = 'npm install'

# if we're developing libraries along with this project, then:
if localDeps:
  # we need to link these libraries - making sure Vite's cache is refreshed
  webDepsCmd = 'rm -fr node_modules/.vite && npm link ' + ' '.join(localDeps)
  # dealing with each local library
  for localDep in localDeps:
    # first, let's get its name
    localDepName = localDep
    if "/" in localDep:
      localDepName = localDep[localDep.index("/")+1:]
    # we assume all the Git projects to be in the same folder for now
    localDepDir = '../' + localDepName
    warn("using the local dependency '"+localDep+"' from '"+localDepDir)
    # now making sure we're refreshing it each time it changes
    local_resource(
        name         ='refresh-'+localDepName,
        serve_dir    =localDepDir,
        serve_cmd    ='npx tsc && npx vite build --watch',
        # if running TSC is not needed at each change, else use what's commented below instead
        # dir          =localDepDir,
        # cmd          ='npx tsc && npx vite build',
        # deps         =[localDepDir+'/package.json', localDepDir+'/lib'],
        )

# locally running Vite's dev server - no need to containerize this for now
local_resource(
    name         ='{{.AppName}}-front-serve',
    dir          ='{{.Web.Dir}}',
    cmd          =webDepsCmd,
    deps         =['{{.Web.Dir}}/package.json'],
    serve_cmd    ='LOCAL_API_URL=http://'+apiHost+':{{.API.Port}} npm run dev',
    # serve_cmd    ='LOCAL_API_URL=http://'+apiHost+':'+apiPort+' npm run dev',
    serve_dir    ='{{.Web.Dir}}',
    resource_deps=['{{.AppName}}-back'],
    )
`

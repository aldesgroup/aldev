package templates

const Tiltfile = `# not sending data
analytics_settings(enable=False)

# --- Preparation -------------------------------------------------------------

# custom config
config.define_string_list('use-local', usage='use this to include our own libraries in dev mode (like goaldr)')
cfg = config.parse()
localDeps = cfg.get('use-local')

# describing the deployment of all the backend services, and their configuration
if localDeps:
  k8s_yaml(kustomize('{{.Deploying.Dir}}/overlays/local'))
else:
  # when working with no use of local deps, the base config is enough
  k8s_yaml(kustomize('{{.Deploying.Dir}}/overlays/dev'))

# --- API part ----------------------------------------------------------------

# building the API's code
local_resource(
    name  ='{{.AppName}}-api-compile',
    cmd   ='aldev build',
    deps  =['{{.API.Dir}}', '../goald', '{{.API.I18n.File}}'], # taking into account the dependencies
    ignore=['{{.API.Dir}}/go.sum', '{{.API.Dir}}/_generated', '{{.API.Config}}'], # the API config is ignored here, but Aldev watches it
    )

# describing the containers for the backend - cf https://docs.tilt.dev/extensions.html
load('ext://restart_process', 'docker_build_with_restart')
docker_build_with_restart(
  ref        ='{{.AppName}}-api-image',
  context    ='.',
  entrypoint =['/api/{{.AppName}}-api-local'],
  dockerfile ='{{.Deploying.Dir}}/docker/{{.AppName}}-local-api-docker',
  only       =['./{{.Deploying.Tmp}}'],
  live_update=[
    sync('./{{.Deploying.Tmp}}', '/api'),
  ],
)

# deploying the API
k8s_resource('{{.AppName}}-api-depl', resource_deps=['{{.AppName}}-api-compile'])

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
  # our Vite server will depend on the watching of these libraries
  resource_deps=['{{.AppName}}-api-depl']
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
    # making Vite waiting for thid
    resource_deps.append('refresh-'+localDepName)

  # locally running Vite's dev server - no need to containerize this for now
  local_resource(
      name         ='{{.AppName}}-vite-serve',
      dir          ='{{.Web.Dir}}',
      cmd          =webDepsCmd,
      deps         =['{{.Web.Dir}}/package.json'],
      serve_cmd    ='LOCAL_API_URL=http://'+apiHost+':{{.API.Port}} npm run dev',
      # serve_cmd    ='LOCAL_API_URL=http://'+apiHost+':'+apiPort+' npm run dev',
      serve_dir    ='{{.Web.Dir}}',
      resource_deps=resource_deps,
      )
else:
  docker_build(
    '{{.AppName}}-web-image',
    context='.',
    dockerfile='./{{.Deploying.Dir}}/docker/{{.AppName}}-local-web-docker',
    only=['{{.Web.Dir}}/'],
    ignore=['{{.Web.Dir}}/dist/'],
    live_update=[
        fall_back_on('{{.Web.Dir}}/vite.config.js'),
        sync('{{.Web.Dir}}/', '/web/'),
        run(
            'npm install',
            trigger=['{{.Web.Dir}}/package.json', '{{.Web.Dir}}/package-json.lock']
        )
    ]
  )

  k8s_resource(
      '{{.AppName}}-web-depl',
      port_forwards='{{.Web.Port}}:5173', # 5173 is the port Vite listens on in the container
      resource_deps=['{{.AppName}}-api-depl'],
      # labels=['frontend']
  )
` // end of template

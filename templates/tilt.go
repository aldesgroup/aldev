package templates

const Tiltfile = `# Generated by aldev, do not edit!
# not sending data
analytics_settings(enable=False)

# --- Preparation -------------------------------------------------------------

# custom config
config.define_bool('use-local', usage='use this to include our own libraries in dev mode (like goaldr)')
config.define_bool('api-only', usage='use this to only work with the API part of this project')
cfg = config.parse()
useLocalDeps = cfg.get('use-local')
apiOnly = cfg.get('api-only')

# describing the deployment of all the backend services, and their configuration
if useLocalDeps or apiOnly:
  # when working with local dependencies, or in API-only mode, we do not dockerise the web app
  k8s_yaml(kustomize('{{.Deploying.Dir}}/overlays/local'))
else:
  # when working with no use of local deps, the base config is enough
  k8s_yaml(kustomize('{{.Deploying.Dir}}/overlays/dev'))

# --- API part ----------------------------------------------------------------

# building the API's code
local_resource(
    name  ='{{.AppName}}-api-compile',
    cmd   ='aldev complete',
    # taking into account the dependencies
    deps  =['{{.API.SrcDir}}', '../goald', '../emerald'], # REMOVE EMERALD
    # the API config is also ignored here, because Aldev is already watching it
    ignore=['{{.API.SrcDir}}/go.sum', '{{.API.SrcDir}}/_include', '{{.API.SrcDir}}/**/*/classutils', '{{.API.Config}}'],
    )

# describing the containers for the backend - cf https://docs.tilt.dev/extensions.html
load('ext://restart_process', 'docker_build_with_restart')
docker_build_with_restart(
  ref        ='{{.AppName}}-api-image',
  context    ='.',
  entrypoint =['/api/{{.AppName}}-api-local'],
  dockerfile ='{{.Deploying.Dir}}/docker/{{.AppName}}-local-api-docker',
  only       =['./{{.GetResolvedBinDir}}', './{{.API.DataDir}}'],
  live_update=[
    sync('./{{.GetResolvedBinDir}}', '/api'),
    sync('./{{.API.DataDir}}', '/api/data'),
  ],
)

# deploying the API
k8s_resource('{{.AppName}}-api-depl', resource_deps=['{{.AppName}}-api-compile'])

# getting the load balancer's IP - cf https://docs.tilt.dev/extensions.html
apiHost = str(local(echo_off=True, command="kubectl get services --namespace kube-system "+
  "-o jsonpath='{.items[?(@.spec.type==\"LoadBalancer\")].status.loadBalancer.ingress[0].ip}'"))

# since the load balancer is on host 'apiHost' and not 'localhost', we can reach it by sending calls
# to localhost:apiPort - from Windows for instance - and proxying them inside Linux to apiHost:apiPort
local_resource(
    name         ='{{.AppName}}-lb-proxy',
    cmd          ='killall -q socat || true',
    serve_cmd    ='socat TCP-LISTEN:{{.API.Port}},fork TCP:'+apiHost+':{{.API.Port}}',
    resource_deps=['{{.AppName}}-api-depl'],
    )

# --- WEB part ----------------------------------------------------------------

if not apiOnly:
  # if we're developing libraries along with this project, then:
  if useLocalDeps:
    # these are the env vars defined in the aldev config file (.aldev.yaml)
    webAppEnvVars = "WEB_API_URL=http://localhost:{{.API.Port}}"
    {{range .Web.EnvVars}}webAppEnvVars += " {{.Name}}={{.Value}}"
    {{end}}
    # locally running Vite's dev server - no need to containerize this for now
    local_resource(
        name         ='{{.AppName}}-vite-serve',
        dir          ='{{.Web.SrcDir}}',
        cmd          ='rm -fr node_modules/.vite && npm i --force',
        deps         =['{{.Web.SrcDir}}/package.json'],
        serve_cmd    =webAppEnvVars + ' npm run dev',
        serve_dir    ='{{.Web.SrcDir}}',
        resource_deps=['{{.AppName}}-lb-proxy'],
        )
  else:
    docker_build(
      '{{.AppName}}-web-image',
      context='.',
      dockerfile='./{{.Deploying.Dir}}/docker/{{.AppName}}-local-web-docker',
      only=['{{.Web.SrcDir}}/'],
      ignore=['{{.Web.SrcDir}}/dist/'],
      live_update=[
          fall_back_on('{{.Web.SrcDir}}/vite.config.js'),
          sync('{{.Web.SrcDir}}/', '/web/'),
          run(
              'npm install --force',
              trigger=['{{.Web.SrcDir}}/package.json', '{{.Web.SrcDir}}/package-json.lock']
          )
      ]
    )

    k8s_resource(
        '{{.AppName}}-web-depl',
        port_forwards='{{.Web.Port}}:5173', # 5173 is the port Vite listens on in the container
        resource_deps=['{{.AppName}}-api-depl'],
        # labels=['frontend']
    )
else:
  print("\nAPI only mode")
` // end of template

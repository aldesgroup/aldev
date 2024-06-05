package templates

const WebEnvList = `WEB_API_URL=URL for the API; automatically set in the local dev environment in the Tiltfile
{{range .Web.EnvVars}}{{.Name}}={{.Desc}}
{{end}}
`

// const WebEnvList = `
// `

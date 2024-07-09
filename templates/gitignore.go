package templates

const GitIgnore = `# no hidden file
.*

# except these
!.dockerignore
!.gitignore
!.aldev.yaml
!.gitkeep
!.prettierrc
!.eslintrc.cjs
!.gitlab-ci.yml

# no temp file
bin/
tmp/
node_modules/
`

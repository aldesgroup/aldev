# Aldev
CLI tool to help quick dev with the Goald / GoaldR stack

TODO: more doc.

## Help

```sh
$ aldev -h
Run `aldev` to start or continue developing a Goald / GoaldR or GoaldN application, with automatic deployment in a local k8s cluster and live reloading when applicable. Or use one of the available command to perform a specific action.

Usage:
  aldev [flags]
  aldev [command]

Available Commands:
  bootstrap   Bootstraps a new aldev project
  codegen     Completes the app with additional generated code to speed up your dev
  codeswap    Swaps bit of code - like import paths - in targeted files
  completion  Generate the autocompletion script for the specified shell
  confgen     Generates config files, used notably for local & remote deployment
  deploylocal Locally deploys the app, i.e. its API and / or its client (web) app
  help        Help about any command
  refresh     Refreshes the Aldev environment's required external resources

Flags:
  -a, --api               when developing a pure API (Linux)
  -d, --disable-confgen   disable the generation of all the config files
  -f, --file string       aldev config file (default ".aldev.yaml")
  -h, --help              help for aldev
  -l, --lib               when developing a library (Linux)
  -n, --native            when developing a native app (Windows)
  -s, --swap              use swapping of code, to use the local version of some dependencies for instance
  -v, --verbose           activates debug logging
  -w, --web               when developing a webapp, along with its API (Linux)

Use "aldev [command] --help" for more information about a command.
```
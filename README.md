# Aldev

Aldev is a powerful CLI tool designed to streamline development with the Goald / GoaldR / GoaldN stack. It provides a comprehensive set of tools for rapid development, including automatic deployment in local Kubernetes clusters and live reloading capabilities.

## Features

- **Bootstrap**: Quickly set up new Goald/GoaldR/GoaldN projects
- **Code Generation**: Automatically generate boilerplate code to speed up development
- **Code Swapping**: Easily swap code segments (like import paths) in targeted files
- **Config Generation**: Generate configuration files for both local and remote deployments
- **Local Deployment**: Deploy applications locally, supporting both API and web client applications
- **Environment Management**: Refresh and manage external resources required by the development environment

## Installation

```sh
# Installation instructions will be added here
```

## Usage

Run `aldev` to start or continue developing a Goald / GoaldR or GoaldN application. The tool provides automatic deployment in a local k8s cluster and live reloading when applicable.

```sh
$ aldev -h
```

### Available Commands

- `bootstrap`: Bootstraps a new aldev project
- `codegen`: Completes the app with additional generated code to speed up your dev
- `codeswap`: Swaps bit of code - like import paths - in targeted files
- `completion`: Generate the autocompletion script for the specified shell
- `confgen`: Generates config files, used notably for local & remote deployment
- `deploylocal`: Locally deploys the app, i.e. its API and / or its client (web) app
- `refresh`: Refreshes the Aldev environment's required external resources

### Flags

- `-a, --api`: When developing a pure API (Linux)
- `-d, --disable-confgen`: Disable the generation of all the config files
- `-f, --file`: Specify aldev config file (default ".aldev.yaml")
- `-h, --help`: Show help information
- `-l, --lib`: When developing a library (Linux)
- `-n, --native`: When developing a native app (Windows)
- `-s, --swap`: Use swapping of code, to use the local version of some dependencies
- `-v, --verbose`: Activate debug logging
- `-w, --web`: When developing a webapp, along with its API (Linux)

## Configuration

Aldev uses a configuration file (default: `.aldev.yaml`) to manage project settings. Use the `-f` flag to specify a different configuration file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the terms of the license included in the repository.

## Support

For support and questions, please open an issue in the repository.
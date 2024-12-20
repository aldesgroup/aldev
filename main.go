/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/aldesgroup/aldev/cmd"
	_ "github.com/aldesgroup/aldev/cmd/bootstrap"
	_ "github.com/aldesgroup/aldev/cmd/codegen"
	_ "github.com/aldesgroup/aldev/cmd/codeswap"
	_ "github.com/aldesgroup/aldev/cmd/confgen"
	_ "github.com/aldesgroup/aldev/cmd/deploylocal"
	_ "github.com/aldesgroup/aldev/cmd/refresh"
)

func main() {
	cmd.Execute()
}

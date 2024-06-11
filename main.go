/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/aldesgroup/aldev/cmd"
	_ "github.com/aldesgroup/aldev/cmd/complete"
	_ "github.com/aldesgroup/aldev/cmd/download"
	_ "github.com/aldesgroup/aldev/cmd/generate"
	_ "github.com/aldesgroup/aldev/cmd/init"
	_ "github.com/aldesgroup/aldev/cmd/launch"
	_ "github.com/aldesgroup/aldev/cmd/swap"
)

func main() {
	cmd.Execute()
}

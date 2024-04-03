/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/aldesgroup/aldev/cmd"
	_ "github.com/aldesgroup/aldev/cmd/build"
	_ "github.com/aldesgroup/aldev/cmd/update"
)

func main() {
	cmd.Execute()
}

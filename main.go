package main

import (
	"fmt"
	"os"

	"github.com/KriKri98/gator/internal/cli"
	"github.com/KriKri98/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	state := cli.Status{}
	state.Cfg = &cfg

	commands := cli.Commands{}
	commands.Command = make(map[string]func(*cli.Status, cli.Command) error)
	commands.Register("login", cli.HandlerLogin)

	args := os.Args

	if len(args) < 2 {
		fmt.Print("not enough arguments given\n")
		os.Exit(1)
	}

	command := cli.Command{
		Name: args[1],
		Args: args[2:],
	}

	err = commands.Run(&state, command)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

}

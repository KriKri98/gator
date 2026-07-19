package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/KriKri98/gator/internal/cli"
	"github.com/KriKri98/gator/internal/config"
	"github.com/KriKri98/gator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	state := cli.Status{}
	state.Cfg = &cfg

	db, err := sql.Open("postgres", state.Cfg.DB_url)
	dbQueries := database.New(db)
	state.DB = dbQueries

	commands := cli.Commands{}
	commands.Command = make(map[string]func(*cli.Status, cli.Command) error)
	commands.Register("login", cli.HandlerLogin)
	commands.Register("register", cli.HandlerRegister)
	commands.Register("reset", cli.HandlerReset)
	commands.Register("users", cli.HandlerGetUsers)
	commands.Register("agg", cli.HandlerAgg)
	commands.Register("addfeed", cli.MiddlewareLoggedIn(cli.HandlerAddFeed))
	commands.Register("feeds", cli.HandlerAllFeeds)
	commands.Register("follow", cli.MiddlewareLoggedIn(cli.HandlerFollow))
	commands.Register("following", cli.MiddlewareLoggedIn(cli.HandlerFollowing))

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

package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/KriKri98/gator/internal/config"
	"github.com/KriKri98/gator/internal/database"
	"github.com/google/uuid"
)

type Status struct {
	Cfg *config.Config
	DB  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Command map[string]func(*Status, Command) error
}

func (c *Commands) Run(s *Status, cmd Command) error {
	err := c.Command[cmd.Name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) Register(name string, f func(*Status, Command) error) {
	c.Command[name] = f
}

func HandlerLogin(s *Status, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}

	user, err := s.DB.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		fmt.Printf("username does not exist")
		os.Exit(1)
	}

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User %v has been set\n", user.Name)
	return nil
}

func HandlerRegister(s *Status, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}
	user, err := s.DB.CreateUser(context.Background(), userParams)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("created user: %v", user)

	return nil
}

func HandlerReset(s *Status, cmd Command) error {
	err := s.DB.DeleteUsers(context.Background())
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	return nil
}

func HandlerGetUsers(s *Status, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.Cfg.Current_user_name {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

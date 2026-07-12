package cli

import (
	"fmt"

	"github.com/KriKri98/gator/internal/config"
)

type Status struct {
	Cfg *config.Config
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
	err := s.Cfg.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %v has been set\n", cmd.Args[0])
	return nil
}

package main

import (
	"fmt"

	"github.com/KriKri98/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = cfg.SetUser("michael")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("%v\n", cfg)

}

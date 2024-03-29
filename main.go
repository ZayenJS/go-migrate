package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/zayenjs/go-migrate/cli"
)

func main() {
	cli.Setup()

	args := flag.Args()
	for _, arg := range args {
		switch arg {
		case cli.Commands.Init:
			cli.Init()
		case cli.Commands.Create:
			cli.Create()
		case cli.Commands.Migrate:
			cli.Migrate()
		case cli.Commands.Rollback:
			args := flag.Args()

			var steps int64
			var err error

			if len(args) > 1 && len(args) < 3 {
				stepsArg := flag.Args()[1]
				steps, err = strconv.ParseInt(stepsArg, 10, 64)

				if err != nil {
					fmt.Println("Invalid steps")
					os.Exit(1)
				}
			}

			if steps == 0 {
				steps = 1
			}

			cli.Rollback(steps)
		}
	}
}

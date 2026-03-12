package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aruis/wifictl/internal/app"
	"github.com/aruis/wifictl/internal/cli"
	"github.com/aruis/wifictl/internal/macos"
)

func main() {
	command, err := cli.Parse(os.Args[1:])
	if err != nil {
		if err == cli.ErrHelp {
			fmt.Fprint(os.Stdout, cli.Usage())
			return
		}

		fmt.Fprintf(os.Stderr, "error: %v\n\n%s", err, cli.Usage())
		os.Exit(2)
	}

	if err := app.New(macos.New(), os.Stdout).Run(context.Background(), command); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"os"

	gosocketcli "github.com/awoldt/gosocket/cli"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func main() {
	var StartCommand = &cli.Command{
		Description: "A lightweight websocket server written in Go",
		Flags: []cli.Flag{
			gosocketcli.PortFlag,
			gosocketcli.LoggingFlag,
		},
		Action: gosocketcli.StartServer,
	}

	if err := StartCommand.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

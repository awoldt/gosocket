package main

import (
	"context"
	"os"

	"github.com/awoldt/gosocket/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cli.StartCommand.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

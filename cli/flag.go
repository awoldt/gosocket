package cli

import "github.com/urfave/cli/v3"

var PortFlag = &cli.StringFlag{
	Name:  "port",
	Usage: "Port websocket server listens on",
	Value: "3000",
}

var LoggingFlag = &cli.BoolFlag{
	Name:  "log",
	Usage: "Enables logging on server",
}

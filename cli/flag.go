package cli

import "github.com/urfave/cli/v3"

const PortFlagName = "port"
const LogFlagName = "log"

var PortFlag = &cli.StringFlag{
	Name:  PortFlagName,
	Usage: "Port websocket server listens on",
	Value: "3000",
}

var LoggingFlag = &cli.BoolFlag{
	Name:  LogFlagName,
	Usage: "Enables logging on server",
	Value: false,
}

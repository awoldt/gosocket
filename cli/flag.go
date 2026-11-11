package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

type FlagConfig struct {
	Port string
	Log  bool
}

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

func GetFlagConfig(c *cli.Command) FlagConfig {
	Port := c.String(PortFlagName)
	Log := c.Bool(LogFlagName)

	return FlagConfig{
		Port,
		Log,
	}
}

// little helper function that will log to terminal
// if "log" flag was passed by user
func logInfo(shouldLog bool, message string) {
	if shouldLog {
		logrus.Info(message)
	}
}

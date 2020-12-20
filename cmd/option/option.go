package option

import "github.com/urfave/cli"

var Flags = []cli.Flag{
	cli.BoolFlag{
		Name:     "cr",
		Usage:    "Resource is a Custom resource",
		Required: false,
	},
	cli.StringFlag{
		Name:     "apis",
		Usage:    "Custom resource api def package (without version)",
		Required: false,
	},
	cli.StringFlag{
		Name:     "client, c",
		Usage:    "Custom resource typed client package name",
		Required: false,
	},
	cli.StringFlag{
		Name:     "scheme, s",
		Usage:    "Custom resource scheme package name",
		Required: false,
	},
}

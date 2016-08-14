package main

import (
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	_ "github.com/joho/godotenv/autoload"
)

var version string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "swift artifact plugin"
	app.Usage = "swift artifact plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "endpoint for auth the swift connection",
			EnvVar: "PLUGIN_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "swift user name",
			EnvVar: "PLUGIN_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "swift api key",
			EnvVar: "PLUGIN_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "container",
			Usage:  "swift container",
			EnvVar: "PLUGIN_CONTAINER",
		},
		cli.IntFlag{
			Name:   "auth-version",
			Usage:  "swift auth version",
			EnvVar: "PLUGIN_AUTH_VERSION",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "swift region",
			EnvVar: "PLUGIN_REGION",
		},
		cli.StringFlag{
			Name:   "tenant",
			Usage:  "swift tenant",
			EnvVar: "PLUGIN_TENANT",
		},
		cli.StringFlag{
			Name:   "source",
			Usage:  "upload files from source folder",
			EnvVar: "PLUGIN_SOURCE",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "upload files to target folder",
			EnvVar: "PLUGIN_TARGET",
		},
		cli.StringFlag{
			Name:   "strip-prefix",
			Usage:  "strip the prefix from the target",
			EnvVar: "PLUGIN_STRIP_PREFIX",
		},
		cli.BoolFlag{
			Name:   "recursive",
			Usage:  "upload files recursively",
			EnvVar: "PLUGIN_RECURSIVE",
		},
		cli.StringSliceFlag{
			Name:   "exclude",
			Usage:  "ignore files matching exclude pattern",
			EnvVar: "PLUGIN_EXCLUDE",
		},
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run for debug purposes",
			EnvVar: "PLUGIN_DRY_RUN",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{
		Endpoint:    c.String("endpoint"),
		Key:         c.String("access-key"),
		Secret:      c.String("secret-key"),
		Container:   c.String("continer"),
		AuthVersion: c.Int("auth-version"),
		Region:      c.String("region"),
		Tenant:      c.String("tenant"),
		Source:      c.String("source"),
		Target:      c.String("target"),
		StripPrefix: c.String("strip-prefix"),
		Recursive:   c.Bool("recursive"),
		Exclude:     c.StringSlice("exclude"),
		DryRun:      c.Bool("dry-run"),
	}

	// normalize the target URL
	if strings.HasPrefix(plugin.Target, "/") {
		plugin.Target = plugin.Target[1:]
	}

	return plugin.Exec()
}

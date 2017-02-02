package main

import (
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

// Version set at compile-time
var Version string

func main() {
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "swift artifact plugin"
	app.Usage = "swift artifact plugin"
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "endpoint for auth the swift connection",
			EnvVar: "PLUGIN_ENDPOINT,SWIFT_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "swift user name",
			EnvVar: "PLUGIN_ACCESS_KEY,SWIFT_USER",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "swift api key",
			EnvVar: "PLUGIN_SECRET_KEY,SWIFT_KEY",
		},
		cli.StringFlag{
			Name:   "container",
			Usage:  "swift container",
			EnvVar: "PLUGIN_CONTAINER,SWIFT_CONTAINER",
		},
		cli.IntFlag{
			Name:   "auth-version",
			Usage:  "swift auth version",
			EnvVar: "PLUGIN_AUTH_VERSION,SWIFT_VERSION",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "swift region",
			EnvVar: "PLUGIN_REGION,SWIFT_REGION",
		},
		cli.StringFlag{
			Name:   "timeout",
			Usage:  "timeout",
			EnvVar: "PLUGIN_TIMEOUT,SWIFT_TIMEOUT",
		},
		cli.StringFlag{
			Name:   "tenant",
			Usage:  "swift tenant",
			EnvVar: "PLUGIN_TENANT,SWIFT_TENANT",
		},
		cli.StringFlag{
			Name:   "source",
			Usage:  "upload files from source folder",
			EnvVar: "PLUGIN_SOURCE,SWIFT_SOURCE",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "upload files to target folder",
			EnvVar: "PLUGIN_TARGET,SWIFT_TARGET",
		},
		cli.StringFlag{
			Name:   "strip-prefix",
			Usage:  "strip the prefix from the target",
			EnvVar: "PLUGIN_STRIP_PREFIX,SWIFT_STRIP_PREFIX",
		},
		cli.StringSliceFlag{
			Name:   "exclude",
			Usage:  "ignore files matching exclude pattern",
			EnvVar: "PLUGIN_EXCLUDE,SWIFT_EXCLUDE",
		},
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run for debug purposes",
			EnvVar: "PLUGIN_DRY_RUN",
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {

	plugin := &Plugin{
		Endpoint:    c.String("endpoint"),
		Timeout:     c.String("timeout"),
		Key:         c.String("access-key"),
		Secret:      c.String("secret-key"),
		Container:   c.String("container"),
		AuthVersion: c.Int("auth-version"),
		Region:      c.String("region"),
		Tenant:      c.String("tenant"),
		Source:      c.String("source"),
		Target:      c.String("target"),
		StripPrefix: c.String("strip-prefix"),
		Exclude:     c.StringSlice("exclude"),
		DryRun:      c.Bool("dry-run"),
	}

	// normalize the target URL
	if strings.HasPrefix(plugin.Target, "/") {
		plugin.Target = plugin.Target[1:]
	}
	return plugin.Exec()
}

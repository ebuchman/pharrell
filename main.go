package main

import (
	"os"
	"path"

	"github.com/codegangsta/cli"
)

var (
	rootDir = path.Join(home(), ".pharrell")
)

func init() {
	if _, err := os.Stat(rootDir); err != nil {
		ifExit(os.MkdirAll(rootDir, 0700))
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "pharrell"
	app.Usage = "pharrell <scp/ssh> args..."
	app.Version = "0.0.1"
	app.Author = "Ethan Buchman"
	app.Email = "ethan@erisindustries.com"

	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		sshCmd,
		scpCmd,
	}

	app.Run(os.Args)
}

var (
	sshCmd = cli.Command{
		Name:   "ssh",
		Usage:  "pharrell ssh -u minty -h mumbojumbo:22  <\"cmds\" | cmdsfile>",
		Action: cliSSH,
		Flags: []cli.Flag{
			userFlag,
			hostFlag,
			outFlag,
		},
	}

	scpCmd = cli.Command{
		Name:   "scp",
		Usage:  "pharrell scp -u minty -h mumbojump:22 src dst",
		Action: cliSCP,
		Flags: []cli.Flag{
			userFlag,
			hostFlag,
			outFlag,
		},
	}

	userFlag = cli.StringFlag{
		Name:  "user, u",
		Usage: "username",
		Value: "root",
	}

	hostFlag = cli.StringFlag{
		Name:  "host",
		Usage: "host or file containing host names, one per line",
	}

	outFlag = cli.StringFlag{
		Name:  "out, o",
		Usage: "output is either file, stdout, or unknown",
		Value: "file",
	}
)

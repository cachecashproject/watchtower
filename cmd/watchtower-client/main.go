package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	app := cli.NewApp()
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "cert, c",
			Usage: "Path to Client Certificate",
			Value: "client.crt",
		},
		cli.StringFlag{
			Name:  "host, t",
			Usage: "host:port of watchtower update control service",
			Value: "localhost:4001",
		},
		cli.BoolFlag{
			Name:  "insecure, i",
			Usage: "Use insecure connections",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "update-version",
			ShortName:   "uv",
			Usage:       "Update the version of a container image in watchtower.",
			Description: "Update the version of a container image in watchtower.",
			Subcommands: []cli.Command{
				{
					Name:        "image",
					ShortName:   "i",
					ArgsUsage:   "[options] [tag]",
					Usage:       "Updates the version of an image from its tag. You must have access to docker.",
					Description: "Updates the version of an image from its tag. You must have access to docker.",
				},
				{
					Name:        "literal",
					ShortName:   "l",
					ArgsUsage:   "[options] [tag] [version]",
					Usage:       "Updates the version of an image from its provided literal. Only string validation is performed; no image validation.",
					Description: "Updates the version of an image from its provided literal. Only string validation is performed; no image validation.",
					Action:      updateFromLiteral,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getClient(ctx *cli.Context) (grpcmsg.UpdateControlClient, error) {
	opts := []grpc.DialOption{}

	if ctx.GlobalBool("insecure") {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tls, err := credentials.NewClientTLSFromFile(ctx.GlobalString("cert"), "")
		if err != nil {
			return nil, errors.Wrap(err, "error loading certificate")
		}

		opts = append(opts, grpc.WithTransportCredentials(tls))
	}

	cc, err := grpc.Dial(ctx.GlobalString("host"), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "while dialing watchtower @ %q", ctx.GlobalString("host"))
	}

	return grpcmsg.NewUpdateControlClient(cc), nil
}

func updateFromLiteral(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return errors.New("invalid arguments; seek --help")
	}

	client, err := getClient(ctx)
	if err != nil {
		return errors.Wrap(err, "while constructing client")
	}

	_, err = client.SetLatestUpdate(context.Background(), &grpcmsg.ContainerImage{Name: ctx.Args()[0], Version: ctx.Args()[1]})
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	fmt.Println("Update successful!")

	return nil
}

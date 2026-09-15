package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strconv"

	"klabautermann/internal/adb"
	"klabautermann/internal/pipeline"
)

type command struct {
	run func(ctx context.Context, c *adb.Client, args []string, stdout, stderr io.Writer) int
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("klabautermann", flag.ContinueOnError)
	fs.SetOutput(stderr)
	adbPath := fs.String("adb", "adb", "path to adb binary")
	serial := fs.String("serial", adb.DefaultSerial, "device serial")
	pipe := fs.Bool("pipe", false, "run as daemon with pipe protocol")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *pipe {
		fmt.Fprintln(stderr, "pipe daemon: not implemented")
		return 1
	}
	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(stderr, "usage: klabautermann [--adb <path>] [--serial <serial>] <command> [args]")
		fmt.Fprintln(stderr, "error: no command")
		return 2
	}
	client, err := adb.New(adb.Config{AdbPath: *adbPath, Serial: *serial})
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	commands := map[string]command{
		"connect": {
			run: func(ctx context.Context, c *adb.Client, args []string, stdout, stderr io.Writer) int {
				if len(args) != 0 {
					fmt.Fprintln(stderr, "usage: klabautermann connect")
					fmt.Fprintln(stderr, "error: connect takes no arguments")
					return 2
				}
				if err := c.Connect(ctx); err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				fmt.Fprintf(stdout, "connected to %s\n", *serial)
				return 0
			},
		},
		"screencap": {
			run: func(ctx context.Context, c *adb.Client, args []string, stdout, stderr io.Writer) int {
				sfs := flag.NewFlagSet("screencap", flag.ContinueOnError)
				sfs.SetOutput(stderr)
				out := sfs.String("o", "screen.png", "output path")
				if err := sfs.Parse(args); err != nil {
					fmt.Fprintln(stderr, "usage: klabautermann screencap [-o <path>]")
					return 2
				}
				if sfs.NArg() != 0 {
					fmt.Fprintln(stderr, "usage: klabautermann screencap [-o <path>]")
					fmt.Fprintln(stderr, "error: screencap takes no positional arguments")
					return 2
				}
				data, err := c.Screencap(ctx)
				if err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				if err := os.WriteFile(*out, data, 0644); err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				fmt.Fprintf(stdout, "wrote %s (%d bytes)\n", *out, len(data))
				return 0
			},
		},
		"tap": {
			run: func(ctx context.Context, c *adb.Client, args []string, stdout, stderr io.Writer) int {
				if len(args) != 2 {
					fmt.Fprintln(stderr, "usage: klabautermann tap <x> <y>")
					fmt.Fprintln(stderr, "error: tap requires 2 arguments")
					return 2
				}
				x, err := strconv.Atoi(args[0])
				if err != nil {
					fmt.Fprintln(stderr, "usage: klabautermann tap <x> <y>")
					fmt.Fprintln(stderr, "error: invalid integer "+args[0])
					return 2
				}
				y, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Fprintln(stderr, "usage: klabautermann tap <x> <y>")
					fmt.Fprintln(stderr, "error: invalid integer "+args[1])
					return 2
				}
				if err := c.Tap(ctx, x, y); err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				fmt.Fprintf(stdout, "tapped %d %d\n", x, y)
				return 0
			},
		},
		"swipe": {
			run: func(ctx context.Context, c *adb.Client, args []string, stdout, stderr io.Writer) int {
				if len(args) != 5 {
					fmt.Fprintln(stderr, "usage: klabautermann swipe <x1> <y1> <x2> <y2> <ms>")
					fmt.Fprintln(stderr, "error: swipe requires 5 arguments")
					return 2
				}
				vals := make([]int, 5)
				for i, s := range args {
					v, err := strconv.Atoi(s)
					if err != nil {
						fmt.Fprintln(stderr, "usage: klabautermann swipe <x1> <y1> <x2> <y2> <ms>")
						fmt.Fprintln(stderr, "error: invalid integer "+s)
						return 2
					}
					vals[i] = v
				}
				if err := c.Swipe(ctx, vals[0], vals[1], vals[2], vals[3], vals[4]); err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				fmt.Fprintf(stdout, "swiped %d %d %d %d %dms\n", vals[0], vals[1], vals[2], vals[3], vals[4])
				return 0
			},
		},
		"run": {
			run: func(ctx context.Context, c *adb.Client, args []string, stdout, stderr io.Writer) int {
				if len(args) != 1 {
					fmt.Fprintln(stderr, "usage: klabautermann run <pipeline.json>")
					fmt.Fprintln(stderr, "error: run requires 1 argument")
					return 2
				}
				p, err := pipeline.Load(args[0])
				if err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				w := &pipeline.Walker{
					Nodes: p.Nodes,
					Capture: func(ctx context.Context) (image.Image, error) {
						data, err := c.Screencap(ctx)
						if err != nil {
							return nil, err
						}
						img, _, err := image.Decode(bytes.NewReader(data))
						if err != nil {
							return nil, err
						}
						return img, nil
					},
					Tap: func(ctx context.Context, x, y int) error {
						return c.Tap(ctx, x, y)
					},
				}
				hits, err := w.Run(ctx, p.Start)
				if err != nil {
					fmt.Fprintln(stderr, err.Error())
					return 1
				}
				for _, h := range hits {
					fmt.Fprintf(stdout, "hit %s %.3f %v\n", h.Node, h.Score, h.At)
				}
				return 0
			},
		},
	}
	cmd, ok := commands[rest[0]]
	if !ok {
		fmt.Fprintln(stderr, "usage: klabautermann [--adb <path>] [--serial <serial>] <command> [args]")
		fmt.Fprintln(stderr, "error: unknown command "+rest[0])
		return 2
	}
	return cmd.run(ctx, client, rest[1:], stdout, stderr)
}

package adb

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

const DefaultSerial = "127.0.0.1:16384"

const DefaultTimeout = 10 * time.Second

var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

type Config struct {
	AdbPath string
	Serial  string
	Timeout time.Duration
}

type Client struct {
	adbPath string
	serial  string
	timeout time.Duration
	run     func(ctx context.Context, args []string) ([]byte, error)
}

func New(cfg Config) (*Client, error) {
	if cfg.AdbPath == "" {
		cfg.AdbPath = "adb"
	}
	if cfg.Serial == "" {
		cfg.Serial = DefaultSerial
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	c := &Client{adbPath: cfg.AdbPath, serial: cfg.Serial, timeout: cfg.Timeout}
	c.run = func(ctx context.Context, args []string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, c.adbPath, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("adb: adb %v: %w: %s", args, err, stderr.String())
		}
		return stdout.Bytes(), nil
	}
	return c, nil
}

func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}

func (c *Client) Screencap(ctx context.Context) ([]byte, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	out, err := c.run(ctx, []string{"-s", c.serial, "exec-out", "screencap", "-p"})
	if err != nil {
		return nil, err
	}
	if len(out) < len(pngMagic) || !bytes.Equal(out[:len(pngMagic)], pngMagic) {
		return nil, fmt.Errorf("adb: screencap returned %d bytes without PNG signature", len(out))
	}
	return out, nil
}

func (c *Client) Tap(ctx context.Context, x, y int) error {
	if x < 0 || y < 0 {
		return fmt.Errorf("adb: tap coordinates must be non-negative, got (%d, %d)", x, y)
	}
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	_, err := c.run(ctx, []string{
		"-s", c.serial, "shell", "input", "tap",
		strconv.Itoa(x), strconv.Itoa(y),
	})
	return err
}

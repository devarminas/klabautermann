package adb

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func mustNew(t *testing.T, cfg Config) *Client {
	t.Helper()
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("New(%+v) error = %v", cfg, err)
	}
	return c
}

func TestNewDefaults(t *testing.T) {
	c := mustNew(t, Config{})
	if c.adbPath != "adb" {
		t.Errorf("adbPath = %q, want adb", c.adbPath)
	}
	if c.serial != DefaultSerial {
		t.Errorf("serial = %q, want %q", c.serial, DefaultSerial)
	}
	if c.timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.timeout, DefaultTimeout)
	}
}

func TestNewKeepsExplicitConfig(t *testing.T) {
	c := mustNew(t, Config{AdbPath: "C:/emulator/adb.exe", Serial: "127.0.0.1:7555", Timeout: time.Second})
	if c.adbPath != "C:/emulator/adb.exe" || c.serial != "127.0.0.1:7555" || c.timeout != time.Second {
		t.Errorf("config not preserved: %+v", c)
	}
}

func TestScreencapArgsAndPNG(t *testing.T) {
	c := mustNew(t, Config{Serial: "test-serial"})
	var gotArgs []string
	png := append(append([]byte{}, pngMagic...), 0x00, 0x01)
	c.run = func(_ context.Context, args []string) ([]byte, error) {
		gotArgs = args
		return png, nil
	}
	out, err := c.Screencap(context.Background())
	if err != nil {
		t.Fatalf("Screencap error = %v", err)
	}
	if !reflect.DeepEqual(out, png) {
		t.Errorf("Screencap returned %v, want %v", out, png)
	}
	want := []string{"-s", "test-serial", "exec-out", "screencap", "-p"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestScreencapRejectsNonPNG(t *testing.T) {
	c := mustNew(t, Config{})
	c.run = func(_ context.Context, _ []string) ([]byte, error) {
		return []byte("not a png"), nil
	}
	if _, err := c.Screencap(context.Background()); err == nil ||
		!strings.Contains(err.Error(), "PNG signature") {
		t.Fatalf("Screencap error = %v, want PNG signature complaint", err)
	}
}

func TestScreencapPropagatesExecError(t *testing.T) {
	c := mustNew(t, Config{})
	sentinel := errors.New("device offline")
	c.run = func(_ context.Context, _ []string) ([]byte, error) { return nil, sentinel }
	if _, err := c.Screencap(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("Screencap error = %v, want %v", err, sentinel)
	}
}

func TestScreencapAppliesTimeout(t *testing.T) {
	c := mustNew(t, Config{Timeout: 50 * time.Millisecond})
	c.run = func(ctx context.Context, _ []string) ([]byte, error) {
		dl, ok := ctx.Deadline()
		if !ok {
			t.Error("run ctx has no deadline")
			return append([]byte{}, pngMagic...), nil
		}
		if time.Until(dl) > 50*time.Millisecond {
			t.Errorf("deadline %v exceeds timeout", dl)
		}
		return append([]byte{}, pngMagic...), nil
	}
	if _, err := c.Screencap(context.Background()); err != nil {
		t.Fatalf("Screencap error = %v", err)
	}
}

func TestTapArgs(t *testing.T) {
	c := mustNew(t, Config{Serial: "test-serial"})
	var gotArgs []string
	c.run = func(_ context.Context, args []string) ([]byte, error) {
		gotArgs = args
		return nil, nil
	}
	if err := c.Tap(context.Background(), 640, 360); err != nil {
		t.Fatalf("Tap error = %v", err)
	}
	want := []string{"-s", "test-serial", "shell", "input", "tap", "640", "360"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestTapRejectsNegativeCoords(t *testing.T) {
	c := mustNew(t, Config{})
	called := false
	c.run = func(_ context.Context, _ []string) ([]byte, error) {
		called = true
		return nil, nil
	}
	for _, tc := range [][2]int{{-1, 0}, {0, -1}, {-5, -5}} {
		if err := c.Tap(context.Background(), tc[0], tc[1]); err == nil {
			t.Errorf("Tap(%d, %d) error = nil, want non-negative complaint", tc[0], tc[1])
		}
	}
	if called {
		t.Error("run invoked for invalid coordinates")
	}
}

func TestTapPropagatesExecError(t *testing.T) {
	c := mustNew(t, Config{})
	sentinel := errors.New("device offline")
	c.run = func(_ context.Context, _ []string) ([]byte, error) { return nil, sentinel }
	if err := c.Tap(context.Background(), 1, 1); !errors.Is(err, sentinel) {
		t.Fatalf("Tap error = %v, want %v", err, sentinel)
	}
}

func TestConnectArgs(t *testing.T) {
	c := mustNew(t, Config{Serial: "test-serial"})
	var gotArgs []string
	c.run = func(_ context.Context, args []string) ([]byte, error) {
		gotArgs = args
		return nil, nil
	}
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect error = %v", err)
	}
	want := []string{"connect", "test-serial"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestConnectPropagatesExecError(t *testing.T) {
	c := mustNew(t, Config{})
	sentinel := errors.New("device offline")
	c.run = func(_ context.Context, _ []string) ([]byte, error) { return nil, sentinel }
	if err := c.Connect(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("Connect error = %v, want %v", err, sentinel)
	}
}

func TestSwipeArgs(t *testing.T) {
	c := mustNew(t, Config{Serial: "test-serial"})
	var gotArgs []string
	c.run = func(_ context.Context, args []string) ([]byte, error) {
		gotArgs = args
		return nil, nil
	}
	if err := c.Swipe(context.Background(), 100, 200, 300, 400, 500); err != nil {
		t.Fatalf("Swipe error = %v", err)
	}
	want := []string{"-s", "test-serial", "shell", "input", "swipe", "100", "200", "300", "400", "500"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestSwipeRejectsNegativeInput(t *testing.T) {
	c := mustNew(t, Config{})
	called := false
	c.run = func(_ context.Context, _ []string) ([]byte, error) {
		called = true
		return nil, nil
	}
	for _, tc := range [][5]int{{-1, 0, 0, 0, 100}, {0, 0, -5, 0, 100}, {0, 0, 100, 100, -1}} {
		if err := c.Swipe(context.Background(), tc[0], tc[1], tc[2], tc[3], tc[4]); err == nil {
			t.Errorf("Swipe(%d, %d, %d, %d, %d) error = nil, want non-negative complaint", tc[0], tc[1], tc[2], tc[3], tc[4])
		}
	}
	if called {
		t.Error("run invoked for invalid swipe input")
	}
}

func TestSwipePropagatesExecError(t *testing.T) {
	c := mustNew(t, Config{})
	sentinel := errors.New("device offline")
	c.run = func(_ context.Context, _ []string) ([]byte, error) { return nil, sentinel }
	if err := c.Swipe(context.Background(), 100, 200, 300, 400, 500); !errors.Is(err, sentinel) {
		t.Fatalf("Swipe error = %v, want %v", err, sentinel)
	}
}

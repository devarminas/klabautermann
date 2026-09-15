package main

import (
	"bytes"
	"context"
	"testing"
)

func TestRunUsageErrors(t *testing.T) {
	cases := [][]string{
		{},
		{"bogus"},
		{"connect", "extra"},
		{"tap", "1"},
		{"tap", "a", "b"},
		{"swipe", "a", "b", "c", "d", "e"},
		{"screencap", "--bogus"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, &stdout, &stderr); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
		if stderr.Len() == 0 {
			t.Errorf("run(%v) wrote no stderr", args)
		}
	}
}

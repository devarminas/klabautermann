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
		{"--display", "2", "tap", "1"},
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

func TestRunRejectsUnknownCapture(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"run", "--capture", "bogus", "no-such-pipeline.json"}, &stdout, &stderr); code != 2 {
		t.Errorf("run([run --capture bogus no-such-pipeline.json]) = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Error("run([run --capture bogus no-such-pipeline.json]) wrote no stderr")
	}
}

func TestRunPipelineArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"run"}, &stdout, &stderr); code != 2 {
		t.Errorf("run([run]) = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Errorf("run([run]) wrote no stderr")
	}
	var out2, err2 bytes.Buffer
	if code := run(context.Background(), []string{"run", "no-such-pipeline.json"}, &out2, &err2); code != 1 {
		t.Errorf("run([run no-such-pipeline.json]) = %d, want 1", code)
	}
	if err2.Len() == 0 {
		t.Errorf("run([run no-such-pipeline.json]) wrote no stderr")
	}
}

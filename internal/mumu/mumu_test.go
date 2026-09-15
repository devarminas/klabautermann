package mumu

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"
)

func TestCaptureFlipsBottomUpRGBA(t *testing.T) {
	bottomUp := []byte{
		0, 0, 255, 255, 255, 255, 255, 255,
		255, 0, 0, 255, 0, 255, 0, 255,
	}
	calls := 0
	c := &Client{
		handle: 7,
		raw: rawAPI{
			capture: func(handle uintptr, displayID int, size int, w *int32, h *int32, buf *byte) int32 {
				calls++
				if handle != 7 {
					t.Errorf("capture handle = %d, want 7", handle)
				}
				if displayID != 3 {
					t.Errorf("capture displayID = %d, want 3", displayID)
				}
				*w = 2
				*h = 2
				if buf == nil {
					if size != 0 {
						t.Errorf("probe size = %d, want 0", size)
					}
					return 0
				}
				if size != len(bottomUp) {
					t.Errorf("fill size = %d, want %d", size, len(bottomUp))
				}
				copy(unsafe.Slice(buf, size), bottomUp)
				return 0
			},
		},
	}
	img, err := c.Capture(3)
	if err != nil {
		t.Fatalf("Capture error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("capture calls = %d, want 2", calls)
	}
	rgba, ok := img.(*image.RGBA)
	if !ok {
		t.Fatalf("Capture returned %T, want *image.RGBA", img)
	}
	if rgba.Bounds() != image.Rect(0, 0, 2, 2) {
		t.Fatalf("Bounds = %v, want 2x2", rgba.Bounds())
	}
	wantPix := []byte{
		255, 0, 0, 255, 0, 255, 0, 255,
		0, 0, 255, 255, 255, 255, 255, 255,
	}
	if !bytes.Equal(rgba.Pix, wantPix) {
		t.Fatalf("Pix = %v, want %v", rgba.Pix, wantPix)
	}
}

func TestCaptureFailureNamesDisplay(t *testing.T) {
	c := &Client{
		handle: 1,
		raw: rawAPI{
			capture: func(handle uintptr, displayID int, size int, w *int32, h *int32, buf *byte) int32 {
				*w = 2
				*h = 2
				return 5
			},
		},
	}
	_, err := c.Capture(42)
	if err == nil {
		t.Fatal("Capture error = nil, want failure")
	}
	if !strings.HasPrefix(err.Error(), "mumu: ") {
		t.Errorf("error %q missing mumu prefix", err.Error())
	}
	if !strings.Contains(err.Error(), "42") {
		t.Errorf("error %q does not name display 42", err.Error())
	}
}

func TestDisplayIDNegativeFails(t *testing.T) {
	c := &Client{
		handle: 1,
		raw: rawAPI{
			displayID: func(handle uintptr, pkg *byte, appIndex int) int32 {
				return -1
			},
		},
	}
	_, err := c.DisplayID("com.example.game", 0)
	if err == nil {
		t.Fatal("DisplayID error = nil, want failure")
	}
	if !strings.HasPrefix(err.Error(), "mumu: ") {
		t.Errorf("error %q missing mumu prefix", err.Error())
	}
	if !strings.Contains(err.Error(), "com.example.game") {
		t.Errorf("error %q does not name the package", err.Error())
	}
}

func TestDisplayIDRejectsEmptyPackage(t *testing.T) {
	called := false
	c := &Client{
		handle: 1,
		raw: rawAPI{
			displayID: func(handle uintptr, pkg *byte, appIndex int) int32 {
				called = true
				return 0
			},
		},
	}
	_, err := c.DisplayID("", 0)
	if err == nil {
		t.Fatal("DisplayID error = nil, want usage complaint")
	}
	if !strings.Contains(err.Error(), "display 0") {
		t.Errorf("error %q does not point at display 0", err.Error())
	}
	if called {
		t.Error("raw displayID invoked for empty package")
	}
}

func TestCloseZeroHandleIsNoop(t *testing.T) {
	called := false
	c := &Client{
		raw: rawAPI{
			disconnect: func(handle uintptr) uintptr {
				called = true
				return 0
			},
		},
	}
	c.Close()
	if called {
		t.Error("disconnect invoked for zero handle")
	}
}

func TestCloseDisconnectsOnce(t *testing.T) {
	calls := 0
	c := &Client{
		handle: 9,
		raw: rawAPI{
			disconnect: func(handle uintptr) uintptr {
				calls++
				if handle != 9 {
					t.Errorf("disconnect handle = %d, want 9", handle)
				}
				return 0
			},
		},
	}
	c.Close()
	c.Close()
	if calls != 1 {
		t.Errorf("disconnect calls = %d, want 1", calls)
	}
	if c.handle != 0 {
		t.Errorf("handle = %d, want 0 after Close", c.handle)
	}
}

func writeCandidate(t *testing.T, root string, candidate string) string {
	t.Helper()
	path := filepath.Join(root, candidate)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(path, []byte("dll"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	return path
}

func TestFindDLLResolvesFirstExisting(t *testing.T) {
	root := t.TempDir()
	writeCandidate(t, root, dllCandidates[1])
	writeCandidate(t, root, dllCandidates[2])
	got, err := findDLL(root)
	if err != nil {
		t.Fatalf("findDLL error = %v", err)
	}
	if want := filepath.Join(root, dllCandidates[1]); got != want {
		t.Fatalf("findDLL = %q, want %q", got, want)
	}
	writeCandidate(t, root, dllCandidates[0])
	got, err = findDLL(root)
	if err != nil {
		t.Fatalf("findDLL error = %v", err)
	}
	if want := filepath.Join(root, dllCandidates[0]); got != want {
		t.Fatalf("findDLL = %q, want %q", got, want)
	}
}

func TestFindDLLErrorsWhenNoneExist(t *testing.T) {
	root := t.TempDir()
	_, err := findDLL(root)
	if err == nil {
		t.Fatal("findDLL error = nil, want failure")
	}
	if !strings.HasPrefix(err.Error(), "mumu: ") {
		t.Errorf("error %q missing mumu prefix", err.Error())
	}
	if want := fmt.Sprintf("mumu: external_renderer_ipc.dll not found under root %q", root); err.Error() != want {
		t.Errorf("error %q, want %q", err.Error(), want)
	}
}

func TestCaptureRejectsAbsurdDimensions(t *testing.T) {
	for _, dims := range [][2]int32{{100000, 2}, {2, 100000}, {0, 2}, {2, -1}} {
		calls := 0
		c := &Client{
			handle: 1,
			raw: rawAPI{
				capture: func(handle uintptr, displayID int, size int, w *int32, h *int32, buf *byte) int32 {
					calls++
					*w = dims[0]
					*h = dims[1]
					return 0
				},
			},
		}
		if _, err := c.Capture(0); err == nil {
			t.Errorf("Capture(%dx%d) error = nil, want failure", dims[0], dims[1])
		}
		if calls != 1 {
			t.Errorf("Capture(%dx%d) calls = %d, want 1 (no fill call)", dims[0], dims[1], calls)
		}
	}
}

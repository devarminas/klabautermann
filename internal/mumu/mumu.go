package mumu

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const maxDimension = 8192

var dllCandidates = []string{
	"nx_device/15.0/shell/sdk/external_renderer_ipc.dll",
	"nx_device/12.0/shell/sdk/external_renderer_ipc.dll",
	"shell/sdk/external_renderer_ipc.dll",
	"nx_main/sdk/external_renderer_ipc.dll",
}

type rawAPI struct {
	disconnect func(handle uintptr) uintptr
	displayID  func(handle uintptr, pkg *byte, appIndex int) int32
	capture    func(handle uintptr, displayID int, size int, w *int32, h *int32, buf *byte) int32
}

type Client struct {
	handle uintptr
	dll    *syscall.LazyDLL
	raw    rawAPI
}

func findDLL(root string) (string, error) {
	for _, candidate := range dllCandidates {
		path := filepath.Join(root, candidate)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("mumu: external_renderer_ipc.dll not found under root %q", root)
}

func Connect(installRoot string, index int) (*Client, error) {
	path, err := findDLL(installRoot)
	if err != nil {
		return nil, err
	}
	dll := syscall.NewLazyDLL(path)
	names := []string{"nemu_connect", "nemu_disconnect", "nemu_get_display_id", "nemu_capture_display"}
	procs := make(map[string]*syscall.LazyProc, len(names))
	for _, name := range names {
		proc := dll.NewProc(name)
		if err := proc.Find(); err != nil {
			return nil, fmt.Errorf("mumu: resolve %s in %q: %w", name, path, err)
		}
		procs[name] = proc
	}
	pathPtr, err := syscall.UTF16PtrFromString(installRoot)
	if err != nil {
		return nil, fmt.Errorf("mumu: connect to root %q index %d: %w", installRoot, index, err)
	}
	handle, _, _ := syscall.SyscallN(procs["nemu_connect"].Addr(), uintptr(unsafe.Pointer(pathPtr)), uintptr(index))
	if handle == 0 {
		return nil, fmt.Errorf("mumu: connect to root %q index %d failed: zero handle", installRoot, index)
	}
	c := &Client{handle: handle, dll: dll}
	c.raw = rawAPI{
		disconnect: func(handle uintptr) uintptr {
			r, _, _ := syscall.SyscallN(procs["nemu_disconnect"].Addr(), handle)
			return r
		},
		displayID: func(handle uintptr, pkg *byte, appIndex int) int32 {
			r, _, _ := syscall.SyscallN(procs["nemu_get_display_id"].Addr(), handle, uintptr(unsafe.Pointer(pkg)), uintptr(appIndex))
			return int32(r)
		},
		capture: func(handle uintptr, displayID int, size int, w *int32, h *int32, buf *byte) int32 {
			r, _, _ := syscall.SyscallN(procs["nemu_capture_display"].Addr(), handle, uintptr(displayID), uintptr(size), uintptr(unsafe.Pointer(w)), uintptr(unsafe.Pointer(h)), uintptr(unsafe.Pointer(buf)))
			return int32(r)
		},
	}
	return c, nil
}

func (c *Client) Close() {
	if c.handle == 0 {
		return
	}
	c.raw.disconnect(c.handle)
	c.handle = 0
}

func (c *Client) DisplayID(pkg string, appIndex int) (int, error) {
	if pkg == "" {
		return 0, fmt.Errorf("mumu: display id requires a package name, pass a package or use display 0")
	}
	pkgPtr, err := syscall.BytePtrFromString(pkg)
	if err != nil {
		return 0, fmt.Errorf("mumu: display id for package %q index %d: %w", pkg, appIndex, err)
	}
	id := c.raw.displayID(c.handle, pkgPtr, appIndex)
	if id < 0 {
		return 0, fmt.Errorf("mumu: display id for package %q index %d failed: code %d", pkg, appIndex, id)
	}
	return int(id), nil
}

func (c *Client) Capture(displayID int) (image.Image, error) {
	var w, h int32
	if rc := c.raw.capture(c.handle, displayID, 0, &w, &h, nil); rc != 0 {
		return nil, fmt.Errorf("mumu: capture display %d probe failed: code %d", displayID, rc)
	}
	if w <= 0 || h <= 0 || w > maxDimension || h > maxDimension {
		return nil, fmt.Errorf("mumu: capture display %d returned invalid dimensions %dx%d", displayID, w, h)
	}
	buf := make([]byte, int(w)*int(h)*4)
	if rc := c.raw.capture(c.handle, displayID, len(buf), &w, &h, &buf[0]); rc != 0 {
		return nil, fmt.Errorf("mumu: capture display %d failed: code %d", displayID, rc)
	}
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	row := int(w) * 4
	for y := 0; y < int(h); y++ {
		copy(img.Pix[y*img.Stride:(y+1)*img.Stride], buf[(int(h)-1-y)*row:(int(h)-y)*row])
	}
	return img, nil
}

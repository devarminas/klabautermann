package fgo_test

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"klabautermann/internal/fgo"
	"klabautermann/internal/pipeline"
)

func menuScript() *fgo.Script {
	return fgo.New().OpenHome().OpenFormation().CloseView()
}

func questScript() *fgo.Script {
	return fgo.New().OpenChaldeaGate().OpenDailyQuests().CheckTendency().OpenSupport().PickGuest().ConfirmParty().StartQuest().ClassMoonCancer().ClassCaster().ClassPretender().ClassRider().ClassShielder()
}

type decodedRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type decodedNode struct {
	Name      string       `json:"name"`
	Template  string       `json:"template"`
	Threshold float64      `json:"threshold"`
	Tap       string       `json:"tap"`
	ROI       *decodedRect `json:"roi"`
	Next      []string     `json:"next"`
}

type decodedPipeline struct {
	Start string        `json:"start"`
	Nodes []decodedNode `json:"nodes"`
}

func decodeJSON(t *testing.T, data []byte) decodedPipeline {
	t.Helper()
	var decoded decodedPipeline
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("BuildJSON output is not valid JSON: %v", err)
	}
	return decoded
}

func TestMenuTaskFields(t *testing.T) {
	built, err := menuScript().Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if built.Start != "Home" {
		t.Fatalf("Start = %q, want %q", built.Start, "Home")
	}
	wantNodes := map[string]pipeline.Node{
		"Home": {
			Name:      "Home",
			ROI:       image.Rect(1620, 870, 1920, 1040),
			Threshold: 0.8,
			TapCenter: true,
			Next:      []string{"Formation"},
		},
		"Formation": {
			Name:      "Formation",
			ROI:       image.Rect(310, 720, 550, 990),
			TapCenter: true,
			Next:      []string{"Close"},
		},
		"Close": {
			Name:      "Close",
			ROI:       image.Rect(0, 30, 260, 160),
			Threshold: 0.85,
			TapCenter: true,
			Next:      []string{},
		},
	}
	if len(built.Nodes) != len(wantNodes) {
		t.Fatalf("node count = %d, want %d", len(built.Nodes), len(wantNodes))
	}
	for name, want := range wantNodes {
		got, ok := built.Nodes[name]
		if !ok {
			t.Fatalf("missing node %q", name)
		}
		if got.ROI != want.ROI {
			t.Errorf("node %q ROI = %v, want %v", name, got.ROI, want.ROI)
		}
		if got.Threshold != want.Threshold {
			t.Errorf("node %q threshold = %v, want %v", name, got.Threshold, want.Threshold)
		}
		if got.TapCenter != want.TapCenter {
			t.Errorf("node %q TapCenter = %v, want %v", name, got.TapCenter, want.TapCenter)
		}
		if !reflect.DeepEqual(got.Next, want.Next) {
			t.Errorf("node %q next = %v, want %v", name, got.Next, want.Next)
		}
	}
	data, err := menuScript().BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON failed: %v", err)
	}
	decoded := decodeJSON(t, data)
	if decoded.Start != "Home" {
		t.Errorf("JSON start = %q, want %q", decoded.Start, "Home")
	}
	wantJSON := []decodedNode{
		{Name: "Home", Template: "../templates/menu.png", Threshold: 0.8, Tap: "center", ROI: &decodedRect{X: 1620, Y: 870, W: 300, H: 170}, Next: []string{"Formation"}},
		{Name: "Formation", Template: "../templates/formation.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 310, Y: 720, W: 240, H: 270}, Next: []string{"Close"}},
		{Name: "Close", Template: "../templates/close.png", Threshold: 0.85, Tap: "center", ROI: &decodedRect{X: 0, Y: 30, W: 260, H: 130}, Next: []string{}},
	}
	if !reflect.DeepEqual(decoded.Nodes, wantJSON) {
		t.Errorf("JSON nodes = %+v, want %+v", decoded.Nodes, wantJSON)
	}
}

func TestQuestPathFields(t *testing.T) {
	built, err := questScript().Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if built.Start != "ChaldeaGate" {
		t.Fatalf("Start = %q, want %q", built.Start, "ChaldeaGate")
	}
	wantNodes := map[string]pipeline.Node{
		"ChaldeaGate": {
			Name:      "ChaldeaGate",
			ROI:       image.Rect(0, 30, 820, 170),
			Threshold: 0.85,
			TapCenter: true,
			Next:      []string{"DailyQuests"},
		},
		"DailyQuests": {
			Name:      "DailyQuests",
			ROI:       image.Rect(80, 180, 1100, 650),
			TapCenter: true,
		},
		"Tendency": {
			Name:      "Tendency",
			ROI:       image.Rect(900, 200, 1500, 500),
			TapCenter: true,
			Next:      []string{"SelectSupport"},
		},
		"SelectSupport": {
			Name:      "SelectSupport",
			ROI:       image.Rect(300, 20, 1620, 160),
			Threshold: 0.85,
			TapCenter: true,
			Next:      []string{"GuestServant"},
		},
		"GuestServant": {
			Name:      "GuestServant",
			ROI:       image.Rect(100, 200, 1820, 700),
			TapCenter: true,
		},
		"ConfirmParty": {
			Name:      "ConfirmParty",
			ROI:       image.Rect(1200, 0, 1920, 180),
			Threshold: 0.85,
			TapCenter: true,
			Next:      []string{"StartQuest"},
		},
		"StartQuest": {
			Name:      "StartQuest",
			ROI:       image.Rect(1400, 830, 1920, 1080),
			TapCenter: true,
		},
		"ClassMoonCancer": {
			Name:      "ClassMoonCancer",
			ROI:       image.Rect(0, 150, 220, 350),
			TapCenter: true,
		},
		"ClassCaster": {
			Name:      "ClassCaster",
			ROI:       image.Rect(320, 150, 530, 350),
			TapCenter: true,
		},
		"ClassPretender": {
			Name:      "ClassPretender",
			ROI:       image.Rect(590, 150, 800, 350),
			TapCenter: true,
		},
		"ClassRider": {
			Name:      "ClassRider",
			ROI:       image.Rect(920, 150, 1130, 350),
			TapCenter: true,
		},
		"ClassShielder": {
			Name:      "ClassShielder",
			ROI:       image.Rect(1500, 150, 1700, 350),
			TapCenter: true,
		},
	}
	if len(built.Nodes) != len(wantNodes) {
		t.Fatalf("node count = %d, want %d", len(built.Nodes), len(wantNodes))
	}
	for name, want := range wantNodes {
		got, ok := built.Nodes[name]
		if !ok {
			t.Fatalf("missing node %q", name)
		}
		if got.ROI != want.ROI {
			t.Errorf("node %q ROI = %v, want %v", name, got.ROI, want.ROI)
		}
		if got.Threshold != want.Threshold {
			t.Errorf("node %q threshold = %v, want %v", name, got.Threshold, want.Threshold)
		}
		if got.TapCenter != want.TapCenter {
			t.Errorf("node %q TapCenter = %v, want %v", name, got.TapCenter, want.TapCenter)
		}
		if !reflect.DeepEqual(got.Next, want.Next) {
			t.Errorf("node %q next = %v, want %v", name, got.Next, want.Next)
		}
	}
	data, err := questScript().BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON failed: %v", err)
	}
	decoded := decodeJSON(t, data)
	if decoded.Start != "ChaldeaGate" {
		t.Errorf("JSON start = %q, want %q", decoded.Start, "ChaldeaGate")
	}
	wantJSON := []decodedNode{
		{Name: "ChaldeaGate", Template: "../templates/chaldea-gate.png", Threshold: 0.85, Tap: "center", ROI: &decodedRect{X: 0, Y: 30, W: 820, H: 140}, Next: []string{"DailyQuests"}},
		{Name: "DailyQuests", Template: "../templates/daily-quests.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 80, Y: 180, W: 1020, H: 470}, Next: []string{}},
		{Name: "Tendency", Template: "../templates/tendency-saber-rider.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 900, Y: 200, W: 600, H: 300}, Next: []string{"SelectSupport"}},
		{Name: "SelectSupport", Template: "../templates/select-support.png", Threshold: 0.85, Tap: "center", ROI: &decodedRect{X: 300, Y: 20, W: 1320, H: 140}, Next: []string{"GuestServant"}},
		{Name: "GuestServant", Template: "../templates/guest-servant.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 100, Y: 200, W: 1720, H: 500}, Next: []string{}},
		{Name: "ConfirmParty", Template: "../templates/confirm-party.png", Threshold: 0.85, Tap: "center", ROI: &decodedRect{X: 1200, Y: 0, W: 720, H: 180}, Next: []string{"StartQuest"}},
		{Name: "StartQuest", Template: "../templates/start-quest.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 1400, Y: 830, W: 520, H: 250}, Next: []string{}},
		{Name: "ClassMoonCancer", Template: "../templates/icon-mooncancer.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 0, Y: 150, W: 220, H: 200}, Next: []string{}},
		{Name: "ClassCaster", Template: "../templates/icon-caster.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 320, Y: 150, W: 210, H: 200}, Next: []string{}},
		{Name: "ClassPretender", Template: "../templates/icon-pretender.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 590, Y: 150, W: 210, H: 200}, Next: []string{}},
		{Name: "ClassRider", Template: "../templates/icon-rider.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 920, Y: 150, W: 210, H: 200}, Next: []string{}},
		{Name: "ClassShielder", Template: "../templates/icon-shielder.png", Threshold: 0, Tap: "center", ROI: &decodedRect{X: 1500, Y: 150, W: 200, H: 200}, Next: []string{}},
	}
	if !reflect.DeepEqual(decoded.Nodes, wantJSON) {
		t.Errorf("JSON nodes = %+v, want %+v", decoded.Nodes, wantJSON)
	}
}

func writeStubPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{R: 200, G: 50, B: 50, A: 255})
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create stub template: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
		t.Fatalf("encode stub template: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close stub template: %v", err)
	}
}

func TestBuildJSONRoundTripLoads(t *testing.T) {
	data, err := menuScript().BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON failed: %v", err)
	}
	root := t.TempDir()
	taskDir := filepath.Join(root, "task")
	templateDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("create task dir: %v", err)
	}
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("create template dir: %v", err)
	}
	jsonPath := filepath.Join(taskDir, "pipeline.json")
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatalf("write pipeline JSON: %v", err)
	}
	for _, name := range []string{"menu.png", "formation.png", "close.png"} {
		writeStubPNG(t, filepath.Join(templateDir, name))
	}
	loaded, err := pipeline.Load(jsonPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Start != "Home" {
		t.Errorf("Start = %q, want %q", loaded.Start, "Home")
	}
	if len(loaded.Nodes) != 3 {
		t.Fatalf("node count = %d, want 3", len(loaded.Nodes))
	}
	for _, name := range []string{"Home", "Formation", "Close"} {
		node, ok := loaded.Nodes[name]
		if !ok {
			t.Fatalf("missing node %q", name)
		}
		if node.Template == nil {
			t.Errorf("node %q has no decoded template", name)
		}
	}
}

func TestBuildJSONGolden(t *testing.T) {
	data, err := menuScript().BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON failed: %v", err)
	}
	var got any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal BuildJSON output: %v", err)
	}
	raw, err := os.ReadFile("../../resources/pipelines/menu-formation-close.json")
	if err != nil {
		t.Fatalf("read committed pipeline: %v", err)
	}
	var want any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("unmarshal committed pipeline: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildJSON output differs from committed JSON:\n got: %v\nwant: %v", got, want)
	}
}

func TestDuplicateNamesFailBuild(t *testing.T) {
	_, err := fgo.New().
		Tap("Choice", "a.png", image.Rectangle{}).
		Tap("Choice", "b.png", image.Rectangle{}).
		Build()
	if err == nil {
		t.Fatalf("Build succeeded with duplicate names, want error")
	}
	if !strings.Contains(err.Error(), "Choice") {
		t.Errorf("error %q does not name the offending node", err.Error())
	}
}

func TestUnknownNextFailsBuild(t *testing.T) {
	_, err := fgo.New().
		Tap("Lonely", "a.png", image.Rectangle{}, "Missing").
		Build()
	if err == nil {
		t.Fatalf("Build succeeded with unknown next, want error")
	}
	if !strings.Contains(err.Error(), "Lonely") {
		t.Errorf("error %q does not name the offending node", err.Error())
	}
}

func TestFirstNodeIsStart(t *testing.T) {
	built, err := fgo.New().
		Tap("First", "first.png", image.Rect(0, 0, 10, 10), "Home").
		OpenHome().
		OpenFormation().
		CloseView().
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if built.Start != "First" {
		t.Errorf("Start = %q, want %q", built.Start, "First")
	}
	if len(built.Nodes) != 4 {
		t.Errorf("node count = %d, want 4", len(built.Nodes))
	}
}

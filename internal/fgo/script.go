package fgo

import (
	"encoding/json"
	"fmt"
	"image"

	"klabautermann/internal/pipeline"
)

type entry struct {
	name      string
	template  string
	roi       image.Rectangle
	threshold float64
	next      []string
}

type Script struct {
	entries []entry
}

func New() *Script {
	return &Script{}
}

func (s *Script) add(name, template string, roi image.Rectangle, threshold float64, next []string) *Script {
	s.entries = append(s.entries, entry{name: name, template: template, roi: roi, threshold: threshold, next: next})
	return s
}

func (s *Script) Tap(name, template string, roi image.Rectangle, next ...string) *Script {
	return s.add(name, template, roi, 0, next)
}

func (s *Script) OpenHome() *Script {
	return s.add("Home", "../templates/menu.png", image.Rect(1620, 870, 1920, 1040), 0.8, []string{"Formation"})
}

func (s *Script) OpenFormation() *Script {
	return s.add("Formation", "../templates/formation.png", image.Rect(310, 720, 550, 990), 0, []string{"Close"})
}

func (s *Script) CloseView() *Script {
	return s.add("Close", "../templates/close.png", image.Rect(0, 30, 260, 160), 0.85, []string{})
}

func (s *Script) OpenChaldeaGate() *Script {
	return s.add("ChaldeaGate", "../templates/chaldea-gate.png", image.Rect(0, 30, 820, 170), 0.85, []string{"DailyQuests"})
}

func (s *Script) OpenDailyQuests() *Script {
	return s.Tap("DailyQuests", "../templates/daily-quests.png", image.Rect(80, 180, 1100, 650))
}

func (s *Script) CheckTendency() *Script {
	return s.Tap("Tendency", "../templates/tendency-saber-rider.png", image.Rect(900, 200, 1500, 500), "SelectSupport")
}

func (s *Script) OpenSupport() *Script {
	return s.add("SelectSupport", "../templates/select-support.png", image.Rect(300, 20, 1620, 160), 0.85, []string{"GuestServant"})
}

func (s *Script) PickGuest() *Script {
	return s.Tap("GuestServant", "../templates/guest-servant.png", image.Rect(100, 200, 1820, 700))
}

func (s *Script) ConfirmParty() *Script {
	return s.add("ConfirmParty", "../templates/confirm-party.png", image.Rect(1200, 0, 1920, 180), 0.85, []string{"StartQuest"})
}

func (s *Script) StartQuest() *Script {
	return s.Tap("StartQuest", "../templates/start-quest.png", image.Rect(1400, 830, 1920, 1080))
}

func (s *Script) ClassMoonCancer() *Script {
	return s.Tap("ClassMoonCancer", "../templates/icon-mooncancer.png", image.Rect(0, 150, 220, 350))
}

func (s *Script) ClassCaster() *Script {
	return s.Tap("ClassCaster", "../templates/icon-caster.png", image.Rect(320, 150, 530, 350))
}

func (s *Script) ClassPretender() *Script {
	return s.Tap("ClassPretender", "../templates/icon-pretender.png", image.Rect(590, 150, 800, 350))
}

func (s *Script) ClassRider() *Script {
	return s.Tap("ClassRider", "../templates/icon-rider.png", image.Rect(920, 150, 1130, 350))
}

func (s *Script) ClassShielder() *Script {
	return s.Tap("ClassShielder", "../templates/icon-shielder.png", image.Rect(1500, 150, 1700, 350))
}

func (s *Script) Build() (pipeline.Pipeline, error) {
	if len(s.entries) == 0 {
		return pipeline.Pipeline{}, fmt.Errorf("fgo: script has no nodes")
	}
	nodes := make(map[string]pipeline.Node, len(s.entries))
	for _, e := range s.entries {
		if _, dup := nodes[e.name]; dup {
			return pipeline.Pipeline{}, fmt.Errorf("fgo: duplicate node %q", e.name)
		}
		nodes[e.name] = pipeline.Node{
			Name:      e.name,
			ROI:       e.roi,
			Threshold: e.threshold,
			TapCenter: true,
			Next:      e.next,
		}
	}
	for _, e := range s.entries {
		for _, name := range e.next {
			if _, ok := nodes[name]; !ok {
				return pipeline.Pipeline{}, fmt.Errorf("fgo: node %q next %q names unknown node", e.name, name)
			}
		}
		if onError := nodes[e.name].OnError; onError != "" {
			if _, ok := nodes[onError]; !ok {
				return pipeline.Pipeline{}, fmt.Errorf("fgo: node %q on_error %q names unknown node", e.name, onError)
			}
		}
	}
	return pipeline.Pipeline{Start: s.entries[0].name, Nodes: nodes}, nil
}

type jsonRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type jsonNode struct {
	Name      string    `json:"name"`
	Template  string    `json:"template"`
	Threshold float64   `json:"threshold,omitempty"`
	Tap       string    `json:"tap"`
	ROI       *jsonRect `json:"roi,omitempty"`
	Next      []string  `json:"next"`
	OnError   string    `json:"on_error,omitempty"`
}

type jsonPipeline struct {
	Start string     `json:"start"`
	Nodes []jsonNode `json:"nodes"`
}

func (s *Script) BuildJSON() ([]byte, error) {
	if _, err := s.Build(); err != nil {
		return nil, err
	}
	out := jsonPipeline{Start: s.entries[0].name}
	for _, e := range s.entries {
		node := jsonNode{Name: e.name, Template: e.template, Threshold: e.threshold, Tap: "center", Next: e.next}
		if node.Next == nil {
			node.Next = []string{}
		}
		if !e.roi.Empty() {
			node.ROI = &jsonRect{X: e.roi.Min.X, Y: e.roi.Min.Y, W: e.roi.Dx(), H: e.roi.Dy()}
		}
		out.Nodes = append(out.Nodes, node)
	}
	return json.MarshalIndent(out, "", "  ")
}

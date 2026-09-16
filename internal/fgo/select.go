package fgo

var nodeClasses = map[string]Class{
	"ClassMoonCancer": MoonCancer,
	"ClassCaster":     Caster,
	"ClassPretender":  Pretender,
	"ClassRider":      Rider,
	"ClassShielder":   Shielder,
	"EnemyRider":      Rider,
	"EnemySaber":      Saber,
}

func ClassOfNode(name string) (Class, bool) {
	c, ok := nodeClasses[name]
	return c, ok
}

func BestAlly(enemies []Class, allies []Class) (Class, float64) {
	if len(allies) == 0 {
		return "", 0
	}
	best := allies[0]
	var bestScore float64
	for _, e := range enemies {
		bestScore += Advantage(best, e)
	}
	for _, a := range allies[1:] {
		var s float64
		for _, e := range enemies {
			s += Advantage(a, e)
		}
		if s > bestScore {
			best, bestScore = a, s
		}
	}
	return best, bestScore
}

func WeakTo(enemy Class, allies []Class) []Class {
	var out []Class
	for _, a := range allies {
		if Advantage(a, enemy) >= 2.0 {
			out = append(out, a)
		}
	}
	return out
}

package fgo

import (
	"context"
	"fmt"
)

type CardKind string

const (
	Buster CardKind = "Buster"
	Arts   CardKind = "Arts"
	Quick  CardKind = "Quick"
)

type Card struct {
	Kind    CardKind
	Servant int
	X, Y    int
}

type Combat struct {
	MaxTurns int
}

func PreferBuster(cards []Card) []Card {
	out := make([]Card, 0, 3)
	for _, kind := range []CardKind{Buster, Arts, Quick} {
		for _, cd := range cards {
			if cd.Kind != kind {
				continue
			}
			out = append(out, cd)
			if len(out) == 3 {
				return out
			}
		}
	}
	return out
}

func OrderFocus(cards []Card) []Card {
	var order []int
	seen := map[int]bool{}
	for _, cd := range cards {
		if !seen[cd.Servant] {
			seen[cd.Servant] = true
			order = append(order, cd.Servant)
		}
	}
	out := make([]Card, 0, 3)
	for _, s := range order {
		for _, cd := range cards {
			if cd.Servant != s {
				continue
			}
			out = append(out, cd)
			if len(out) == 3 {
				return out
			}
		}
	}
	return out
}

func OrderSpread(cards []Card) []Card {
	var order []int
	groups := map[int][]Card{}
	for _, cd := range cards {
		if _, ok := groups[cd.Servant]; !ok {
			order = append(order, cd.Servant)
		}
		groups[cd.Servant] = append(groups[cd.Servant], cd)
	}
	out := make([]Card, 0, 3)
	for i := 0; ; i++ {
		progressed := false
		for _, s := range order {
			if i < len(groups[s]) {
				out = append(out, groups[s][i])
				progressed = true
				if len(out) == 3 {
					return out
				}
			}
		}
		if !progressed {
			return out
		}
	}
}

func (c *Combat) Run(ctx context.Context, sense func(ctx context.Context) (clear bool, cards []Card, err error), pick func([]Card) []Card, attack func(ctx context.Context) error, tap func(ctx context.Context, x, y int) error, focus func(ctx context.Context) (x, y int, tapTarget bool, err error)) (turns int, err error) {
	max := 30
	if c != nil && c.MaxTurns > 0 {
		max = c.MaxTurns
	}
	choose := pick
	if choose == nil {
		choose = PreferBuster
	}
	for {
		if err := ctx.Err(); err != nil {
			return turns, fmt.Errorf("turn %d: %w", turns+1, err)
		}
		clear, _, err := sense(ctx)
		if err != nil {
			return turns, fmt.Errorf("turn %d: sense: %w", turns+1, err)
		}
		if clear {
			return turns, nil
		}
		if turns >= max {
			return turns, fmt.Errorf("turn %d: max turns %d exceeded", turns+1, max)
		}
		if err := attack(ctx); err != nil {
			return turns, fmt.Errorf("turn %d: attack: %w", turns+1, err)
		}
		if err := ctx.Err(); err != nil {
			return turns, fmt.Errorf("turn %d: %w", turns+1, err)
		}
		// Attack redeals the hand, so only cards sensed after it are tappable.
		_, dealt, err := sense(ctx)
		if err != nil {
			return turns, fmt.Errorf("turn %d: sense: %w", turns+1, err)
		}
		picked := choose(dealt)
		if focus != nil {
			fx, fy, ok, err := focus(ctx)
			if err != nil {
				return turns, fmt.Errorf("turn %d: focus: %w", turns+1, err)
			}
			if ok {
				if err := tap(ctx, fx, fy); err != nil {
					return turns, fmt.Errorf("turn %d: tap %d,%d: %w", turns+1, fx, fy, err)
				}
			}
		}
		for _, cd := range picked {
			if err := ctx.Err(); err != nil {
				return turns, fmt.Errorf("turn %d: %w", turns+1, err)
			}
			if err := tap(ctx, cd.X, cd.Y); err != nil {
				return turns, fmt.Errorf("turn %d: tap %d,%d: %w", turns+1, cd.X, cd.Y, err)
			}
		}
		turns++
	}
}

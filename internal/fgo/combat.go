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
	Kind CardKind
	X, Y int
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

func (c *Combat) Run(ctx context.Context, sense func(ctx context.Context) (clear bool, cards []Card, err error), pick func([]Card) []Card, attack func(ctx context.Context) error, tap func(ctx context.Context, x, y int) error) (turns int, err error) {
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
		for _, cd := range choose(dealt) {
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

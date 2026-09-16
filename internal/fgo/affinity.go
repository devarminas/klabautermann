package fgo

type Class string

const (
	Shielder   Class = "Shielder"
	Saber      Class = "Saber"
	Archer     Class = "Archer"
	Lancer     Class = "Lancer"
	Rider      Class = "Rider"
	Caster     Class = "Caster"
	Assassin   Class = "Assassin"
	Berserker  Class = "Berserker"
	Ruler      Class = "Ruler"
	Avenger    Class = "Avenger"
	MoonCancer Class = "MoonCancer"
	AlterEgo   Class = "AlterEgo"
	Foreigner  Class = "Foreigner"
	Pretender  Class = "Pretender"
	Beast      Class = "Beast"
)

func (c Class) Valid() bool {
	_, ok := classBases[c]
	return ok
}

func Advantage(atk, def Class) float64 {
	if !atk.Valid() || !def.Valid() {
		return 1.0
	}
	if v, ok := affinities[atk][def]; ok {
		return v
	}
	return 1.0
}

func ClassBase(c Class) float64 {
	if v, ok := classBases[c]; ok {
		return v
	}
	return 1.0
}

var classBases = map[Class]float64{
	Saber:      1.0,
	Archer:     0.95,
	Lancer:     1.05,
	Rider:      1.0,
	Caster:     0.9,
	Assassin:   0.9,
	Berserker:  1.1,
	Shielder:   1.0,
	Ruler:      1.1,
	Avenger:    1.1,
	MoonCancer: 1.0,
	AlterEgo:   1.0,
	Foreigner:  1.0,
	Pretender:  1.0,
	Beast:      1.0,
}

var affinities = map[Class]map[Class]float64{
	Saber:      {Archer: 0.5, Lancer: 2.0, Berserker: 2.0, Ruler: 0.5},
	Archer:     {Saber: 2.0, Lancer: 0.5, Berserker: 2.0, Ruler: 0.5},
	Lancer:     {Saber: 0.5, Archer: 2.0, Berserker: 2.0, Ruler: 0.5},
	Rider:      {Caster: 2.0, Assassin: 0.5, Berserker: 2.0, Ruler: 0.5},
	Caster:     {Caster: 0.5, Assassin: 2.0, Berserker: 2.0, Ruler: 0.5},
	Assassin:   {Rider: 2.0, Caster: 0.5, Berserker: 2.0, Ruler: 0.5},
	Berserker:  {Saber: 1.5, Archer: 1.5, Lancer: 1.5, Rider: 1.5, Caster: 1.5, Assassin: 1.5, Berserker: 1.5, Ruler: 1.5, Avenger: 1.5, MoonCancer: 1.5, AlterEgo: 0.5, Foreigner: 1.5, Pretender: 0.5, Beast: 1.5},
	Ruler:      {Berserker: 2.0, Avenger: 0.5, MoonCancer: 2.0},
	Avenger:    {Ruler: 2.0, MoonCancer: 0.5, Berserker: 2.0},
	MoonCancer: {Ruler: 0.5, Avenger: 2.0, Berserker: 2.0},
	AlterEgo:   {Saber: 0.5, Archer: 0.5, Lancer: 0.5, Rider: 1.5, Caster: 1.5, Assassin: 1.5, Berserker: 2.0, Foreigner: 2.0, Pretender: 0.5},
	Foreigner:  {AlterEgo: 0.5, Foreigner: 2.0, Pretender: 2.0, Berserker: 2.0},
	Pretender:  {Saber: 1.5, Archer: 1.5, Lancer: 1.5, Rider: 0.5, Caster: 0.5, Assassin: 0.5, Berserker: 2.0, AlterEgo: 2.0, Foreigner: 0.5},
	// Beast uses Beast I values because Beast matchups vary per Beast, source is the FGO wiki Battle Damage Calculation class triangle table.
	Beast: {Saber: 2.0, Archer: 2.0, Lancer: 2.0, Berserker: 2.0, Avenger: 0.5},
}

package confucius

import "github.com/SlothNinja/game"

type Entry struct {
	*game.Entry
}

func (g *Game) newEntry() *Entry {
	e := new(Entry)
	e.Entry = game.NewEntry(g)
	return e
}

func (p *Player) newEntry() *Entry {
	e := new(Entry)
	g := p.Game()
	e.Entry = game.NewEntryFor(p, g)
	return e
}

func (e *Entry) PhaseName() string {
	return PhaseNames[e.Phase()]
}

func pluralize(label string, value int) string {
	if value != 1 {
		return label + "s"
	}
	return label
}

package confucius

import (
	"encoding/gob"
	"fmt"
	"html/template"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.invasionEntry", new(invasionEntry))
}

func (g *Game) invasionPhase(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = Invasion
	for i, land := range g.ForeignLands {
		switch {
		case land.Resolved:
		case land.AllBoxesOccupied():
			g.successfulInvasionOf(i)
		case g.Wall >= 4 && i == 0, g.Wall >= 6 && i == 1, g.Wall >= 8 && i == 2:
			g.unsuccessfulInvasionOf(i)
		}
	}
}

func (g *Game) successfulInvasionOf(landIndex int) {
	land := g.ForeignLands[landIndex]
	entry := g.unsuccessfulInvasionOf(landIndex)
	for _, box := range land.Boxes {
		p := box.Player()
		p.Score += box.Points
	}
	entry.Successful = true
}

func (g *Game) unsuccessfulInvasionOf(landIndex int) *invasionEntry {
	land := g.ForeignLands[landIndex]
	entry := g.newInvasionEntry()
	entry.ForeignLand = land
	for _, box := range land.Boxes {
		p := box.Player()
		if p != nil && box.AwardCard && len(g.EmperorDeck) > 0 {
			card := g.EmperorDeck.Draw()
			p.EmperorHand.Append(card)
			p.EmperorHand.Reveal()
			entry.AwardCard = true
		}
	}
	land.Resolved = true
	return entry
}

type invasionEntry struct {
	*Entry
	ForeignLand *ForeignLand
	Successful  bool
	AwardCard   bool
}

func (g *Game) newInvasionEntry() *invasionEntry {
	e := new(invasionEntry)
	e.Entry = g.newEntry()
	g.Log = append(g.Log, e)
	return e
}

func (e *invasionEntry) HTML() template.HTML {
	g := e.Game().(*Game)
	var s string
	if e.Successful {
		s = fmt.Sprintf("<div>The invasion of %s succeeded.</div>", e.ForeignLand.Name())
		for _, box := range e.ForeignLand.Boxes {
			player := g.PlayerByID(box.PlayerID)
			s += fmt.Sprintf("<div>%s received %d points.</div>", g.NameFor(player), box.Points)
			if e.AwardCard && box.AwardCard {
				s += fmt.Sprintf("<div>%s awarded an Emperor's Reward card.</div>", g.NameFor(player))
			}
		}
		return restful.HTML(s)
	}
	s = fmt.Sprintf("<div>The invasion of %s failed.</div>", e.ForeignLand.Name())
	if e.AwardCard {
		for _, box := range e.ForeignLand.Boxes {
			if box.AwardCard {
				player := e.Game().(*Game).PlayerByID(box.PlayerID)
				s += fmt.Sprintf("<div>%s awarded an Emperor's Reward card.</div>", g.NameFor(player))
			}
		}
	}
	return restful.HTML(s)
}

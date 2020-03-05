package confucius

import (
	"encoding/gob"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.scoreChiefMinisterEntry", new(scoreChiefMinisterEntry))
	gob.RegisterName("*game.scoreAdmiralEntry", new(scoreAdmiralEntry))
	gob.RegisterName("*game.scoreGeneralEntry", new(scoreGeneralEntry))
	gob.RegisterName("*game.announceWinnersEntry", new(announceWinnersEntry))
}

func (g *Game) endOfRoundPhase(c *gin.Context) (cs contest.Contests) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = EndOfRound
	g.placeNewOfficialsPhase(c)
	if completed := g.discardPhase(c); completed {
		g.returnActionCubesPhase(c)
		cs = g.endOfGamePhase(c)
	}
	return
}

func (g *Game) placeNewOfficialsPhase(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	for _, m := range g.Ministries {
		g.placeNewOfficialIn(m)
	}
}

func (g *Game) placeNewOfficialIn(m *Ministry) {
	for _, s := range []Seniority{1, 2, 6, 7} {
		if _, ok := m.Officials[s]; !ok {
			o := g.OfficialsDeck.Draw()
			o.Seniority = s
			m.Officials[s] = o
			return
		}
	}
}

func (g *Game) newRoundPhase(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Entering")

	g.Round += 1
	for _, p := range g.Players() {
		p.TakenCommercial = false
	}
}

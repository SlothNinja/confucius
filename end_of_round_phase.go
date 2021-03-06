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

func (client *Client) endOfRoundPhase(c *gin.Context, g *Game) ([]*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	g.Phase = EndOfRound
	g.placeNewOfficialsPhase(c)
	completed := g.discardPhase(c)
	if completed {
		g.returnActionCubesPhase(c)
		return client.endOfGamePhase(c, g)
	}
	return nil, nil
}

func (g *Game) placeNewOfficialsPhase(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

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

func (g *Game) newRoundPhase() {
	log.Debugf(msgEnter)
	defer log.Debugf(msgEnter)

	g.Round += 1
	for _, p := range g.Players() {
		p.TakenCommercial = false
	}
}

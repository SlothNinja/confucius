package confucius

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.forceExamEntry", new(forceExamEntry))
}

func (g *Game) forceExam(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cards, cubes, err := g.validateForceExam(c, cu)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Move played cards from hand to discard pile
	cp.ConCardHand.Remove(cards...)
	g.ConDiscardPile.Append(cards...)

	// Place Action Cubes
	cp.PlaceCubesIn(ForceSpace, cubes)

	// Create Action Object for logging
	e := cp.newForceExamEntry(cards)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type forceExamEntry struct {
	*Entry
	Played ConCards
}

func (p *Player) newForceExamEntry(cards ConCards) *forceExamEntry {
	g := p.Game()
	e := new(forceExamEntry)
	e.Entry = p.newEntry()
	e.Played = cards
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *forceExamEntry) HTML() template.HTML {
	length := len(e.Played)
	return restful.HTML("%s spent %d %s having %d coins to force an examination.",
		e.Player().Name(), length, pluralize("card", length), e.Played.Coins())
}

func (g *Game) validateForceExam(c *gin.Context, cu *user.User) (ConCards, int, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, 0, err
	}

	cards, err := g.getConCards(c, "force-exam")
	if err != nil {
		return nil, 0, err
	}

	coinValue := cards.Coins()
	cp := g.CurrentPlayer()

	switch {
	case g.Round == 1:
		return nil, 0, sn.NewVError("You cannot force an examination during round %d.", g.Round)
	case !cp.canAffordForceExam():
		return nil, 0, sn.NewVError("You selected cards having %d total coins, but you need 2 coins to force an examination.", coinValue)
	}
	return cards, cubes, nil
}

func (g *Game) EnableForceExam(cu *user.User) bool {
	cp := g.CurrentPlayer()
	return g.inActionsOrImperialFavourPhase() && cp != nil && g.IsCurrentPlayer(cu) &&
		g.Round > 1 && !cp.PerformedAction && cp.hasEnoughCubesFor(ForceSpace) && cp.canAffordForceExam()
}

func (p *Player) canAffordForceExam() bool {
	return p.ConCardHand.Coins() >= 2
}

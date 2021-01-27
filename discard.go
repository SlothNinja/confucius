package confucius

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.discardEntry", new(discardEntry))
}

func (g *Game) discardPhase(c *gin.Context) bool {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = Discard
	g.beginningOfPhaseReset()

	ps := make(game.Playerers, 0)
	for _, p := range g.Players() {
		log.Debugf("ConCardHand: %#v", p.ConCardHand)
		log.Debugf("len(ConCardHand): %#v", len(p.ConCardHand))
		if len(p.ConCardHand) > 4 {
			ps = append(ps, p)
		}
	}

	g.SetCurrentPlayerers(ps...)
	return len(ps) == 0
}

func (g *Game) discard(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cards, err := g.validateDiscard(c, cu)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.discard(cards...)

	// Set flash message
	restful.AddNoticef(c, string(cp.newDiscardEntry(cards...).HTML()))
	return "", game.Cache, nil
}

func (p *Player) discard(cards ...*ConCard) {
	p.PerformedAction = true

	// Move played cards from hand to discard pile
	p.ConCardHand.Remove(cards...)
	p.Game().ConDiscardPile.Append(cards...)
}

type discardEntry struct {
	*Entry
	Discarded ConCards
}

func (p *Player) newDiscardEntry(cards ...*ConCard) *discardEntry {
	g := p.Game()
	e := new(discardEntry)
	e.Entry = p.newEntry()
	e.Discarded = cards
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *discardEntry) HTML() template.HTML {
	return restful.HTML("%s discarded %d cards.", e.Player().Name(), len(e.Discarded))
}

func (g *Game) validateDiscard(c *gin.Context, cu *user.User) (ConCards, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cards, err := g.getConCards(c, "discard")
	if err != nil {
		return nil, err
	}

	cp := g.CurrentPlayer()
	newHandCount := len(cp.ConCardHand) - len(cards)
	switch {
	case !g.IsCurrentPlayer(cu):
		return nil, sn.NewVError("Only a current player may discard cards.")
	case g.Phase != Discard:
		return nil, sn.NewVError("You cannot discard cards during the %s phase.", g.PhaseName())
	case newHandCount != 4:
		return nil, sn.NewVError("You must discard down to 4 cards.  You have discarded to %d cards.",
			newHandCount)
	default:
		return cards, nil
	}
}

func (g *Game) EnableDiscard(cu *user.User) bool {
	return g.IsCurrentPlayer(cu) && g.Phase == Discard && g.CurrentPlayer() != nil &&
		!g.CurrentPlayer().PerformedAction
}

func (client *Client) discardPhaseFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	cp := g.CurrentPlayer()
	g.RemoveCurrentPlayers(cp)

	if len(g.CurrentPlayerers()) == 0 {
		g.returnActionCubesPhase(c)
		cs, err := client.endOfGamePhase(c, g)
		if err != nil {
			return nil, nil, err
		}
		return s, cs, nil
	}
	return s, nil, nil
}

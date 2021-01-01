package confucius

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"strconv"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.buyJunksEntry", new(buyJunksEntry))
}

func (g *Game) buyJunks(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// Get Junks and Cards
	js, cds, cbs, err := g.validateBuyJunks(c, cu)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cubes
	cp.PlaceCubesIn(JunksVoyageSpace, cbs)

	// Give Player Junks
	g.Junks -= js
	cp.Junks += js

	// Move played cards from hand to discard pile
	cp.ConCardHand.Remove(cds...)
	g.ConDiscardPile.Append(cds...)

	// Create Action Object for logging
	entry := cp.newBuyJunksEntry(js, cds)

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	return "", game.Cache, nil
}

type buyJunksEntry struct {
	*Entry
	Junks  int
	Played ConCards
}

func (p *Player) newBuyJunksEntry(js int, played ConCards) *buyJunksEntry {
	g := p.Game()
	e := new(buyJunksEntry)
	e.Entry = p.newEntry()
	e.Junks = js
	e.Played = played
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *buyJunksEntry) HTML() template.HTML {
	length := len(e.Played)
	coins := e.Played.Coins()
	return restful.HTML("%s spent %d Confucius %s having %d %s to buy %d %s.",
		e.Player().Name(), length, pluralize("card", length), coins, pluralize("coin", coins), e.Junks, pluralize("junk", e.Junks))
}

func (g *Game) validateBuyJunks(c *gin.Context, cu *user.User) (int, ConCards, int, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cbs, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return 0, nil, 0, err
	}

	cds, err := g.getConCards(c, "buy-junks")
	if err != nil {
		return 0, nil, 0, err
	}

	js, err := strconv.Atoi(c.PostForm("junks"))
	if err != nil {
		return 0, nil, 0, fmt.Errorf(`Form value for "junks" is invalid.`)
	}

	cp := g.CurrentPlayer()
	cv := cds.Coins()
	cost := cp.junkCostFor(js)

	switch {
	case cv < cost:
		return 0, nil, 0, sn.NewVError("You selected cards having %d total coins, but you need %d coins to buy the selected junks.", cv, cost)
	case js > g.Junks:
		return 0, nil, 0, sn.NewVError("You selected more junks than there are available in stock.")
	default:
		return js, cds, cbs, err
	}
}

func (g *Game) EnableBuyJunks(cu *user.User) bool {
	return g.IsCurrentPlayer(cu) && g.CurrentPlayer().canBuyJunks()
}

func (p *Player) canBuyJunks() bool {
	g := p.Game()
	return g.inActionsOrImperialFavourPhase() && !p.PerformedAction &&
		p.hasEnoughCubesFor(JunksVoyageSpace) && p.canAffordJunk()
}

func (p *Player) canAffordJunk() bool {
	return p.ConCardHand.Coins() >= p.junkCostFor(1)
}

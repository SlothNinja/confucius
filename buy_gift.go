package confucius

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.buyGiftEntry", new(buyGiftEntry))
}

func (g *Game) buyGift(c *gin.Context) (tmpl string, a game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// Get Cards and Gift
	cds, gc, cbs, err := g.validateBuyGift(c)
	if err != nil {
		a = game.None
		return
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cube(s) In BuyGiftSpace
	cp.PlaceCubesIn(BuyGiftSpace, cbs)

	// Remove Gift From GiftCardHand
	cp.GiftCardHand.Remove(gc)

	// Move played cards from hand to discard pile
	cp.ConCardHand.Remove(cds...)
	g.ConDiscardPile.Append(cds...)

	// Place Gift With Those Bought
	cp.GiftsBought.Append(gc)

	// Create Action Object for logging
	entry := cp.newBuyGiftEntry(gc, cds)

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	a = game.Cache
	return
}

type buyGiftEntry struct {
	*Entry
	Gift   *GiftCard
	Played ConCards
}

func (p *Player) newBuyGiftEntry(gc *GiftCard, played ConCards) *buyGiftEntry {
	g := p.Game()
	e := new(buyGiftEntry)
	e.Entry = p.newEntry()
	e.Gift = gc
	e.Played = played
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *buyGiftEntry) HTML() template.HTML {
	length := len(e.Played)
	return restful.HTML("%s used %d %s to buy %s gift for %d coins.",
		e.Player().Name(), length, pluralize("card", length), e.Gift.Name(), e.Gift.Value)
}

func (g *Game) validateBuyGift(c *gin.Context) (cds ConCards, gc *GiftCard, cbs int, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if cbs, err = g.validatePlayerAction(c); err != nil {
		return
	}

	if cds, err = g.getConCards(c, "buy-gift"); err != nil {
		return
	}

	var gv GiftCardValue
	if gv, err = g.getGiftValue(c, "buy-gift"); err != nil {
		return
	}

	cp := g.CurrentPlayer()
	cv := cds.Coins()
	gc = cp.GetGift(gv)

	switch {
	case gc == nil:
		err = sn.NewVError("You don't have a gift of value %d to buy.", gv)
	case cv < gc.Cost():
		err = sn.NewVError("You selected cards having %d total coins, but the %s gift costs %d coins.", cv, gc.Name(), gc.Value)
	}
	return
}

func (g *Game) EnableBuyGift(c *gin.Context) bool {
	return g.CUserIsCPlayerOrAdmin(c) && g.CurrentPlayer().canBuyGift()
}

func (p *Player) canBuyGift() bool {
	g := p.Game()
	return g.inActionsOrImperialFavourPhase() && !p.PerformedAction &&
		p.hasEnoughCubesFor(BuyGiftSpace) && p.canAffordGift()
}

func (p *Player) canAffordGift() (b bool) {
	coins := p.ConCardHand.Coins()
	for _, gc := range p.GiftCardHand {
		if b = coins >= gc.Cost(); b {
			return
		}
	}
	return
}

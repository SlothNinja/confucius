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
	gob.RegisterName("*game.giveGiftEntry", new(giveGiftEntry))
}

func (g *Game) giveGift(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	// Get Recipient and Gift
	recipient, gift, cubes, err := g.validateGiveGift(c, cu)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cube(s) In GiveGiftSpace
	cp.PlaceCubesIn(GiveGiftSpace, cubes)

	// Give Gift
	canceledGift := cp.GiveGiftTo(gift, recipient)
	cp.GiftCardHand.Remove(gift)

	// Create Action Object for logging
	entry := cp.newGiveGiftEntry(recipient, gift, canceledGift)

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	return "", game.Cache, nil
}

type giveGiftEntry struct {
	*Entry
	Gift         *GiftCard
	CanceledGift bool
}

func (p *Player) newGiveGiftEntry(op *Player, gift *GiftCard, canceled bool) *giveGiftEntry {
	g := p.Game()
	e := new(giveGiftEntry)
	e.Entry = p.newEntry()
	e.Gift = gift
	e.CanceledGift = canceled
	e.SetOtherPlayer(op)
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *giveGiftEntry) HTML() template.HTML {
	if !e.CanceledGift {
		return restful.HTML("%s gave value %d gift (%s) to %s.",
			e.Player().Name(), e.Gift.Value, e.Gift.Name(), e.OtherPlayer().Name())
	}
	return restful.HTML("%s gave value %d gift (%s) to %s and canceled gift from %s.",
		e.Player().Name(), e.Gift.Value, e.Gift.Name(), e.OtherPlayer().Name(), e.OtherPlayer().Name())
}

func (g *Game) validateGiveGift(c *gin.Context, cu *user.User) (*Player, *GiftCard, int, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, nil, 0, err
	}

	recipient, err := g.getPlayer(c, "give-gift-player")
	if err != nil {
		return nil, nil, 0, err
	}

	giftValue, err := g.getGiftValue(c, "give-gift")
	if err != nil {
		return nil, nil, 0, err
	}

	cp := g.CurrentPlayer()
	oldGift := recipient.giftFrom(cp)
	givenGift := cp.GetBoughtGift(giftValue)
	receivedGift := cp.giftFrom(recipient)

	switch {
	case recipient == nil:
		return nil, nil, 0, sn.NewVError("Recipient not found.")
	case givenGift == nil:
		return nil, nil, 0, sn.NewVError("You don't have a gift of value %d to give.", giftValue)
	case oldGift != nil && oldGift.Value > givenGift.Value:
		return nil, nil, 0, sn.NewVError("You must give a gift that is greater than your present gift to the player.")
	case cp.Equal(recipient):
		return nil, nil, 0, sn.NewVError("You can't give yourself a gift.")
	case receivedGift != nil && receivedGift.Value > givenGift.Value:
		return nil, nil, 0, sn.NewVError("You must give a gift that is greater than or equal to the gift the player gave you.")
	}
	return recipient, givenGift, cubes, nil
}

func (g *Game) EnableGiveGift(cu *user.User) bool {
	cp := g.CurrentPlayer()
	requiredCubes := cp.RequiredCubesFor(GiveGiftSpace)
	return g.IsCurrentPlayer(cu) && cp.ActionCubes >= requiredCubes && len(cp.GiftsBought) >= 1
}

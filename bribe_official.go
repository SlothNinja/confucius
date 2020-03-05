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
	gob.RegisterName("*game.bribeOfficialEntry", new(bribeOfficialEntry))
}

func (g *Game) bribeOfficial(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cards, ministry, official, cubes, err := g.validateBribeOfficial(c)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cubes
	cp.PlaceCubesIn(BribeSecureSpace, cubes)

	// Place Marker On Official
	official.setPlayer(cp)

	// Move played cards from hand to discard pile
	cp.ConCardHand.Remove(cards...)
	g.ConDiscardPile.Append(cards...)

	// Create Action Object for logging
	entry := cp.newBribeOfficialEntry(ministry, official, cards)

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	return "", game.Cache, nil
}

type bribeOfficialEntry struct {
	*Entry
	MinistryName string
	Seniority    Seniority
	Played       ConCards
}

func (p *Player) newBribeOfficialEntry(m *Ministry, o *OfficialTile, cs ConCards) *bribeOfficialEntry {
	g := p.Game()
	e := new(bribeOfficialEntry)
	e.Entry = p.newEntry()
	e.MinistryName = m.Name()
	e.Seniority = o.Seniority
	e.Played = cs
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (g *bribeOfficialEntry) HTML() template.HTML {
	length := len(g.Played)
	return restful.HTML("%s spent %d %s having %d coins to bribe %s official with level %d seniority.",
		g.Player().Name(), length, pluralize("card", length), g.Played.Coins(), g.MinistryName, g.Seniority)
}

func (g *Game) validateBribeOfficial(c *gin.Context) (cds ConCards, m *Ministry, o *OfficialTile, cbs int, err error) {
	if cbs, err = g.validatePlayerAction(c); err != nil {
		return
	}

	if cds, err = g.getConCards(c, "bribe-official"); err != nil {
		return
	}

	if m, o, err = g.getMinistryAndOfficial(c, "bribe-official"); err != nil {
		return
	}

	cp := g.CurrentPlayer()
	switch gp := cp.hasGiftObligationIn(m); {
	case gp != nil:
		err = sn.NewVError("You have a gift obligation to %s that prevents you from bribing another official in the %s ministry.", g.NameFor(gp), m.Name())
	case o.Bribed():
		err = sn.NewVError("You can't bribe an official that already has a marker.")
	case cds.Coins() < o.CostFor(cp):
		err = sn.NewVError("You selected cards having %d total coins, but you need %d coins to bribe the selected official.", cds.Coins(), cp.CostFor(o))
	}
	return
}

func (p *Player) hasGiftObligationIn(m *Ministry) (ret *Player) {
	g := p.Game()
	for _, p2 := range g.Players() {
		if p2inf := p2.influenceIn(m); p.hasGiftFrom(p2) && p2inf > 0 && p.influenceIn(m) >= p2inf {
			ret = p2
			return
		}
	}
	return
}

func (g *Game) EnableBribeOfficial(c *gin.Context) bool {
	cp := g.CurrentPlayer()
	return g.CUserIsCPlayerOrAdmin(c) && cp.canBribeAnOfficial()
}

func (p *Player) canBribeAnOfficial() bool {
	g := p.Game()
	return g.inActionsOrImperialFavourPhase() && !p.PerformedAction && p.hasEnoughCubesFor(BribeSecureSpace) &&
		g.hasBribableOfficialFor(p)
}

func (g *Game) hasBribableOfficialFor(p *Player) bool {
	for _, m := range g.Ministries {
		if !m.Resolved {
			for _, o := range m.Officials {
				if o.NotBribed() && p.canAffordToBribe(o) {
					return true
				}
			}
		}
	}
	return false
}

func (p *Player) canAffordToBribe(o *OfficialTile) bool {
	return p.ConCardHand.Coins() >= o.CostFor(p)
}

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
	gob.RegisterName("*game.bribeOfficialEntry", new(bribeOfficialEntry))
}

func (g *Game) bribeOfficial(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cards, ministry, official, cubes, err := g.validateBribeOfficial(c, cu)
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

func (g *Game) validateBribeOfficial(c *gin.Context, cu *user.User) (ConCards, *Ministry, *OfficialTile, int, error) {
	cbs, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	cds, err := g.getConCards(c, "bribe-official")
	if err != nil {
		return nil, nil, nil, 0, err
	}

	m, o, err := g.getMinistryAndOfficial(c, "bribe-official")
	if err != nil {
		return nil, nil, nil, 0, err
	}

	cp := g.CurrentPlayer()
	gp := cp.hasGiftObligationIn(m)

	switch {
	case gp != nil:
		return nil, nil, nil, 0, sn.NewVError("You have a gift obligation to %s that prevents you from bribing another official in the %s ministry.", g.NameFor(gp), m.Name())
	case o.Bribed():
		return nil, nil, nil, 0, sn.NewVError("You can't bribe an official that already has a marker.")
	case cds.Coins() < o.CostFor(cp):
		return nil, nil, nil, 0, sn.NewVError("You selected cards having %d total coins, but you need %d coins to bribe the selected official.", cds.Coins(), cp.CostFor(o))
	default:
		return cds, m, o, cbs, nil
	}
}

func (p *Player) hasGiftObligationIn(m *Ministry) *Player {
	g := p.Game()
	for _, p2 := range g.Players() {
		p2inf := p2.influenceIn(m)
		if p.hasGiftFrom(p2) && p2inf > 0 && p.influenceIn(m) >= p2inf {
			return p2
		}
	}
	return nil
}

func (g *Game) EnableBribeOfficial(cu *user.User) bool {
	cp := g.CurrentPlayer()
	return g.IsCurrentPlayer(cu) && cp.canBribeAnOfficial()
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

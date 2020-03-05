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
	gob.RegisterName("*game.recruitArmyEntry", new(recruitArmyEntry))
}

func (g *Game) recruitArmy(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// Validate and get cards and cubes
	cards, cubes, err := g.validateRecruitArmy(c)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cubes
	cp.PlaceCubesIn(RecruitArmySpace, cubes)

	// Recruit Army
	cp.Armies -= 1
	cp.RecruitedArmies += 1

	// Move played cards from hand to discard pile
	cp.ConCardHand.Remove(cards...)
	g.ConDiscardPile.Append(cards...)

	// Create Action Object for logging
	e := cp.newRecruitArmyEntry(cards)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))

	return "", game.Cache, nil
}

type recruitArmyEntry struct {
	*Entry
	Played ConCards
}

func (p *Player) newRecruitArmyEntry(c ConCards) *recruitArmyEntry {
	g := p.Game()
	e := new(recruitArmyEntry)
	e.Entry = p.newEntry()
	e.Played = c
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *recruitArmyEntry) HTML() template.HTML {
	return restful.HTML("%s spent %d Confucius cards having %d licenses to recruit army.",
		e.Player().Name(), len(e.Played), e.Played.Licenses())
}

func (g *Game) validateRecruitArmy(c *gin.Context) (ConCards, int, error) {
	cubes, err := g.validatePlayerAction(c)
	if err != nil {
		return nil, 0, err
	}

	cards, err := g.getConCards(c, "recruit-army")
	if err != nil {
		return nil, 0, err
	}

	cp := g.CurrentPlayer()
	switch {
	case cards.Licenses() < cp.armyCost():
		return nil, 0, sn.NewVError("You selected cards having %d total licenses, but you need %d licenses to recruit and army.", cards.Licenses(), cp.armyCost())
	case !cp.hasArmies():
		return nil, 0, sn.NewVError("You have no armies to recruit.")
	}

	return cards, cubes, nil
}

func (g *Game) EnableRecruitArmy(c *gin.Context) bool {
	cp := g.CurrentPlayer()
	return g.CUserIsCPlayerOrAdmin(c) && cp.canRecruitAnArmy()
}

func (p *Player) canRecruitAnArmy() bool {
	g := p.Game()
	return g.inActionsOrImperialFavourPhase() && !p.PerformedAction && p.hasEnoughCubesFor(RecruitArmySpace) &&
		p.hasArmies() && p.canAffordAnArmy()
}

func (p *Player) canAffordAnArmy() bool {
	return p.ConCardHand.Licenses() >= p.armyCost()
}

func (p *Player) hasArmies() bool {
	return p.Armies > 0
}

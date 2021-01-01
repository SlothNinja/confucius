package confucius

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.taxIncomeEntry", new(taxIncomeEntry))
}

func (g *Game) taxIncome(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return "", game.None, err
	}

	// Create Action Object for logging
	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cube(s) In BuyGiftSpace
	cp.PlaceCubesIn(TaxIncomeSpace, cubes)

	// Perform Tax Action
	cp.ConCardHand.Append(g.DrawConCard(), g.DrawConCard())

	entry := cp.newTaxIncomeEntry()

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	return "", game.Cache, nil
}

type taxIncomeEntry struct {
	*Entry
}

func (p *Player) newTaxIncomeEntry() *taxIncomeEntry {
	g := p.Game()
	e := new(taxIncomeEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *taxIncomeEntry) HTML() template.HTML {
	return restful.HTML("%s received two Confucius cards of tax income.", e.Player().Name())
}

func (g *Game) EnableTaxIncome(cu *user.User) bool {
	cp := g.CurrentPlayer()
	return g.IsCurrentPlayer(cu) && cp.canCollectTaxIncome()
}

func (p *Player) canCollectTaxIncome() bool {
	g := p.Game()
	return g.inActionsOrImperialFavourPhase() && !p.PerformedAction && p.hasEnoughCubesFor(TaxIncomeSpace)
}

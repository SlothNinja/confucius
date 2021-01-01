package confucius

import (
	"encoding/gob"
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
	gob.RegisterName("*game.chooseChiefMinisterEntry", new(chooseChiefMinisterEntry))
}

func (g *Game) chooseChiefMinisterPhase() {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = ChooseChiefMinister
	if g.Round == 1 {
		g.RandomTurnOrder()
		g.SetChiefMinister(g.CurrentPlayer())
		g.ChiefMinister().PlaceCubesIn(ImperialFavourSpace, 1)
		g.SetCurrentPlayerers(g.nextPlayer())
		g.actionsPhase()
	} else {
		g.SetCurrentPlayerers(g.ChiefMinister())
	}
}

func (g *Game) chooseChiefMinister(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	recipient, err := g.validateChooseChiefMinister(c, cu)
	if err != nil {
		return "", game.None, err
	}

	// Appoint New ChiefMinister
	g.SetChiefMinister(recipient)
	g.ChiefMinister().PlaceCubesIn(ImperialFavourSpace, 1)

	// Clear Actions
	cp := g.CurrentPlayer()
	cp.clearActions()
	cp.PerformedAction = true

	// Create Action Object for logging
	e := cp.newChooseChiefMinisterEntry(recipient)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type chooseChiefMinisterEntry struct {
	*Entry
}

func (p *Player) newChooseChiefMinisterEntry(op *Player) *chooseChiefMinisterEntry {
	g := p.Game()
	e := new(chooseChiefMinisterEntry)
	e.Entry = p.newEntry()
	e.OtherPlayerID = op.ID()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *chooseChiefMinisterEntry) HTML() template.HTML {
	return restful.HTML("%s chose %s to be chief minister.", e.Player().Name(), e.OtherPlayer().Name())
}

func (g *Game) validateChooseChiefMinister(c *gin.Context, cu *user.User) (*Player, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	recipientID, err := strconv.Atoi(c.PostForm("player"))
	if err != nil {
		return nil, err
	}

	recipient := g.PlayerByID(recipientID)
	cp := g.CurrentPlayer()
	switch {
	case recipient == nil:
		return nil, sn.NewVError("Recipient not found.")
	case !g.IsCurrentPlayer(cu):
		return nil, sn.NewVError("Only the current player may choose a chief minister.")
	case g.Phase != ChooseChiefMinister:
		return nil, sn.NewVError("You cannot choose a chief minister during the %s phase.", g.Phase)
	case cp.NotEqual(g.ChiefMinister()):
		return nil, sn.NewVError("Only the current chief minister may select the succeeding chief minister.")
	case cp.Equal(recipient):
		return nil, sn.NewVError("You cannot appoint yourself chief minister.")
	}
	return recipient, nil
}

func (g *Game) EnableChooseChiefMinister(cu *user.User) bool {
	return g.IsCurrentPlayer(cu) && g.Phase == ChooseChiefMinister
}

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
	gob.RegisterName("*game.noActionEntry", new(noActionEntry))
}

func (g *Game) noAction(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cube In NoActionSpace
	cp.PlaceCubesIn(NoActionSpace, cubes)

	// Create Action Object for logging
	entry := cp.newNoActionEntry()

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	return "", game.Cache, nil
}

type noActionEntry struct {
	*Entry
}

func (p *Player) newNoActionEntry() *noActionEntry {
	g := p.Game()
	e := new(noActionEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (g *noActionEntry) HTML() template.HTML {
	return restful.HTML("%s performed no action.", g.Player().Name())
}

func (g *Game) EnableNoAction(cu *user.User) bool {
	cp := g.CurrentPlayer()
	return g.inActionsOrImperialFavourPhase() && g.CurrentPlayer() != nil &&
		!cp.PerformedAction && g.IsCurrentPlayer(cu) &&
		cp.hasEnoughCubesFor(NoActionSpace) && cp.hasActionCubes()
}

func (p *Player) hasActionCubes() bool {
	return p.ActionCubes > 0
}

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
	gob.RegisterName("*game.passEntry", new(passEntry))
	gob.RegisterName("*game.autoPassEntry", new(autoPassEntry))
}

func (g *Game) pass(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err := g.validatePlayerAction(c); err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	if err := cp.validatePass(c); err != nil {
		return "", game.None, err
	}

	cp.pass()

	// Create Action Object for logging
	e := cp.newPassEntry()

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

func (p *Player) pass() {
	// Pass
	p.Passed = true
	p.PerformedAction = true
}

type passEntry struct {
	*Entry
}

func (p *Player) newPassEntry() *passEntry {
	e := new(passEntry)
	g := p.Game()
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *passEntry) HTML() template.HTML {
	return restful.HTML("%s passed.", e.Player().Name())
}

func (p *Player) autoPass() {
	p.pass()
	p.newAutoPassEntry()
}

type autoPassEntry struct {
	*Entry
}

func (p *Player) newAutoPassEntry() *autoPassEntry {
	e := new(autoPassEntry)
	g := p.Game()
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *autoPassEntry) HTML() template.HTML {
	return restful.HTML("System auto passed for %s.", e.Player().Name())
}

func (p *Player) validatePass(c *gin.Context) (err error) {
	if _, err = p.Game().validatePlayerAction(c); err == nil && p.hasActionCubes() {
		err = sn.NewVError("You must use all of your action cubes before passing.")
	}
	return
}

func (g *Game) EnablePass(c *gin.Context) bool {
	cp := g.CurrentPlayer()
	return g.CUserIsCPlayerOrAdmin(c) && cp.canPass()
}

func (p *Player) canPass() bool {
	g := p.Game()
	return g.Phase == Actions && !p.PerformedAction && !p.Passed && !p.Game().ExtraAction && !p.hasActionCubes()
}

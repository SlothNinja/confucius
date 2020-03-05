package confucius

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	gtype "github.com/SlothNinja/type"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const kind = "Game"

func New(c *gin.Context, id int64) *Game {
	g := new(Game)
	g.Header = game.NewHeader(c, g, id)
	g.State = newState()
	g.Key.Parent = pk(c)
	g.Type = gtype.Confucius
	return g
}

func newState() *State {
	return new(State)
}

func pk(c *gin.Context) *datastore.Key {
	return datastore.NameKey(gtype.Confucius.SString(), "root", game.GamesRoot(c))
}

func newKey(c *gin.Context, id int64) *datastore.Key {
	return datastore.IDKey(kind, id, pk(c))
}

func (g *Game) NewKey(c *gin.Context, id int64) *datastore.Key {
	return newKey(c, id)
}

func (g *Game) init(c *gin.Context) error {
	if err := g.Header.AfterLoad(g); err != nil {
		return err
	}

	for _, player := range g.Players() {
		player.init(g)
	}

	for _, entry := range g.Log {
		entry.Init(g)
	}

	for _, ministry := range g.Ministries {
		ministry.init(g)
	}

	for _, candidate := range g.Candidates {
		candidate.game = g
	}

	for _, land := range g.ForeignLands {
		land.init(g)
	}

	for _, land := range g.DistantLands {
		land.init(g)
	}
	return nil
}

func (g *Game) AfterCache() error {
	return g.init(g.CTX())
}

func (g *Game) fromForm(c *gin.Context) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s := new(State)

	if err = restful.BindWith(c, s, binding.FormPost); err == nil {
		g.BasicGame = s.BasicGame
		g.AdmiralVariant = s.AdmiralVariant
	}
	return
}

package confucius

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
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

func (client Client) init(c *gin.Context, g *Game) error {
	err := client.Game.AfterLoad(c, g.Header)
	if err != nil {
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

func (client Client) AfterCache(c *gin.Context, g *Game) error {
	return client.init(c, g)
}

func (g *Game) fromForm(c *gin.Context, cu *user.User) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	obj := struct {
		Title          string `form:"title"`
		NumPlayers     int    `form:"num-players" binding"min=0,max=5"`
		Password       string `form:"password"`
		BasicGame      bool   `form:"basic-game"`
		AdmiralVariant bool   `form:"admiral-variant"`
	}{}

	err := c.ShouldBind(&obj)
	if err != nil {
		// if err = restful.BindWith(c, h2, binding.FormPost); err != nil {
		return err
	}

	log.Debugf("obj: %#v", obj)

	// s := new(State)

	// if err = restful.BindWith(c, s, binding.FormPost); err == nil {
	// 	g.BasicGame = s.BasicGame
	// 	g.AdmiralVariant = s.AdmiralVariant
	// }

	g.Title = cu.Name + "'s Game"
	if obj.Title != "" {
		g.Title = obj.Title
	}

	g.NumPlayers = 4
	if obj.NumPlayers >= 1 && obj.NumPlayers <= 5 {
		g.NumPlayers = obj.NumPlayers
	}

	g.BasicGame = obj.BasicGame
	g.AdmiralVariant = obj.AdmiralVariant
	g.Password = obj.Password

	g.Creator = cu
	g.CreatorID = cu.ID()
	g.CreatorSID = user.GenID(cu.GoogleID)
	g.AddUser(cu)
	g.Status = game.Recruiting
	g.Type = gtype.Confucius

	log.Debugf("g: %#v", g)
	return nil
}

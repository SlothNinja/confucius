package confucius

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *Game) imperialFavourPhase(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = ImperialFavour
	g.ChiefMinister().ActionCubes += 1
	g.ChiefMinister().clearActions()
	g.SetCurrentPlayerers(g.ChiefMinister())
	g.ActionSpaces[ImperialFavourSpace].returnActionCubes()
	for _, player := range g.Players() {
		player.Passed = false
	}
	return
}

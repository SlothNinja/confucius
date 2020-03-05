package confucius

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *Game) buildWallPhase(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = BuildWall
	g.beginningOfPhaseReset()
	g.Wall += 1
}

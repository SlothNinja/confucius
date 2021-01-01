package confucius

import (
	"github.com/SlothNinja/log"
)

func (g *Game) buildWallPhase() {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = BuildWall
	g.beginningOfPhaseReset()
	g.Wall += 1
}

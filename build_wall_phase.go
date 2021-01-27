package confucius

import (
	"github.com/SlothNinja/log"
)

func (g *Game) buildWallPhase() {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = BuildWall
	g.beginningOfPhaseReset()
	g.Wall += 1
}

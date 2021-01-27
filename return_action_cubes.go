package confucius

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *Game) returnActionCubesPhase(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	for _, s := range g.ActionSpaces {
		s.returnActionCubes()
	}
}

func (s *ActionSpace) returnActionCubes() {
	s.Cubes = make(Cubes)
}

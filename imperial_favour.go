package confucius

import (
	"github.com/SlothNinja/log"
)

func (g *Game) imperialFavourPhase() {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = ImperialFavour
	g.ChiefMinister().ActionCubes += 1
	g.ChiefMinister().clearActions()
	g.SetCurrentPlayerers(g.ChiefMinister())
	g.ActionSpaces[ImperialFavourSpace].returnActionCubes()
	for _, p := range g.Players() {
		p.Passed = false
	}
}

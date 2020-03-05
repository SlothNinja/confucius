package confucius

import (
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (g *Game) validatePlayerAction(c *gin.Context) (cbs int, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var sid SpaceID
	if sid, err = g.getSpaceID(c); err != nil {
		return
	}

	if cbs, err = g.validatePlaceCubesFor(sid); err != nil {
		return
	}

	switch a, cp := c.PostForm("action"), g.CurrentPlayer(); {
	case !g.CUserIsCPlayerOrAdmin(c):
		err = sn.NewVError("Only the current player may perform the player action %q.", a)
	case (a == "pass" || IsEmperorRewardAction(a)) && g.Phase != Actions:
		err = sn.NewVError("You cannot perform a %q action during the %s phase.", a, g.PhaseName())
	case !g.inActionsOrImperialFavourPhase():
		err = sn.NewVError("You cannot perform a %q action during the %s phase.", a, g.PhaseName())
	case cp.Passed:
		err = sn.NewVError("You cannot perform a player action after passing.")
	}
	return
}

func (g *Game) validatePlaceCubesFor(id SpaceID) (cbs int, err error) {
	cp := g.CurrentPlayer()
	if cbs = cp.RequiredCubesFor(id); !cp.hasEnoughCubesFor(id) {
		err = sn.NewVError("You must have at least %d Action Cubes to perform this action.", cbs)
	}
	return
}

func IsEmperorRewardAction(s string) bool {
	return s == "Take Cash" || s == "Take Gift" || s == "Extra Action" || s == "Bribery Reward" ||
		s == "Avenge Emperor" || s == "Take Army"
}

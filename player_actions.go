package confucius

import (
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (g *Game) validatePlayerAction(c *gin.Context, cu *user.User) (int, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	sid, err := g.getSpaceID(c)
	if err != nil {
		return 0, err
	}

	cbs, err := g.validatePlaceCubesFor(sid, cu)
	if err != nil {
		return 0, err
	}

	a, cp := c.PostForm("action"), g.CurrentPlayer()
	switch {
	case !g.IsCurrentPlayer(cu):
		return 0, sn.NewVError("Only the current player may perform the player action %q.", a)
	case (a == "pass" || IsEmperorRewardAction(a)) && g.Phase != Actions:
		return 0, sn.NewVError("You cannot perform a %q action during the %s phase.", a, g.PhaseName())
	case !g.inActionsOrImperialFavourPhase():
		return 0, sn.NewVError("You cannot perform a %q action during the %s phase.", a, g.PhaseName())
	case cp.Passed:
		return 0, sn.NewVError("You cannot perform a player action after passing.")
	default:
		return cbs, nil
	}
}

func (g *Game) validatePlaceCubesFor(id SpaceID, cu *user.User) (int, error) {
	cp := g.CurrentPlayer()
	cbs := cp.RequiredCubesFor(id)
	if !cp.hasEnoughCubesFor(id) {
		return 0, sn.NewVError("You must have at least %d Action Cubes to perform this action.", cbs)
	}
	return cbs, nil
}

func IsEmperorRewardAction(s string) bool {
	return s == "Take Cash" || s == "Take Gift" || s == "Extra Action" || s == "Bribery Reward" ||
		s == "Avenge Emperor" || s == "Take Army"
}

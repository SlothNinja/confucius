package confucius

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (client *Client) finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := gameFrom(c)

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		oldCP := g.CurrentPlayer()

		var (
			s  *user.Stats
			cs []*contest.Contest
		)

		switch g.Phase {
		case Actions:
			s, err = g.actionsPhaseFinishTurn(c, cu)
		case ImperialFavour:
			s, cs, err = client.imperialFavourFinishTurn(c, g, cu)
		case ChooseChiefMinister:
			s, err = g.chooseChiefMinisterPhaseFinishTurn(c, cu)
		case Discard:
			s, cs, err = client.discardPhaseFinishTurn(c, g, cu)
		case ImperialExamination:
			s, err = g.tutorStudentsPhaseFinishTurn(c, cu)
		case ExaminationResolution:
			s, cs, err = client.examinationResolutionFinishTurn(c, g, cu)
		case MinistryResolution:
			s, cs, err = client.ministryResolutionFinishTurn(c, g, cu)
		default:
			err = sn.NewVError("Improper Phase for finishing turn.")
		}

		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		// cs != nil then game over
		if cs != nil {
			g.Phase = GameOver
			g.Status = game.Completed
			ks, es := wrap(s.GetUpdate(c, g.UpdatedAt), cs)
			err = client.saveWith(c, g, cu, ks, es)
			if err != nil {
				client.Log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(c, prefix))
				return
			}

			err = g.SendEndGameNotifications(c)
			if err != nil {
				client.Log.Errorf(err.Error())
			}

			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		// Game not over
		s = s.GetUpdate(c, g.UpdatedAt)
		err = client.saveWith(c, g, cu, []*datastore.Key{s.Key}, []interface{}{s})
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		newCP := g.CurrentPlayer()
		if newCP != nil && oldCP.ID() != newCP.ID() {
			err = g.SendTurnNotificationsTo(c, newCP)
			if err != nil {
				client.Log.Errorf(err.Error())
			}
		}

		c.Redirect(http.StatusSeeOther, showPath(c, prefix))
		return
	}
}

func (g *Game) validateFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.CurrentPlayer()

	s := user.StatsFetched(c)
	switch {
	case !g.IsCurrentPlayer(cu):
		return nil, sn.NewVError("Only the current player may finish a turn.")
	case !cp.PerformedAction:
		return nil, sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	default:
		return s, nil
	}
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(ps ...*Player) *Player {
	i := game.IndexFor(g.CurrentPlayer(), g.Playerers) + 1
	if len(ps) == 1 {
		i = game.IndexFor(ps[0], g.Playerers) + 1
	}
	return g.Players()[i%g.NumPlayers]
}

func (p *Player) canAutoPass() bool {
	return p.canPass() && !p.canTransferInfluence() && !p.canEmperorReward()
}

func (g *Game) actionsPhaseFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, err
	}

	cp := g.CurrentPlayer()
	restful.AddNoticef(c, "%s finished turn.", g.NameFor(cp))

	// Reveal Cards
	cp.ConCardHand.Reveal()
	cp.EmperorHand.Reveal()

	// Reset Extra Action
	g.ExtraAction = false

	p := g.actionPhaseNextPlayer()
	if p != nil {
		g.SetCurrentPlayerers(p)
		return s, nil
	}

	g.imperialFavourPhase()
	return s, nil
}

func (g *Game) actionPhaseNextPlayer(players ...*Player) *Player {
	ps := g.Players()
	p := g.nextPlayer(players...)
	for !ps.allPassed() {
		if p.Passed {
			p = g.nextPlayer(p)
		} else {
			p.beginningOfTurnReset()
			if p.canAutoPass() {
				p.autoPass()
				p = g.nextPlayer(p)
			} else {
				return p
			}
		}
	}
	return nil
}

func (client *Client) imperialFavourFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	cp := g.CurrentPlayer()
	restful.AddNoticef(c, "%s finished turn.", g.NameFor(cp))

	// Reveal Cards
	cp.ConCardHand.Reveal()
	cp.EmperorHand.Reveal()

	g.buildWallPhase()
	completed := g.examinationPhase(c, cu)
	if !completed {
		return s, nil, nil
	}

	completed = g.ministryResolutionPhase(c, false)
	if !completed {
		return s, nil, nil
	}
	g.invasionPhase(c)
	cs, err := client.endOfRoundPhase(c, g)
	return s, cs, err
}

func (g *Game) chooseChiefMinisterPhaseFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, err
	}

	for _, p := range g.Players() {
		p.PerformedAction = false
	}
	g.SetCurrentPlayerers(g.nextPlayer(g.ChiefMinister()))
	g.actionsPhase()
	return s, nil
}

func (client *Client) examinationResolutionFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	// Place New Candidate
	if len(g.Candidates) > 0 {
		g.Candidates = g.Candidates[1:]
	}

	var i int
	for index, candidate := range g.Candidates {
		i = index
		if candidate.Playable() {
			break
		}
	}
	g.Candidates = g.Candidates[i:]
	completed := g.ministryResolutionPhase(c, false)
	if !completed {
		return s, nil, nil
	}

	g.invasionPhase(c)
	cs, err := client.endOfRoundPhase(c, g)
	return s, cs, err
}

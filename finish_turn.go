package confucius

import (
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

func Finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		defer c.Redirect(http.StatusSeeOther, showPath(c, prefix))

		g := gameFrom(c)
		oldCP := g.CurrentPlayer()

		var (
			s   *stats.Stats
			cs  contest.Contests
			err error
		)

		switch g.Phase {
		case Actions:
			s, err = g.actionsPhaseFinishTurn(c)
		case ImperialFavour:
			s, cs, err = g.imperialFavourFinishTurn(c)
		case ChooseChiefMinister:
			s, err = g.chooseChiefMinisterPhaseFinishTurn(c)
		case Discard:
			s, cs, err = g.discardPhaseFinishTurn(c)
		case ImperialExamination:
			s, err = g.tutorStudentsPhaseFinishTurn(c)
		case ExaminationResolution:
			s, cs, err = g.examinationResolutionFinishTurn(c)
		case MinistryResolution:
			s, cs, err = g.ministryResolutionFinishTurn(c)
		default:
			err = sn.NewVError("Improper Phase for finishing turn.")
		}

		if err != nil {
			log.Errorf(err.Error())
			return
		}

		// cs != nil then game over
		if cs != nil {
			g.Phase = GameOver
			g.Status = game.Completed
			ks, es := wrap(s.GetUpdate(c, time.Time(g.UpdatedAt)), cs)
			err = g.saveWith(c, ks, es)
			if err == nil {
				err = g.SendEndGameNotifications(c)
			}
		} else {
			s := s.GetUpdate(c, time.Time(g.UpdatedAt))
			err = g.saveWith(c, []*datastore.Key{s.Key}, []interface{}{s})
			if err == nil {
				if newCP := g.CurrentPlayer(); newCP != nil && oldCP.ID() != newCP.ID() {
					err = g.SendTurnNotificationsTo(c, newCP)
				}
			}
		}

		if err != nil {
			log.Errorf(err.Error())
		}

		return
	}
}

func (g *Game) validateFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cp := g.CurrentPlayer()

	switch s = stats.Fetched(c); {
	case !g.CUserIsCPlayerOrAdmin(c):
		err = sn.NewVError("Only the current player may finish a turn.")
	case !cp.PerformedAction:
		err = sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	}
	return
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

func (g *Game) actionsPhaseFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	if s, err = g.validateFinishTurn(c); err != nil {
		return
	}

	cp := g.CurrentPlayer()
	restful.AddNoticef(c, "%s finished turn.", g.NameFor(cp))

	// Reveal Cards
	cp.ConCardHand.Reveal()
	cp.EmperorHand.Reveal()

	// Reset Extra Action
	g.ExtraAction = false

	if p := g.actionPhaseNextPlayer(); p != nil {
		g.SetCurrentPlayerers(p)
	} else {
		g.imperialFavourPhase(c)
	}
	return
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

func (g *Game) imperialFavourFinishTurn(c *gin.Context) (s *stats.Stats, cs contest.Contests, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateFinishTurn(c); err != nil {
		return
	}

	cp := g.CurrentPlayer()
	restful.AddNoticef(c, "%s finished turn.", g.NameFor(cp))

	// Reveal Cards
	cp.ConCardHand.Reveal()
	cp.EmperorHand.Reveal()

	g.buildWallPhase(c)
	if completed := g.examinationPhase(c); completed {
		if completed := g.ministryResolutionPhase(c, false); completed {
			g.invasionPhase(c)
			cs = g.endOfRoundPhase(c)
		}
	}
	return
}

func (g *Game) chooseChiefMinisterPhaseFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateFinishTurn(c); err != nil {
		return
	}

	for _, player := range g.Players() {
		player.PerformedAction = false
	}
	g.SetCurrentPlayerers(g.nextPlayer(g.ChiefMinister()))
	g.actionsPhase(c)
	return
}

func (g *Game) examinationResolutionFinishTurn(c *gin.Context) (s *stats.Stats, cs contest.Contests, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateFinishTurn(c); err != nil {
		return
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
	if completed := g.ministryResolutionPhase(c, false); completed {
		g.invasionPhase(c)
		cs = g.endOfRoundPhase(c)
	}
	return
}

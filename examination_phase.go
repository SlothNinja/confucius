package confucius

import (
	"encoding/gob"
	"fmt"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.studentPromotionEntry", new(studentPromotionEntry))
}

// Returns true if no further player actions are needed in order
// to resolve examination phase.
func (g *Game) examinationPhase(c *gin.Context) (completed bool) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch g.Phase, completed = ImperialExamination, true; {
	case g.canResolveExamination():
		completed = false
		g.resolveExamination(c)
	case g.canHoldExamination():
		// Select player before Chief Minister, so nextplayer selects Chief Minister.
		// This seems round-about, but it triggers the logic to skip nextplayers(s), if he has no cards to play.
		completed = false
		i := game.IndexFor(g.ChiefMinister(), g.Playerers) - 1
		if i == -1 {
			i += g.NumPlayers
		}

		if p := g.tutorStudentsPhaseNextPlayer(c, g.PlayerByIndex(i)); p != nil {
			g.SetCurrentPlayerers(p)
		} else {
			g.resolveExamination(c)
		}
	}
	return
}

func (g *Game) canHoldExamination() bool {
	c := g.Candidate()
	return g.Phase == ImperialExamination && g.Round > 1 &&
		((c.hasOnePlayer() && g.examinationForced()) || c.hasTwoPlayers())
	//	return g.Phase == ImperialExamination && g.Round > 1 && g.hasPlaceForCandidate() &&
	//		((c.hasOnePlayer() && g.examinationForced()) || c.hasTwoPlayers())
}

func (g *Game) canResolveExamination() bool {
	c := g.Candidate()
	return g.Phase == ImperialExamination && g.Round > 1 && (c.hasTwoSamePlayers() ||
		(c.hasOnePlayer() && g.examinationForced()))
	//	return g.Phase == ImperialExamination && g.Round > 1 && (c.hasTwoSamePlayers() ||
	//		(c.hasOnePlayer() && g.examinationForced()) || (c.hasTwoPlayers() && !g.hasPlaceForCandidate()))
}

//func (g *Game) hasPlaceForCandidate() bool {
//	return len(g.MinistriesFor(g.Candidate())) > 0
//}

func (g *Game) examinationForced() bool {
	return g.ActionSpaces[ForceSpace].CubeCount() > 0
}

func (g *Game) resolveExamination(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var winningCards, losingCards ConCards
	var loser *Player

	g.Phase = ExaminationResolution
	g.beginningOfPhaseReset()

	can := g.Candidate()
	if can.hasOnePlayer() || can.hasTwoSamePlayers() {
		cp := can.Player()
		g.SetCurrentPlayerers(cp)
		cp.newStudentPromotionEntry(nil, nil, nil, false)
	} else {
		coins0 := can.PlayerCards.Coins()
		coins1 := can.OtherPlayerCards.Coins()
		if coins0 >= coins1 {
			g.SetCurrentPlayerers(can.Player())
			winningCards = can.PlayerCards
			losingCards = can.OtherPlayerCards
			loser = can.OtherPlayer()
		} else {
			g.SetCurrentPlayerers(can.OtherPlayer())
			winningCards = can.OtherPlayerCards
			losingCards = can.PlayerCards
			loser = can.Player()
		}
		// Move played cards to discard pile
		g.ConDiscardPile.Append(winningCards...)
		g.ConDiscardPile.Append(losingCards...)
		cp := g.CurrentPlayer()
		cp.newStudentPromotionEntry(loser, winningCards, losingCards, true)
	}
}

type studentPromotionEntry struct {
	*Entry
	WinningCards ConCards
	LosingCards  ConCards
	Contested    bool
}

func (p *Player) newStudentPromotionEntry(player *Player, wcards, lcards ConCards, contested bool) *studentPromotionEntry {
	g := p.Game()
	e := &studentPromotionEntry{
		Entry:        p.newEntry(),
		WinningCards: wcards,
		LosingCards:  lcards,
		Contested:    contested,
	}
	if player != nil {
		e.OtherPlayerID = player.ID()
	} else {
		e.OtherPlayerID = NoPlayerID
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *studentPromotionEntry) HTML() template.HTML {
	var s string
	if e.Contested {
		s = fmt.Sprintf("<div>%s won the Imperial Examination.</div>", e.Player().Name())

		winningLength := len(e.WinningCards)
		s += fmt.Sprintf("<div>The student of %s received %d coins on %d %s.</div>",
			e.Player().Name(), e.WinningCards.Coins(), winningLength, pluralize("card", winningLength))

		losingLength := len(e.LosingCards)
		s += fmt.Sprintf("<div>The student of %s received %d coins on %d %s.</div>",
			e.OtherPlayer().Name(), e.LosingCards.Coins(), losingLength, pluralize("card", losingLength))
		return template.HTML(s)
	}
	return template.HTML(fmt.Sprintf("<div>%s won the Imperial Examination uncontested.</div>", e.Player().Name()))
}

//func (g *Game) oneStudent() bool {
//	return g.Candidate().Player() != nil && g.Candidate().OtherPlayer() == nil
//}
//
//func (g *Game) twoStudents() bool {
//	return g.Candidate().hasTwoStudents()
//}
//
//func (g *Game) bothStudentsSamePlayer() bool {
//	return g.twoStudents() && g.Candidate().Player().Equal(g.Candidate().OtherPlayer())
//}

func (g *Game) EnableTutorStudent(c *gin.Context) bool {
	return g.CUserIsCPlayerOrAdmin(c) && g.Phase == ImperialExamination && !g.CurrentPlayer().PerformedAction
}

package confucius

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.nominateStudentEntry", new(nominateStudentEntry))
}

func (g *Game) nominateStudent(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cds, cbs, err := g.validateNominateStudent(c)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Place Action Cubes
	cp.PlaceCubesIn(NominateSpace, cbs)

	// Move played cards from hand to discard pile
	cp.ConCardHand.Remove(cds...)
	g.ConDiscardPile.Append(cds...)

	// Place Student
	can := g.Candidate()
	if can.hasOnePlayer() {
		can.setOtherPlayer(cp)
	} else {
		can.setPlayer(cp)
	}

	// Create Action Object for logging
	e := cp.newNominateStudentEntry(cds)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type nominateStudentEntry struct {
	*Entry
	Played ConCards
}

func (p *Player) newNominateStudentEntry(cds ConCards) *nominateStudentEntry {
	g := p.Game()
	e := new(nominateStudentEntry)
	e.Entry = p.newEntry()
	e.Played = cds
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *nominateStudentEntry) HTML() template.HTML {
	length := len(e.Played)
	return restful.HTML("%s spent %d %s having %d coins to nominate student.",
		e.Player().Name(), length, pluralize("card", length), e.Played.Coins())
}

func (g *Game) validateNominateStudent(c *gin.Context) (ConCards, int, error) {
	cbs, err := g.validatePlayerAction(c)
	if err != nil {
		return nil, 0, err
	}

	cds, err := g.getConCards(c, "nominate-student")
	if err != nil {
		return nil, 0, err
	}

	cp := g.CurrentPlayer()
	can := g.Candidate()
	coinValue := cds.Coins()
	switch {
	case g.Round == 1:
		return nil, 0, sn.NewVError("You cannot nominate a student during round %d.", g.Round)
	case can.hasTwoPlayers():
		return nil, 0, sn.NewVError("There are already two students.")
	case cp.Equal(can.Player()):
		fallthrough
	case cp.Equal(can.OtherPlayer()):
		return nil, 0, sn.NewVError("You already have a nominated student.")
	case !cp.canAffordNomination():
		return nil, 0, sn.NewVError("You selected cards having %d total coins, but you need 2 coins to nominate a student.", coinValue)
	}
	return cds, cbs, nil
}

func (g *Game) EnableNominateStudent(c *gin.Context) bool {
	cp := g.CurrentPlayer()
	cd := g.Candidate()
	return g.inActionsOrImperialFavourPhase() && cp != nil && g.CUserIsCPlayerOrAdmin(c) &&
		g.Round > 1 && !cp.PerformedAction && cp.hasEnoughCubesFor(NominateSpace) && cd.hasSpaceFor(cp) &&
		cp.canAffordNomination()
}

func (p *Player) canAffordNomination() bool {
	return p.ConCardHand.Coins() >= 2
}

func (c *CandidateTile) hasSpaceFor(p *Player) bool {
	return !c.hasTwoPlayers() && p.NotEqual(c.Player())
}

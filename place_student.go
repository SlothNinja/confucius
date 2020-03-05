package confucius

import (
	"encoding/gob"
	"fmt"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.placeStudentEntry", new(placeStudentEntry))
}

func (g *Game) placeStudent(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	ministry, seniority, err := g.validatePlaceStudent(c)
	if err != nil {
		return "", game.None, err
	}

	var replacedOfficial *OfficialTile
	if ministry != nil {
		replacedOfficial = ministry.Officials[seniority]
	}
	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Create Action Object for logging
	entry := cp.newPlaceStudentEntry()
	if ministry != nil {
		entry.MinistryName = ministry.Name()
	} else {
		entry.MinistryName = "None"
	}
	entry.Seniority = seniority
	if replacedOfficial != nil && replacedOfficial.Player() != nil {
		entry.OtherPlayerID = replacedOfficial.Player().ID()
	} else {
		entry.OtherPlayerID = NoPlayerID
	}

	var official *OfficialTile
	if ministry != nil {
		// Place Student In Spot
		official = g.Candidate().OfficialTile
		official.game = g
		official.ministry = ministry
		official.Seniority = seniority
		ministry.Officials[seniority] = official
	}

	// Place Secured Marker on Student
	if official != nil {
		official.setPlayer(cp)
		official.Secured = true
	}

	// Remove Candidate and Show Back of Candidates in Stack
	tileBack := newCandidateTile()

	// Set Variant to tile back
	tileBack.Variant = TileBack
	tileBack.PlayerID = NoPlayerID
	tileBack.OtherPlayerID = NoPlayerID

	// Display Back
	g.Candidates[0] = tileBack

	// Set flash message
	restful.AddNoticef(c, string(entry.HTML()))
	return "", game.Cache, nil
}

type placeStudentEntry struct {
	*Entry
	MinistryName string
	Seniority    Seniority
}

func (p *Player) newPlaceStudentEntry() *placeStudentEntry {
	g := p.Game()
	e := new(placeStudentEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *placeStudentEntry) HTML() template.HTML {
	if e.MinistryName == "None" {
		return template.HTML(fmt.Sprintf("%s was unable to place student.", e.Player().Name()))
	}
	if e.OtherPlayer() == nil {
		return template.HTML(fmt.Sprintf("%s placed student in seniority spot %d of %s ministry.",
			e.Player().Name(), e.Seniority, e.MinistryName))
	}

	return template.HTML(fmt.Sprintf("%s placed student in seniority spot %d of %s ministry replacing official of %s.",
		e.Player().Name(), e.Seniority, e.MinistryName, e.OtherPlayer().Name()))
}

func (g *Game) validatePlaceStudent(c *gin.Context) (*Ministry, Seniority, error) {
	if len(g.MinistriesFor(g.Candidate())) == 0 {
		return nil, 0, nil
	}
	m, s, err := g.getMinistryAndSeniority(c, "official")
	if err != nil {
		return nil, 0, err
	}

	if !g.CUserIsCPlayerOrAdmin(c) {
		return nil, 0, sn.NewVError("Only the current player may place a student in a ministry.")
	}

	if g.Phase != ExaminationResolution {
		return nil, 0, sn.NewVError("You cannot place a student in a ministry during the %s phase.", g.Phase)
	}

	if !g.MinistriesFor(g.Candidate()).Include(m) {
		return nil, 0, sn.NewVError("You cannot place a student in ministry %s.", m.Name())
	}

	if spots := m.emptyCandidateSpots(); len(spots) > 0 {
		if !spots.Include(s) {
			return nil, 0, sn.NewVError("You cannot place a student in seniority spot %d of ministry %s.", s, m.Name())
		}
		return m, s, nil
	}

	if !m.unbribedUnsecuredCandidateSpots().Include(s) {
		return nil, 0, sn.NewVError("You cannot place a student in seniority spot %d of ministry %s.", s, m.Name())
	}

	return m, s, nil
}

func (g *Game) EnablePlaceStudent(c *gin.Context) bool {
	return g.CUserIsCPlayerOrAdmin(c) && g.Phase == ExaminationResolution && !g.CurrentPlayer().PerformedAction
}

func (g *Game) MinistriesFor(c *CandidateTile) Ministries {
	var ms Ministries
	switch c.Variant {
	case BingbuCandidate:
		ms = Ministries{Bingbu: g.Ministries[Bingbu]}
	case HubuCandidate:
		ms = Ministries{Hubu: g.Ministries[Hubu]}
	case GongbuCandidate:
		ms = Ministries{Gongbu: g.Ministries[Gongbu]}
	case AnyCandidate1, AnyCandidate2, AnyCandidate3:
		ms = g.Ministries
	default:
		return nil
	}

	// Check ministries of candidate variant
	if ms2 := ms.withEmptyCandidateSpots(); len(ms2) > 0 {
		return ms2
	}

	if ms2 := ms.withUnbribedUnsecuredCandidateSpots(); len(ms2) > 0 {
		return ms2
	}

	// Check all ministries
	ms = g.Ministries
	if ms2 := ms.withEmptyCandidateSpots(); len(ms2) > 0 {
		return ms2
	}

	return ms.withUnbribedUnsecuredCandidateSpots()
}

// Filter received ministries and return only those with at least one empty spot for placing a candidate.
func (ms Ministries) withEmptyCandidateSpots() Ministries {
	ms2 := make(Ministries)
	for k, m := range ms {
		if len(m.emptyCandidateSpots()) > 0 {
			ms2[k] = m
		}
	}
	return ms2
}

// Return those seniority spots without an official/candidate.
func (m *Ministry) emptyCandidateSpots() Seniorities {
	var spots Seniorities
	if m.Resolved {
		return spots
	}

	senioritySpots := []Seniority{1, 2, 6, 7}
	for _, seniority := range senioritySpots {
		if _, ok := m.Officials[seniority]; !ok {
			spots = append(spots, seniority)
		}
	}
	return spots
}

// Filter received ministries and return only those with at least one spot having an unbribed or unsecured official.
func (ms Ministries) withUnbribedUnsecuredCandidateSpots() Ministries {
	ms2 := make(Ministries)
	for k, m := range ms {
		if len(m.unbribedUnsecuredCandidateSpots()) > 0 {
			ms2[k] = m
		}
	}
	return ms2
}

// Return those seniority spots having an unbribed or unsecured official/candidate.
func (m *Ministry) unbribedUnsecuredCandidateSpots() Seniorities {
	var spots Seniorities
	if m.Resolved {
		return spots
	}

	senioritySpots := []Seniority{1, 2, 3, 4, 5, 6, 7}
	for _, seniority := range senioritySpots {
		if official, ok := m.Officials[seniority]; ok && (!official.Secured || official.Player() == nil) {
			spots = append(spots, seniority)
		}
	}
	return spots
}

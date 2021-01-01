package confucius

import (
	"fmt"
	"html/template"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

func (g *Game) ministryResolutionPhase(c *gin.Context, ending bool) bool {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = MinistryResolution

	// Clear Players to enable them to perform an action this phase.
	g.beginningOfPhaseReset()

	for _, mid := range []MinistryID{Bingbu, Hubu, Gongbu} {
		m := g.Ministries[mid]
		if !m.Resolved && (ending || m.MarkerCount() == 7) {
			completed := g.initMinistryResolution(c, m)
			if !completed {
				return completed
			}
		}
	}
	return true
}

func (g *Game) ministryInProgress() *Ministry {
	for _, m := range g.Ministries {
		if m.InProgress {
			return m
		}
	}
	return nil
}

func (g *Game) initMinistryResolution(c *gin.Context, m *Ministry) bool {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	m.InProgress = true
	for _, o := range m.Officials {
		o.Secured = true
		o.setTempPlayer(o.Player())
	}
	return g.resolve(c, m)
}

func (g *Game) playerCountsIn(m *Ministry) map[int]int {
	counts := make(map[int]int)
	for _, o := range m.Officials {
		if o.TempPlayer() != nil {
			counts[o.TempPlayer().ID()] += 1
		}
	}
	return counts
}

func (g *Game) resolve(c *gin.Context, m *Ministry) bool {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var playerCounts map[int]int
	for playerCounts = g.playerCountsIn(m); len(playerCounts) > 2; playerCounts = g.playerCountsIn(m) {
		from := m.playerToTempTransfer(playerCounts)
		g.SetCurrentPlayerers(from)
		to := from.TempPlayers()

		if len(to) != 1 {
			log.Debugf("to: %#v", to)
			log.Debugf("ministry: %#v", m)
			log.Debugf("playerCounts: %#v", playerCounts)
			return false
		}
		g.autoTempTransferInfluence(from, to[0])
	}

	ministerID, secretaryID := NoPlayerID, NoPlayerID
	switch len(playerCounts) {
	case 1:
		for pid := range playerCounts {
			ministerID, secretaryID = pid, pid
		}
	case 2:
		var ministerCount int
		for pid, count := range playerCounts {
			switch {
			case ministerCount == 0:
				ministerID, ministerCount = pid, count
			case ministerCount > count:
				secretaryID = pid
			case ministerCount < count:
				secretaryID = ministerID
				ministerID, ministerCount = pid, count
			case ministerCount == count:
				var official *OfficialTile
				for _, seniority := range []Seniority{1, 2, 3, 4, 5, 6, 7} {
					var ok bool
					official, ok = m.Officials[seniority]
					if ok && official.TempID != NoPlayerID {
						break
					}
				}
				switch official.TempID {
				case pid:
					secretaryID = ministerID
					ministerID, ministerCount = pid, count
				default:
					secretaryID = pid
				}
			}
		}
	}

	log.Debugf("ministerID: %#v", ministerID)
	minister := g.PlayerByID(ministerID)
	if minister != nil {
		minister.Score += m.MinisterChit.Value()
	}

	log.Debugf("secretaryID: %#v", secretaryID)
	secretary := g.PlayerByID(secretaryID)
	if secretary != nil {
		secretary.Score += m.SecretaryChit.Value()
	}

	// Create ActionLog Entry
	entry := g.newResolvedMinistryEntry()
	entry.MinistryName = m.Name()
	entry.MinisterID = ministerID
	entry.MinisterScore = m.MinisterChit.Value()
	entry.SecretaryID = secretaryID
	entry.SecretaryScore = m.SecretaryChit.Value()

	// Remove Temp Influence
	for _, o := range m.Officials {
		o.setTempPlayer(nil)
	}

	m.setMinister(minister)
	m.setSecretary(secretary)
	m.Resolved = true
	m.InProgress = false

	return true
}

type resolvedMinistryEntry struct {
	*Entry
	MinistryName   string
	MinisterID     int
	MinisterScore  int
	SecretaryID    int
	SecretaryScore int
}

func (g *Game) newResolvedMinistryEntry() *resolvedMinistryEntry {
	e := new(resolvedMinistryEntry)
	e.Entry = g.newEntry()
	g.Log = append(g.Log, e)
	return e
}

func (m *resolvedMinistryEntry) HTML() template.HTML {
	g := m.Game().(*Game)
	s := fmt.Sprintf("<div>%s Ministry Resolved</div>", m.MinistryName)
	if minister := g.PlayerByID(m.MinisterID); minister != nil {
		s += fmt.Sprintf("<div>%s awarded Minister position and %d points</div>", g.NameFor(minister), m.MinisterScore)
	} else {
		s += fmt.Sprintf("<div>No one awarded Minister position</div>")
	}
	if secretary := m.Game().(*Game).PlayerByID(m.SecretaryID); secretary != nil {
		s += fmt.Sprintf("<div>%s awarded Secretary position and %d points</div>", g.NameFor(secretary), m.SecretaryScore)
	} else {
		s += fmt.Sprintf("<div>No one awarded Secretary position</div>")
	}
	return restful.HTML(s)
}

func (m *Ministry) playerToTempTransfer(playerCounts map[int]int) *Player {
	min := 7
	var pids []int

	for pid, count := range playerCounts {
		if count == min {
			pids = append(pids, pid)
		} else if count < min {
			min = count
			pids = []int{pid}
		}
	}

	if len(pids) == 1 {
		return m.Game().PlayerByID(pids[0])
	}

	for _, seniority := range []Seniority{1, 2, 3, 4, 5, 6, 7} {
		official, ok := m.Officials[seniority]
		if !ok {
			continue
		}
		for i, pid := range pids {
			if official.TempID == pid {
				pids = append(pids[:i], pids[i+1:]...)
			}
			if len(pids) == 1 {
				return m.Game().PlayerByID(pids[0])
			}
		}
	}
	return nil
}

func (client Client) ministryResolutionFinishTurn(c *gin.Context, g *Game, cu *user.User) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	cp := g.CurrentPlayer()
	restful.AddNoticef(c, "%s finished turn.", g.NameFor(cp))
	resolved := g.resolve(c, g.ministryInProgress())
	if !resolved {
		return s, nil, nil
	}

	completed := g.ministryResolutionPhase(c, false)
	if !completed {
		return s, nil, nil
	}

	g.invasionPhase(c)
	cs, err := client.endOfRoundPhase(c, g)
	return s, cs, err
}

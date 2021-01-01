package confucius

import (
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/SlothNinja/sn"
)

func init() {
	gob.RegisterName("game.DistantLands", make(DistantLands, 0))
}

type DistantLandChit int

const NoChit DistantLandChit = -1

func (chit *DistantLandChit) Value() int {
	return int(*chit)
}

type DistantLandChits []DistantLandChit

type DistantLandID int

const (
	SpiceIslands DistantLandID = iota
	India
	Arabia
	Africa
	Americas
)

var distanLandIDStrings = map[DistantLandID]string{SpiceIslands: "The Spice Islands", India: "India", Arabia: "Arabia", Africa: "Africa", Americas: "The Americas"}

func (lid DistantLandID) String() string {
	return distanLandIDStrings[lid]
}

type DistantLand struct {
	game      *Game
	ID        DistantLandID
	Chit      DistantLandChit
	PlayerIDS []int
}
type DistantLands []*DistantLand

func (l *DistantLand) init(game *Game) {
	l.game = game
}

func (l *DistantLand) Name() string {
	return l.ID.String()
}

func (l *DistantLand) Players() Players {
	var ps Players
	for _, id := range l.PlayerIDS {
		p := l.game.PlayerByID(id)
		if p != nil {
			ps = append(ps, p)
		}
	}
	return ps
}

func (l *DistantLand) SetPlayers(ps Players) {
	switch {
	case len(ps) == 0:
		l.PlayerIDS = nil
	default:
		ids := make([]int, len(ps))
		for i, player := range ps {
			ids[i] = player.ID()
		}
		l.PlayerIDS = ids
	}
}

func (g *Game) hasDistantLandFor(p *Player) bool {
	for _, l := range g.DistantLands {
		if !l.Players().Include(p) {
			return true
		}
	}
	return false
}

var distanLandIDS = []DistantLandID{SpiceIslands, India, Arabia, Africa, Americas}

func (g *Game) CreateDistantLands() {
	distantLandChits := DistantLandChits{2, 2, 3, 3, 4, 4, 4}
	g.DistantLands = make(DistantLands, len(distanLandIDS))

	for _, key := range distanLandIDS {
		g.DistantLands[key] = new(DistantLand)
		g.DistantLands[key].ID = key
		g.DistantLands[key].Chit = distantLandChits.Draw()
	}
}

func (cs *DistantLandChits) Draw() DistantLandChit {
	var chit DistantLandChit
	*cs, chit = cs.DrawS()
	return chit
}

func (cs DistantLandChits) DrawS() (DistantLandChits, DistantLandChit) {
	i := sn.MyRand.Intn(len(cs))
	chit := cs[i]
	chits := append(cs[:i], cs[i+1:]...)
	return chits, chit
}

func (l *DistantLand) NameID() string {
	return strings.Replace(l.Name(), " ", "-", -1)
}

func (chit DistantLandChit) Image() string {
	if chit == 0 {
		return ""
	}
	return fmt.Sprintf("<img src=\"/images/confucius/land-chit-%dVP.jpg\" />", chit)
}

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

func (this *DistantLandChit) Value() int {
	return int(*this)
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

func (this DistantLandID) String() string {
	return distanLandIDStrings[this]
}

type DistantLand struct {
	game      *Game
	ID        DistantLandID
	Chit      DistantLandChit
	PlayerIDS []int
}
type DistantLands []*DistantLand

func (this *DistantLand) init(game *Game) {
	this.game = game
}

func (this *DistantLand) Name() string {
	return this.ID.String()
}

func (this *DistantLand) Players() (players Players) {
	for _, id := range this.PlayerIDS {
		player := this.game.PlayerByID(id)
		if player != nil {
			players = append(players, player)
		}
	}
	return
}

func (this *DistantLand) SetPlayers(players Players) {
	switch {
	case len(players) == 0:
		this.PlayerIDS = nil
	default:
		ids := make([]int, len(players))
		for i, player := range players {
			ids[i] = player.ID()
		}
		this.PlayerIDS = ids
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

func (this *Game) CreateDistantLands() {
	distantLandChits := DistantLandChits{2, 2, 3, 3, 4, 4, 4}
	this.DistantLands = make(DistantLands, len(distanLandIDS))

	for _, key := range distanLandIDS {
		this.DistantLands[key] = new(DistantLand)
		this.DistantLands[key].ID = key
		this.DistantLands[key].Chit = distantLandChits.Draw()
	}
}

func (this *DistantLandChits) Draw() (chit DistantLandChit) {
	*this, chit = this.DrawS()
	return
}

func (this DistantLandChits) DrawS() (chits DistantLandChits, chit DistantLandChit) {
	i := sn.MyRand.Intn(len(this))
	chit = this[i]
	chits = append(this[:i], this[i+1:]...)
	return
}

func (this *DistantLand) NameID() string {
	return strings.Replace(this.Name(), " ", "-", -1)
}

func (this DistantLandChit) Image() (image string) {
	if this == 0 {
		return ""
	}
	return fmt.Sprintf("<img src=\"/images/confucius/land-chit-%dVP.jpg\" />", this)
}

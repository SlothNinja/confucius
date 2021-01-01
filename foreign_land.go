package confucius

import (
	"encoding/gob"
	"strings"

	"github.com/SlothNinja/sn"
)

func init() {
	gob.RegisterName("*game.ForeignLand", new(ForeignLand))
	gob.RegisterName("*game.ForeignLandBox", new(ForeignLandBox))
}

type ForeignLandBox struct {
	land *ForeignLand
	//        Index           int
	Position  int
	PlayerID  int
	Points    int
	AwardCard bool
}
type ForeignLandBoxes []*ForeignLandBox

func (box *ForeignLandBox) Game() *Game {
	return box.land.Game()
}

func (box *ForeignLandBox) Player() *Player {
	if box.Invaded() {
		return box.Game().PlayerByID(box.PlayerID)
	}
	return nil
}

func (box *ForeignLandBox) setPlayer(p *Player) {
	switch {
	case p == nil:
		box.PlayerID = NoPlayerID
	default:
		box.PlayerID = p.ID()
	}
}

func (box *ForeignLandBox) Invaded() bool {
	return box.PlayerID != NoPlayerID
}

func (box *ForeignLandBox) NotInvaded() bool {
	return !box.Invaded()
}

type ForeignLand struct {
	game     *Game
	ID       ForeignLandID
	Boxes    ForeignLandBoxes
	Resolved bool
}

func (l *ForeignLand) init(game *Game) {
	l.game = game
	for _, box := range l.Boxes {
		box.land = l
	}
}

func (l *ForeignLand) Game() *Game {
	return l.game
}

func (l *ForeignLand) Name() string {
	return l.ID.String()
}

func (l *ForeignLand) Box(index int) *ForeignLandBox {
	return l.Boxes[index]
}

type ForeignLands []*ForeignLand
type ForeignLandID int

const (
	Annam ForeignLandID = iota
	Yunnan
	Mongolia
	Korea
	Manchuria
)

var foreignLandIDS = []ForeignLandID{Annam, Yunnan, Mongolia, Korea, Manchuria}
var foreignLandIDStrings = map[ForeignLandID]string{Annam: "Annam", Yunnan: "Yunnan", Mongolia: "Mongolia", Korea: "Korea", Manchuria: "Manchuria"}
var foreignLandIDCost = map[ForeignLandID]int{Annam: 8, Yunnan: 4, Mongolia: 6, Korea: 7, Manchuria: 5}

func (lid ForeignLandID) String() string {
	return foreignLandIDStrings[lid]
}

func (l *ForeignLand) Cost() int {
	return foreignLandIDCost[l.ID]
}

func (lid ForeignLandID) CreateBoxes(land *ForeignLand) ForeignLandBoxes {
	switch lid {
	case Annam:
		boxes := make(ForeignLandBoxes, 2)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
		return boxes
	case Yunnan:
		boxes := make(ForeignLandBoxes, 2)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 2, AwardCard: true}
		return boxes
	case Mongolia:
		boxes := make(ForeignLandBoxes, 3)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 2, AwardCard: false}
		boxes[2] = &ForeignLandBox{land: land, Position: 3, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		return boxes
	case Korea:
		boxes := make(ForeignLandBoxes, 3)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
		boxes[2] = &ForeignLandBox{land: land, Position: 2, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		return boxes
	case Manchuria:
		boxes := make(ForeignLandBoxes, 4)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 3, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 2, AwardCard: false}
		boxes[2] = &ForeignLandBox{land: land, Position: 2, PlayerID: NoPlayerID, Points: 5, AwardCard: false}
		boxes[3] = &ForeignLandBox{land: land, Position: 3, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
		return boxes
	default:
		return nil
	}
}

func (g *Game) CreateForeignLands() {
	// Create Foreign Lands
	lands := make(ForeignLands, len(foreignLandIDS))
	for i, id := range foreignLandIDS {
		land := new(ForeignLand)
		lands[i] = land
		lands[i].ID = id
		lands[i].Boxes = id.CreateBoxes(land)
	}

	// Select three random lands for the game
	selectedLands := make(ForeignLands, 3)
	for i := range selectedLands {
		index := sn.MyRand.Intn(len(lands))
		selectedLands[i] = lands[index]
		lands = append(lands[:index], lands[index+1:]...)
	}
	g.ForeignLands = selectedLands
}

func (l *ForeignLand) LString() string {
	return strings.ToLower(l.Name())
}

func (l *ForeignLand) AllBoxesOccupied() bool {
	for _, box := range l.Boxes {
		if box.NotInvaded() {
			return false
		}
	}
	return true
}

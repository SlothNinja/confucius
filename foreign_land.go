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

func (this *ForeignLandBox) Game() *Game {
	return this.land.Game()
}

func (this *ForeignLandBox) Player() (player *Player) {
	if this.Invaded() {
		player = this.Game().PlayerByID(this.PlayerID)
	}
	return
}

func (this *ForeignLandBox) setPlayer(player *Player) {
	switch {
	case player == nil:
		this.PlayerID = NoPlayerID
	default:
		this.PlayerID = player.ID()
	}
}

func (this *ForeignLandBox) Invaded() bool {
	return this.PlayerID != NoPlayerID
}

func (this *ForeignLandBox) NotInvaded() bool {
	return !this.Invaded()
}

type ForeignLand struct {
	game     *Game
	ID       ForeignLandID
	Boxes    ForeignLandBoxes
	Resolved bool
}

func (this *ForeignLand) init(game *Game) {
	this.game = game
	for _, box := range this.Boxes {
		box.land = this
	}
}

func (this *ForeignLand) Game() *Game {
	return this.game
}

func (this *ForeignLand) Name() string {
	return this.ID.String()
}

func (this *ForeignLand) Box(index int) *ForeignLandBox {
	return this.Boxes[index]
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

func (this ForeignLandID) String() string {
	return foreignLandIDStrings[this]
}

func (this *ForeignLand) Cost() int {
	return foreignLandIDCost[this.ID]
}

func (this ForeignLandID) CreateBoxes(land *ForeignLand) (boxes ForeignLandBoxes) {
	switch this {
	case Annam:
		boxes = make(ForeignLandBoxes, 2)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
	case Yunnan:
		boxes = make(ForeignLandBoxes, 2)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 2, AwardCard: true}
	case Mongolia:
		boxes = make(ForeignLandBoxes, 3)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 2, AwardCard: false}
		boxes[2] = &ForeignLandBox{land: land, Position: 3, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
	case Korea:
		boxes = make(ForeignLandBoxes, 3)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
		boxes[2] = &ForeignLandBox{land: land, Position: 2, PlayerID: NoPlayerID, Points: 4, AwardCard: false}
	case Manchuria:
		boxes = make(ForeignLandBoxes, 4)
		boxes[0] = &ForeignLandBox{land: land, Position: 0, PlayerID: NoPlayerID, Points: 3, AwardCard: false}
		boxes[1] = &ForeignLandBox{land: land, Position: 1, PlayerID: NoPlayerID, Points: 2, AwardCard: false}
		boxes[2] = &ForeignLandBox{land: land, Position: 2, PlayerID: NoPlayerID, Points: 5, AwardCard: false}
		boxes[3] = &ForeignLandBox{land: land, Position: 3, PlayerID: NoPlayerID, Points: 3, AwardCard: true}
	}
	return
}

func (this *Game) CreateForeignLands() {
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
	this.ForeignLands = selectedLands
}

func (this *ForeignLand) LString() string {
	return strings.ToLower(this.Name())
}

func (this *ForeignLand) AllBoxesOccupied() (result bool) {
	result = true
	for _, box := range this.Boxes {
		if box.NotInvaded() {
			result = false
			break
		}
	}
	return
}

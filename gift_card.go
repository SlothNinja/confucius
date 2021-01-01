package confucius

import "encoding/gob"

func init() {
	gob.RegisterName("*game.GiftCard", new(GiftCard))
}

type GiftCardValue int

const (
	NoGiftValue GiftCardValue = iota
	Hanging
	Tile
	Vase
	Coat
	Necklace
	Junk
)

func (cd GiftCardValue) Int() int {
	return int(cd)
}

func (cd GiftCardValue) String() string {
	return giftCardStrings[cd]
}

var giftCardValues = []GiftCardValue{Hanging, Tile, Vase, Coat, Necklace, Junk}
var giftCardStrings = map[GiftCardValue]string{
	Hanging:  "Hanging",
	Tile:     "Tile",
	Vase:     "Vase",
	Coat:     "Coat",
	Necklace: "Necklace",
	Junk:     "Junk",
}

func (g *Game) GiftCardValues() []GiftCardValue {
	return giftCardValues
}

//func (this GiftCardValue) String() string {
//        return giftCardStrings[this]
//}

type GiftCard struct {
	game     *Game
	Value    GiftCardValue
	PlayerID int
}
type GiftCards []*GiftCard

func (cd *GiftCard) Game() *Game {
	return cd.game
}

func (cd *GiftCard) SetGame(game *Game) {
	cd.game = game
}

func (cd *GiftCard) Cost() int {
	return cd.Value.Int()
}

func (cd *GiftCard) Player() *Player {
	if cd.PlayerID != NoPlayerID {
		return cd.Game().PlayerByID(cd.PlayerID)
	}
	return nil
}

func (cd *GiftCard) setPlayer(p *Player) {
	switch {
	case p == nil:
		cd.PlayerID = NoPlayerID
	default:
		cd.PlayerID = p.ID()
	}
}

func (cd *GiftCard) isFrom(p *Player) bool {
	return cd.Player().Equal(p)
}

func (cd *GiftCard) Name() string {
	return giftCardStrings[cd.Value]
}

func (cd *GiftCard) Equal(card *GiftCard) bool {
	return cd != nil && card != nil && cd.Value == card.Value && cd.Player().Equal(card.Player())
}

func (cds *GiftCards) Append(cards ...*GiftCard) {
	*cds = cds.AppendS(cards...)
}

func (cds GiftCards) AppendS(cards ...*GiftCard) GiftCards {
	return append(cds, cards...)
}

func (cds *GiftCards) Remove(cards ...*GiftCard) {
	*cds = cds.removeMulti(cards...)
}

func (cds GiftCards) removeMulti(cards ...*GiftCard) GiftCards {
	gcs := cds
	for _, c := range cards {
		gcs = gcs.remove(c)
	}
	return gcs
}

func (cds GiftCards) remove(card *GiftCard) GiftCards {
	cards := cds
	for i, c := range cds {
		if c.Equal(card) {
			return cards.removeAt(i)
		}
	}
	return cds
}

func (cds GiftCards) removeAt(i int) GiftCards {
	return append(cds[:i], cds[i+1:]...)
}

func (cds GiftCards) include(card *GiftCard) bool {
	for _, c := range cds {
		if c.Equal(card) {
			return true
		}
	}
	return false
}

func (g *Game) GiftCardNames() []string {
	var ss []string
	for _, v := range giftCardValues {
		ss = append(ss, giftCardStrings[v])
	}
	return ss
}

func (cds GiftCards) OfValue(v GiftCardValue) GiftCards {
	var cards GiftCards
	for _, card := range cds {
		if card.Value == v {
			cards = append(cards, card)
		}
	}
	return cards
}

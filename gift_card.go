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

func (this GiftCardValue) Int() int {
	return int(this)
}

func (this GiftCardValue) String() string {
	return giftCardStrings[this]
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

func (this *Game) GiftCardValues() []GiftCardValue {
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

func (this *GiftCard) Game() *Game {
	return this.game
}

func (this *GiftCard) SetGame(game *Game) {
	this.game = game
}

func (this *GiftCard) Cost() int {
	return this.Value.Int()
}

func (this *GiftCard) Player() (player *Player) {
	if this.PlayerID != NoPlayerID {
		player = this.Game().PlayerByID(this.PlayerID)
	}
	return
}

func (this *GiftCard) setPlayer(player *Player) {
	switch {
	case player == nil:
		this.PlayerID = NoPlayerID
	default:
		this.PlayerID = player.ID()
	}
}

func (c *GiftCard) isFrom(p *Player) bool {
        return c.Player().Equal(p)
}

func (this *GiftCard) Name() string {
	return giftCardStrings[this.Value]
}

func (c *GiftCard) Equal(card *GiftCard) bool {
	return c != nil && card != nil && c.Value == card.Value && c.Player().Equal(card.Player())
}

func (this *GiftCards) Append(cards ...*GiftCard) {
	*this = this.AppendS(cards...)
}

func (this GiftCards) AppendS(cards ...*GiftCard) (gs GiftCards) {
	gs = append(this, cards...)
	return
}

func (cs *GiftCards) Remove(cards ...*GiftCard) {
	*cs = cs.removeMulti(cards...)
}

func (cs GiftCards) removeMulti(cards ...*GiftCard) GiftCards {
        gcs := cs
	for _, c := range cards {
		gcs = gcs.remove(c)
	}
	return gcs
}

func (cs GiftCards) remove(card *GiftCard) GiftCards {
        cards := cs
	for i, c := range cs {
		if c.Equal(card) {
			return cards.removeAt(i)
		}
	}
	return cs
}

func (cs GiftCards) removeAt(i int) GiftCards {
	return append(cs[:i], cs[i+1:]...)
}

func (cs GiftCards) include(card *GiftCard) bool {
        for _, c := range cs {
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

func (this GiftCards) OfValue(v GiftCardValue) (cards GiftCards) {
	for _, card := range this {
		if card.Value == v {
			cards = append(cards, card)
		}
	}
	return cards
}

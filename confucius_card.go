package confucius

import (
	"encoding/gob"

	"github.com/SlothNinja/sn"
)

func init() {
	gob.RegisterName("game.ConCards", make(ConCards, 0))
}

type ConCard struct {
	Coins    int
	Revealed bool
}
type ConCards []*ConCard

func (c *ConCard) equal(card *ConCard) bool {
	return c != nil && card != nil && c.Coins == card.Coins
}

func NewConDeck(nplayers int) (deck ConCards) {
	for i := 0; i < 22-nplayers; i++ {
		deck = append(deck, &ConCard{Coins: 1}, &ConCard{Coins: 2}, &ConCard{Coins: 3})
	}
	return
}

func (this *ConCards) AppendNß(v, n int) {
	*this = this.AppendN(v, n)
}

func (this ConCards) AppendN(v, n int) (cs ConCards) {
	cs = this
	for i := 0; i < n; i++ {
		cs = append(cs, &ConCard{Coins: v})
	}
	return
}

func (this *ConCards) Append(cards ...*ConCard) {
	*this = this.AppendS(cards...)
}

func (cs ConCards) AppendS(cards ...*ConCard) ConCards {
	if len(cards) == 0 {
		return cs
	}
	return append(cs, cards...)
}

func (this *ConCards) Remove(cards ...*ConCard) {
	*this = this.RemoveS(cards...)
}

func (this ConCards) RemoveS(cards ...*ConCard) (cs ConCards) {
	cs = this
	for _, c := range cards {
		cs = cs.remove(c)
	}
	return
}

func (cs ConCards) remove(card *ConCard) ConCards {
	cards := cs
	for i, c := range cs {
		if c.equal(card) {
			return cards.removeAt(i)
		}
	}
	return cs
}

func (cs ConCards) removeAt(i int) ConCards {
	return append(cs[:i], cs[i+1:]...)
}

func (this *ConCards) Draw() (card *ConCard) {
	*this, card = this.DrawS()
	return
}

func (this ConCards) DrawS() (cs ConCards, card *ConCard) {
	i := sn.MyRand.Intn(len(this))
	card = this[i]
	cs = this.removeAt(i)
	return
}

func (this ConCards) Licenses() (count int) {
	for _, card := range this {
		count += card.Licenses()
	}
	return count
}

func (this ConCards) Coins() (count int) {
	for _, card := range this {
		count += card.Coins
	}
	return count
}

func (this ConCards) Count(v int) (count int) {
	for _, card := range this {
		if card.Coins == v {
			count += 1
		}
	}
	return count
}

func (this ConCard) Licenses() int {
	return 4 - this.Coins
}

func (this ConCards) Reveal() {
	for i := range this {
		this[i].Revealed = true
	}
}

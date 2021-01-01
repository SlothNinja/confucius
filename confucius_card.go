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

func NewConDeck(nplayers int) ConCards {
	var deck ConCards
	for i := 0; i < 22-nplayers; i++ {
		deck = append(deck, &ConCard{Coins: 1}, &ConCard{Coins: 2}, &ConCard{Coins: 3})
	}
	return deck
}

func (cds *ConCards) AppendNß(v, n int) {
	*cds = cds.AppendN(v, n)
}

func (cds ConCards) AppendN(v, n int) ConCards {
	cs := cds
	for i := 0; i < n; i++ {
		cs = append(cs, &ConCard{Coins: v})
	}
	return cs
}

func (cds *ConCards) Append(cards ...*ConCard) {
	*cds = cds.AppendS(cards...)
}

func (cds ConCards) AppendS(cards ...*ConCard) ConCards {
	if len(cards) == 0 {
		return cds
	}
	return append(cds, cards...)
}

func (cds *ConCards) Remove(cards ...*ConCard) {
	*cds = cds.RemoveS(cards...)
}

func (cds ConCards) RemoveS(cards ...*ConCard) ConCards {
	cs := cds
	for _, c := range cards {
		cs = cs.remove(c)
	}
	return cs
}

func (cds ConCards) remove(card *ConCard) ConCards {
	cards := cds
	for i, c := range cds {
		if c.equal(card) {
			return cards.removeAt(i)
		}
	}
	return cds
}

func (cds ConCards) removeAt(i int) ConCards {
	return append(cds[:i], cds[i+1:]...)
}

func (cds *ConCards) Draw() *ConCard {
	var card *ConCard
	*cds, card = cds.DrawS()
	return card
}

func (cds ConCards) DrawS() (ConCards, *ConCard) {
	i := sn.MyRand.Intn(len(cds))
	card := cds[i]
	cs := cds.removeAt(i)
	return cs, card
}

func (cds ConCards) Licenses() int {
	count := 0
	for _, card := range cds {
		count += card.Licenses()
	}
	return count
}

func (cds ConCards) Coins() int {
	count := 0
	for _, card := range cds {
		count += card.Coins
	}
	return count
}

func (cds ConCards) Count(v int) int {
	count := 0
	for _, card := range cds {
		if card.Coins == v {
			count += 1
		}
	}
	return count
}

func (cds ConCard) Licenses() int {
	return 4 - cds.Coins
}

func (cds ConCards) Reveal() {
	for i := range cds {
		cds[i].Revealed = true
	}
}

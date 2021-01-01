package confucius

import (
	"html/template"

	"github.com/SlothNinja/sn"
)

type EmperorCardType int

type EmperorCard struct {
	Type     EmperorCardType
	Revealed bool
}

type EmperorCards []*EmperorCard

const (
	Cash EmperorCardType = iota
	FreeGift
	ExtraAction
	BingbuBribery
	HubuBribery
	GongbuBribery
	AnyBribery1
	AnyBribery2
	EmperorInsulted
	RecruitFreeArmy
)

var emperorCardTypes = []EmperorCardType{Cash, FreeGift, ExtraAction, BingbuBribery, HubuBribery, GongbuBribery, AnyBribery1, AnyBribery2, EmperorInsulted, RecruitFreeArmy}
var briberyCardTypes = []EmperorCardType{BingbuBribery, HubuBribery, GongbuBribery, AnyBribery1, AnyBribery2}

func (g *Game) BriberyCards() EmperorCards {
	cards := make(EmperorCards, len(briberyCardTypes))
	for i, cardType := range briberyCardTypes {
		cards[i] = &EmperorCard{Type: cardType}
	}
	return cards
}

func NewEmperorDeck() EmperorCards {
	deck := make(EmperorCards, len(emperorCardTypes))
	for i, t := range emperorCardTypes {
		deck[i] = NewEmperorCard(t)
	}
	return deck
}

func (c *EmperorCard) equal(card *EmperorCard) bool {
	return c != nil && card != nil && c.Type == card.Type
}

func (c *EmperorCard) hasType(t EmperorCardType) bool {
	return c.Type == t
}

func NewEmperorCard(t EmperorCardType) *EmperorCard {
	card := new(EmperorCard)
	card.Type = t
	return card
}

func (cds *EmperorCards) Append(cards ...*EmperorCard) {
	*cds = cds.AppendS(cards...)
}

func (cds EmperorCards) AppendS(cards ...*EmperorCard) EmperorCards {
	return append(cds, cards...)
}

func (cds *EmperorCards) Draw() *EmperorCard {
	var c *EmperorCard
	*cds, c = cds.DrawS()
	return c
}

func (cds EmperorCards) DrawS() (EmperorCards, *EmperorCard) {
	i := sn.MyRand.Intn(len(cds))
	c := cds[i]
	deck := append(cds[:i], cds[i+1:]...)
	return deck, c
}

func (cds EmperorCards) Reveal() {
	for i := range cds {
		cds[i].Revealed = true
	}
}

func (cds *EmperorCards) Remove(cards ...*EmperorCard) {
	*cds = cds.RemoveS(cards...)
}

func (cds EmperorCards) RemoveS(cards ...*EmperorCard) EmperorCards {
	es := cds
	for _, c := range cards {
		es = es.remove(c)
	}
	return es
}

func (cds EmperorCards) remove(card *EmperorCard) EmperorCards {
	es := cds
	for i, c := range es {
		if c.equal(card) {
			es = es.removeAt(i)
			break
		}
	}
	return es
}

func (cds EmperorCards) removeAt(i int) EmperorCards {
	return append(cds[:i], cds[i+1:]...)
}

var emperorCardTitleStrings = map[EmperorCardType]string{
	Cash:            "Cash",
	FreeGift:        "Gift",
	ExtraAction:     "Extra Action",
	BingbuBribery:   "Bribery in Bingbu Ministry",
	HubuBribery:     "Bribery in Hubu Ministry",
	GongbuBribery:   "Bribery in Gongbu Ministry",
	AnyBribery1:     "Bribery in Any Ministry",
	AnyBribery2:     "Bribery in Any Ministry",
	EmperorInsulted: "Emperor Insulted",
	RecruitFreeArmy: "Recruit an Army",
}

func (cds *EmperorCard) Title() string {
	return emperorCardTitleStrings[cds.Type]
}

var emperorCardDescriptionStrings = map[EmperorCardType]string{
	Cash:     "Take 4 cards from the Confucius deck.",
	FreeGift: "Choose one of your \"not bought\" gifts and turn it into a \"bought\" gift for no cost.",
	ExtraAction: `Take any Actions Box action except the Imperial Favour action without playing any action cubes.
                      This can be a 0, 1 or 2 cube action and can be a repeat of a previous action.
                      All other restrictions on the actions apply.`,
	BingbuBribery: `<p><i>Gift obligations do not apply to this action.</i></p>
                        <p>You may bribe an official in Bingbu, or if Bingbu has been resolved, in any unresolved ministry.</p>
                        <p><b>EITHER:</b> Place an unsecured marker on an unbribed official;</p>
                        <p><b>OR:</b> Choose an unsecured official owned by another player and pay the cash shown on the tile.
                        If you have one or more bribed officials in Hubu, reduce the cost by 1.
                        The other player's unsecured marker is replaced with your unsecured markers.</p>`,
	HubuBribery: `<p><i>Gift obligations do not apply to this action.</i></p>
                      <p>You may bribe an official in Hubu, or if Hubu has been resolved, in any unresolved ministry.</p>
                      <p><b>EITHER:</b> Place an unsecured marker on an unbribed official;</p>
                      <p><b>OR:</b> Choose an unsecured official owned by another player and pay the cash shown on the tile.
                      If you have one or more bribed officials in Hubu, reduce the cost by 1.
                      The other player's unsecured marker is replaced with your unsecured markers.</p>`,
	GongbuBribery: `<p><i>Gift obligations do not apply to this action.</i></p>
                  <p>You may bribe an official in Gongbu, or if Gongbu has been resolved, in any unresolved ministry.</p>
                  <p><b>EITHER:</b> Place an unsecured marker on an unbribed official;</p>
                  <p><b>OR:</b> Choose an unsecured official owned by another player and pay the cash shown on the tile.
                  If you have one or more bribed officials in Hubu, reduce the cost by 1.
                  The other player's unsecured marker is replaced with your unsecured markers.</p>`,
	AnyBribery1: `<p><i>Gift obligations do not apply to this action.</i></p>
                  <p>You may bribe an official in any unresolved ministry.</p>
                  <p><b>EITHER:</b> Place an unsecured marker on an unbribed official;</p>
                  <p><b>OR:</b> Choose an unsecured official owned by another player and pay the cash shown on the tile.
                  If you have one or more bribed officials in Hubu, reduce the cost by 1.
                  The other player's unsecured marker is replaced with your unsecured markers.</p>`,
	AnyBribery2: `<p><i>Gift obligations do not apply to this action.</i></p>
                  <p>You may bribe an official in any unresolved ministry.</p>
                  <p><b>EITHER:</b> Place an unsecured marker on an unbribed official;</p>
                  <p><b>OR:</b> Choose an unsecured official owned by another player and pay the cash shown on the tile.
                  If you have one or more bribed officials in Hubu, reduce the cost by 1.
                  The other player's unsecured marker is replaced with your unsecured markers.</p>`,
	EmperorInsulted: `<p><i>A minor foreign ruler has insulted the Emperor.
                You can avenge him by conquering his lands.</i></p>
                <p>Place this card face up in your playing area and move one of your
                armies from the military colonies to the card for no cost.</p>
                <p><b>Gain 2 VP.</b></p>`,
	RecruitFreeArmy: "Place one of your armies in the military colonies for no cost.",
}

func (cds *EmperorCard) Description() template.HTML {
	return template.HTML(emperorCardDescriptionStrings[cds.Type])
}

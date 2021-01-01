package confucius

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"sort"

	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("ConfuciusPlayer", NewPlayer())
}

type Player struct {
	*game.Player
	Log game.GameLog

	TakenCommercial bool
	ActionCubes     int
	Junks           int
	OnVoyage        int
	Armies          int
	RecruitedArmies int
	ConCardHand     ConCards
	GiftCardHand    GiftCards
	GiftsBought     GiftCards
	GiftsReceived   GiftCards
	EmperorHand     EmperorCards
}

type Players []*Player

// sort.Interface interface
func (ps Players) Len() int { return len(ps) }

func (ps Players) Swap(i, j int) { ps[i], ps[j] = ps[j], ps[i] }

type ByScore struct{ Players }

func (bs ByScore) Less(i, j int) bool { return bs.Players[i].Score < bs.Players[j].Score }

type ByAll struct{ Players }

func (ba ByAll) Less(i, j int) bool {
	return ba.Players[i].compare(ba.Players[j]) == game.LessThan
}

type Reverse struct{ sort.Interface }

func (r Reverse) Less(i, j int) bool { return r.Interface.Less(j, i) }

const NoPlayerID int = -1

func (p *Player) Game() *Game {
	return p.Player.Game().(*Game)
}

func (p *Player) ScoreTrack() int { return p.Score % 31 }

func (p *Player) compare(player *Player) game.Comparison {
	if c := p.CompareByScore(player.Player); c != game.EqualTo {
		return c
	}

	if c := p.compareByAdmiral(player); c != game.EqualTo {
		return c
	}

	if c := p.compareByChiefMinister(player); c != game.EqualTo {
		return c
	}

	if c := p.compareByGeneral(player); c != game.EqualTo {
		return c
	}

	return game.EqualTo
}

func (p *Player) IsAdmiral() bool {
	return p.Equal(p.Game().Admiral())
}

func (p *Player) compareByAdmiral(player *Player) game.Comparison {
	if p.IsAdmiral() {
		return game.GreaterThan
	}

	if player.IsAdmiral() {
		return game.LessThan
	}

	return game.EqualTo
}

func (p *Player) IsChiefMinister() bool {
	return p.Equal(p.Game().ChiefMinister())
}

func (p *Player) compareByChiefMinister(player *Player) game.Comparison {
	if p.IsChiefMinister() {
		return game.GreaterThan
	}

	if player.IsChiefMinister() {
		return game.LessThan
	}

	return game.EqualTo
}

func (p *Player) IsGeneral() bool {
	return p.Equal(p.Game().General())
}

func (p *Player) compareByGeneral(player *Player) game.Comparison {
	if p.IsGeneral() {
		return game.GreaterThan
	}

	if player.IsGeneral() {
		return game.LessThan
	}

	return game.EqualTo
}

func (ps Players) removeAt(i int) Players {
	return append(ps[:i], ps[i+1:]...)
}

func (client Client) determinePlaces(c *gin.Context, g *Game) (contest.Places, error) {
	// sort players by score
	players := g.Players()
	sort.Sort(Reverse{ByAll{players}})
	g.setPlayers(players)

	places := make(contest.Places, 0)
	if g.AdmiralVariant {
		winner := g.Players()[0]
		if g.Players()[0].Score == g.Players()[1].Score {
			// Admiral win.  Find Admiral and place at g.Players[0]
			winner = g.Admiral()
			ps := Players{}
			for i, p := range g.Players() {
				if p.IsAdmiral() {
					ps = g.Players().removeAt(i)
					break
				}
			}
			g.setPlayers(append(Players{winner}, ps...))
		}
	}
	for i, p1 := range g.Players() {
		rmap := make(contest.ResultsMap, 0)
		results := make(contest.Results, 0)
		for j, p2 := range g.Players() {
			r, err := client.Rating.For(c, p2.User(), g.Type)
			if err != nil {
				return nil, err
			}
			result := &contest.Result{
				GameID: g.ID(),
				Type:   g.Type,
				R:      r.R,
				RD:     r.RD,
			}
			switch c := p1.compare(p2); {
			case i == j:
			case i == 0 && g.AdmiralVariant:
				result.Outcome = 1
				results = append(results, result)
			case c == game.EqualTo:
				result.Outcome = 0.5
				results = append(results, result)
			case c == game.LessThan:
				result.Outcome = 0
				results = append(results, result)
			case c == game.GreaterThan:
				result.Outcome = 1
				results = append(results, result)
			}
		}
		rmap[p1.User().Key] = results
		places = append(places, rmap)
	}
	return places, nil
}

func (p *Player) init(gr game.Gamer) {
	p.SetGame(gr)

	g, ok := gr.(*Game)
	if !ok {
		return
	}

	for _, card := range p.GiftCardHand {
		card.SetGame(g)
	}

	for _, card := range p.GiftsBought {
		card.SetGame(g)
	}

	for _, card := range p.GiftsReceived {
		card.SetGame(g)
	}

	for _, entry := range p.Log {
		entry.Init(g)
	}
}

func (p *Player) beginningOfTurnReset() {
	p.clearActions()
}

func (g *Game) beginningOfPhaseReset() {
	for _, p := range g.Players() {
		p.clearActions()
	}
}

func NewPlayer() *Player {
	p := new(Player)
	p.Player = game.NewPlayer()
	return p
}

func CreatePlayer(g *Game, u *user.User) *Player {
	p := NewPlayer()
	p.SetID(int(len(g.Players())))
	p.SetGame(g)

	colorMap := g.DefaultColorMap()
	p.SetColorMap(make(color.Colors, g.NumPlayers))

	for i := 0; i < g.NumPlayers; i++ {
		index := (i - p.ID()) % g.NumPlayers
		if index < 0 {
			index += g.NumPlayers
		}
		color := colorMap[index]
		p.ColorMap()[i] = color
	}

	p.ConCardHand = ConCards{&ConCard{Coins: 1}, &ConCard{Coins: 2}, &ConCard{Coins: 3}}

	p.ConCardHand.Reveal()
	p.NewGiftCardHand()
	p.NewGiftsBought()
	p.Armies = 6
	return p
}

func (p *Player) NewGiftCardHand() {
	values := []GiftCardValue{Tile, Vase, Coat, Necklace, Junk}
	newHand := make(GiftCards, len(values))
	for i, value := range values {
		newHand[i] = new(GiftCard)
		newHand[i].SetGame(p.Game())
		newHand[i].Value = value
		newHand[i].setPlayer(p)
	}
	p.GiftCardHand = newHand
}

func (p *Player) NewGiftsBought() {
	card := new(GiftCard)
	card.SetGame(p.Game())
	card.Value = Hanging
	card.setPlayer(p)
	p.GiftsBought = GiftCards{card}
}

func (p *Player) displayBarrel() string {
	return fmt.Sprintf(`<img src="/images/confucius/%s-barrel-shadowed.png" alt="%s Barrel" />`, p.Color(), p.Color())
}

func (p *Player) DisplayBarrel() template.HTML {
	return template.HTML(p.displayBarrel())
}

func (p *Player) DisplaySecuredBarrel() template.HTML {
	result := p.displayBarrel()
	result += fmt.Sprintf(`<div class="text %s">S</div>`, p.TextColor())
	return template.HTML(result)
}

func (p *Player) DisplayTempBarrel() template.HTML {
	result := p.displayBarrel()
	result += fmt.Sprintf(`<div class="text %s">T</div>`, p.TextColor())
	return template.HTML(result)
}

func (p *Player) DisplayArmies() string {
	s := ""
	for i := 0; i < p.Armies; i++ {
		s += fmt.Sprintf("<img src=\"/images/confucius/%s-army-shadowed.png\" alt=\"%s Army\"/>", p.Color(), p.Color())
	}
	return s
}

func (ps Players) Users() user.Users {
	us := make(user.Users, len(ps))
	for i, player := range ps {
		us[i] = player.User()
	}
	return us
}

func (ps Players) Include(p *Player) bool {
	for _, player := range ps {
		if player.Equal(p) {
			return true
		}
	}
	return false
}

func (ps Players) IncludeUser(u *user.User) bool {
	for _, p := range ps {
		if p.User().Equal(u) {
			return true
		}
	}
	return false
}

func (ps Players) allPassed() bool {
	for _, p := range ps {
		if !p.Passed {
			return false
		}
	}
	return true
}

func (ps Players) allPerformedAction() bool {
	for _, p := range ps {
		if !p.PerformedAction {
			return false
		}
	}
	return true
}

func (p *Player) CardCount(v int) int {
	count := 0
	for _, c := range p.ConCardHand {
		if c.Coins == v {
			count += 1
		}
	}
	return count
}

func (p *Player) clearActions() {
	p.PerformedAction = false
	p.Passed = false
	p.Log = make(game.GameLog, 0)
}

func (p *Player) HubuDiscount() int {
	if p.HasInfluenceIn(p.Game().Ministries[Hubu]) {
		return 1
	}
	return 0
}

func (p *Player) junkCostFor(j int) int {
	discountedJunks := []int{0, 1, 2, 4, 7}
	normalJunks := []int{0, 1, 3, 6, 10}

	if p.HasInfluenceIn(p.Game().Ministries[Gongbu]) {
		return discountedJunks[j]
	}
	return normalJunks[j]
}

func (p *Player) armyCost() int {
	if p.HasInfluenceIn(p.Game().Ministries[Bingbu]) {
		return 4
	}
	return 6
}

func (p *Player) giftFrom(player *Player) *GiftCard {
	if p.GiftsReceived != nil {
		for _, g := range p.GiftsReceived {
			if g.isFrom(player) {
				return g
			}
		}
	}
	return nil
}

func (p *Player) hasGiftFrom(player *Player) bool {
	return p.giftFrom(player) != nil
}

func (p *Player) GiftsGiven() int {
	count := 0
	for _, player := range p.Game().Players() {
		if player.NotEqual(p) && player.hasGiftFrom(p) {
			count += 1
		}
	}
	return count
}

func (p *Player) GetGift(v GiftCardValue) *GiftCard {
	for _, g := range p.GiftCardHand {
		if g.Value == v {
			return g
		}
	}
	return nil
}

func (p *Player) GetBoughtGift(v GiftCardValue) *GiftCard {
	for _, g := range p.GiftsBought {
		if g.Value == v {
			return g
		}
	}
	return nil
}

func (p *Player) GetEmperorCard(v EmperorCardType) *EmperorCard {
	for _, c := range p.EmperorHand {
		if c.Type == v {
			return c
		}
	}
	return nil
}

func (p *Player) cancelGiftFrom(player *Player) *GiftCard {
	gift := p.giftFrom(player)
	if gift != nil {
		p.GiftsReceived.Remove(gift)
		return gift
	}
	return nil
}

func (p *Player) GiveGiftTo(gift *GiftCard, recipient *Player) bool {
	// Remove Gift From Those Bought
	p.GiftsBought.Remove(gift)

	// Cancel Prior Gift Given To Recipient
	recipient.cancelGiftFrom(p)

	// Place Gift With Recipient
	recipient.GiftsReceived.Append(gift)

	// Cancel Gift From Recipient
	if p.hasGiftFrom(recipient) && p.giftFrom(recipient).Value < gift.Value {
		p.cancelGiftFrom(recipient)
		return true
	}
	return false
}

func (p *Player) Equal(op *Player) bool {
	return p != nil && op != nil && p.Player.Equal(op)
}

func (p *Player) NotEqual(op *Player) bool {
	return !p.Equal(op)
}

func (p *Player) influenceIn(m *Ministry) int {
	count := 0
	for _, t := range m.Officials {
		if t.PlayerID != NoPlayerID && t.PlayerID == p.ID() {
			count += 1
		}
	}
	return count
}

func (p *Player) CostFor(tile *OfficialTile) int {
	return tile.Cost - p.HubuDiscount()
}

func (p *Player) HasInfluenceIn(m *Ministry) bool {
	return p.influenceIn(m) > 0
}

func (p *Player) hasTempInfluence() bool {
	m := p.Game().ministryInProgress()
	for _, o := range m.Officials {
		if o.TempPlayer() != nil && o.TempPlayer().Equal(p) {
			return true
		}
	}
	return false
}

func (p *Player) hasJunks() bool {
	return p.Junks > 0
}

func (p *Player) hasLicenses() bool {
	return p.ConCardHand.Licenses() > 0
}

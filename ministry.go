package confucius

import (
	"encoding/gob"

	"github.com/SlothNinja/sn"
)

func init() {
	gob.RegisterName("*game.resolvedMinistryEntry", new(resolvedMinistryEntry))
}

type MinistryID int
type MinistryIDS []MinistryID

const (
	Bingbu MinistryID = iota
	Hubu
	Gongbu
)

var ministeryIDS = MinistryIDS{Bingbu, Hubu, Gongbu}
var ministryIDStrings = map[MinistryID]string{Bingbu: "Bingbu", Hubu: "Hubu", Gongbu: "Gongbu"}

type Ministry struct {
	game          *Game
	ID            MinistryID
	Officials     OfficialTiles
	MinisterChit  MinistryChit
	SecretaryChit MinistryChit
	Resolved      bool
	InProgress    bool
	MinisterID    int
	SecretaryID   int
}
type Ministries map[MinistryID]*Ministry

func (g *Game) MinistryIDS() MinistryIDS {
	return ministeryIDS
}

func (ms Ministries) Include(ministry *Ministry) bool {
	for _, m := range ms {
		if m.ID == ministry.ID {
			return true
		}
	}
	return false
}

func (m *Ministry) Game() *Game {
	return m.game
}

func (m *Ministry) SetGame(game *Game) {
	m.game = game
}

func (m *Ministry) Name() string {
	return ministryIDStrings[m.ID]
}

func (m *Ministry) Minister() *Player {
	if m.MinisterID != NoPlayerID {
		return m.Game().PlayerByID(m.MinisterID)
	}
	return nil
}

func (m *Ministry) setMinister(p *Player) {
	switch {
	case p == nil:
		m.MinisterID = NoPlayerID
	default:
		m.MinisterID = p.ID()
	}
}

func (m *Ministry) Secretary() *Player {
	if m.SecretaryID != NoPlayerID {
		return m.Game().PlayerByID(m.SecretaryID)
	}
	return nil
}

func (m *Ministry) setSecretary(p *Player) {
	switch {
	case p == nil:
		m.SecretaryID = NoPlayerID
	default:
		m.SecretaryID = p.ID()
	}
}

func (m *Ministry) init(g *Game) {
	m.SetGame(g)
	for _, official := range m.Officials {
		official.ministry = m
		official.game = m.game
	}
}

type MinistryChit int

func (mc MinistryChit) Value() int {
	return int(mc)
}

type MinistryChits []MinistryChit

func (g *Game) setMinistryChits() {
	mcs := []MinistryChit{4, 4, 5, 5, 6, 6, 7, 7, 8, 8}
	for _, m := range g.Ministries {
		m.setMinistryChits(mcs)
	}
}

func (m *Ministry) setMinistryChits(mcs MinistryChits) {
	i := sn.MyRand.Intn(len(mcs))
	chit1 := mcs[i]
	mcs = append(mcs[:i], mcs[i+1:]...)

	i = sn.MyRand.Intn(len(mcs))
	chit2 := mcs[i]
	mcs = append(mcs[:i], mcs[i+1:]...)

	if chit1 > chit2 {
		m.MinisterChit = chit1
		m.SecretaryChit = chit2
	} else {
		m.MinisterChit = chit2
		m.SecretaryChit = chit1
	}
}

type Seniority int
type Seniorities []Seniority

func (ss Seniorities) Include(seniority Seniority) bool {
	for _, s := range ss {
		if s == seniority {
			return true
		}
	}
	return false
}

func (g *Game) Seniorities() Seniorities {
	return Seniorities{0, 1, 2, 3, 4, 5, 6, 7}
}

func (s Seniority) Equal(seniority Seniority) bool {
	return s == seniority
}

type OfficialTile struct {
	game      *Game
	ministry  *Ministry
	Cost      int
	Variant   VariantID
	PlayerID  int
	TempID    int
	Seniority Seniority
	Secured   bool
}
type OfficialTiles map[Seniority]*OfficialTile
type OfficialsDeck []*OfficialTile

func (o *OfficialTile) Game() *Game {
	return o.game
}

func newOfficialTile() *OfficialTile {
	return &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID}
}

func (o *OfficialTile) Player() *Player {
	if o.PlayerID != NoPlayerID {
		return o.game.PlayerByID(o.PlayerID)
	}
	return nil
}

func (o *OfficialTile) setPlayer(p *Player) {
	switch {
	case p == nil:
		o.PlayerID = NoPlayerID
	default:
		o.PlayerID = p.ID()
	}
}

func (o *OfficialTile) CostFor(p *Player) int {
	return o.Cost - p.HubuDiscount()
}

func (o *OfficialTile) Bribed() bool {
	return !o.NotBribed()
}

func (o *OfficialTile) NotBribed() bool {
	return o.PlayerID == NoPlayerID
}

func (o *OfficialTile) TempPlayer() *Player {
	if o.TempID != NoPlayerID {
		return o.game.PlayerByID(o.TempID)
	}
	return nil
}

func (o *OfficialTile) setTempPlayer(p *Player) {
	switch {
	case p == nil:
		o.TempID = NoPlayerID
	default:
		o.TempID = p.ID()
	}
}

type CandidateTile struct {
	*OfficialTile
	OtherPlayerID    int
	PlayerCards      ConCards
	OtherPlayerCards ConCards
}

type CandidateTiles []*CandidateTile

func newCandidateTile() *CandidateTile {
	tile := new(CandidateTile)
	tile.OfficialTile = newOfficialTile()
	tile.OtherPlayerID = NoPlayerID
	return tile
}

func (cs *CandidateTiles) Swap(i1, i2 int) {
	tiles := *cs
	tiles[i1], tiles[i2] = tiles[i2], tiles[i1]
	*cs = tiles
}

func (c *CandidateTile) OtherPlayer() *Player {
	if c.OtherPlayerID != NoPlayerID {
		return c.Game().PlayerByID(c.OtherPlayerID)
	}
	return nil
}

func (c *CandidateTile) setOtherPlayer(p *Player) {
	switch {
	case p == nil:
		c.OtherPlayerID = NoPlayerID
	default:
		c.OtherPlayerID = p.ID()
	}
}

func (c *CandidateTile) hasTwoPlayers() bool {
	return c.Player() != nil && c.OtherPlayer() != nil
}

func (c *CandidateTile) hasOnePlayer() bool {
	return (c.Player() != nil || c.OtherPlayer() != nil) && !c.hasTwoPlayers()
}

func (c *CandidateTile) hasTwoSamePlayers() bool {
	return c.hasTwoPlayers() && c.Player().Equal(c.OtherPlayer())
}

func (c *CandidateTile) Playable() bool {
	for _, ministry := range c.Game().MinistriesFor(c) {
		if !ministry.Resolved {
			return true
		}
	}
	return false
}

func (od *OfficialsDeck) Draw() *OfficialTile {
	var tile *OfficialTile
	*od, tile = od.DrawS()
	return tile
}

func (od OfficialsDeck) DrawS() (OfficialsDeck, *OfficialTile) {
	var tiles OfficialsDeck
	var tile *OfficialTile

	i := Seniority(sn.MyRand.Intn(len(od)))
	tile = od[i]
	tiles = append(od[:i], od[i+1:]...)
	return tiles, tile
}

func (g *Game) OfficialTiles() OfficialsDeck {
	var tiles OfficialsDeck
	for _, m := range g.Ministries {
		for _, official := range m.Officials {
			tiles = append(tiles, official)
		}
	}
	return tiles
}

func (g *Game) CreateMinistries() {
	ids := []MinistryID{Bingbu, Hubu, Gongbu}
	g.Ministries = make(Ministries, len(ids))
	for _, id := range ids {
		official3 := g.OfficialsDeck.Draw()
		official3.Seniority = 3
		official4 := g.OfficialsDeck.Draw()
		official4.Seniority = 4
		official5 := g.OfficialsDeck.Draw()
		official5.Seniority = 5
		g.Ministries[id] = &Ministry{
			ID:          id,
			Officials:   OfficialTiles{3: official3, 4: official4, 5: official5},
			MinisterID:  NoPlayerID,
			SecretaryID: NoPlayerID,
		}
	}
	g.setMinistryChits()
}

type VariantID int

const (
	NoOfficial VariantID = iota
	First
	Second
	Third
	Fourth
	Fifth
	BingbuCandidate
	HubuCandidate
	GongbuCandidate
	AnyCandidate1
	AnyCandidate2
	AnyCandidate3
	TileBack
)

func (g *Game) VariantIDS() []VariantID {
	return []VariantID{NoOfficial, First, Second, Third, Fourth, Fifth, BingbuCandidate, HubuCandidate,
		GongbuCandidate, AnyCandidate1, AnyCandidate2, AnyCandidate3, TileBack}
}

func (v VariantID) Equal(id VariantID) bool {
	return v == id
}

func NewOfficialsDeck() OfficialsDeck {
	var deck OfficialsDeck
	for _, variant := range []VariantID{First, Second, Third} {
		deck = append(deck, &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Cost: 2, Variant: variant})
	}
	for _, variant := range []VariantID{First, Second, Third, Fourth} {
		deck = append(deck, &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Cost: 3, Variant: variant})
	}
	for _, variant := range []VariantID{First, Second, Third, Fourth, Fifth} {
		deck = append(deck, &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Cost: 4, Variant: variant})
	}
	for _, variant := range []VariantID{First, Second, Third, Fourth, Fifth} {
		deck = append(deck, &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Cost: 5, Variant: variant})
	}
	for _, variant := range []VariantID{First, Second, Third, Fourth} {
		deck = append(deck, &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Cost: 6, Variant: variant})
	}
	for _, variant := range []VariantID{First, Second, Third} {
		deck = append(deck, &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Cost: 7, Variant: variant})
	}
	return deck
}

func (g *Game) CreateCandidates() {
	g.Candidates = CandidateTiles{
		&CandidateTile{OtherPlayerID: NoPlayerID, OfficialTile: &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Variant: BingbuCandidate}},
		&CandidateTile{OtherPlayerID: NoPlayerID, OfficialTile: &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Variant: HubuCandidate}},
		&CandidateTile{OtherPlayerID: NoPlayerID, OfficialTile: &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Variant: GongbuCandidate}},
		&CandidateTile{OtherPlayerID: NoPlayerID, OfficialTile: &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Variant: AnyCandidate1}},
		&CandidateTile{OtherPlayerID: NoPlayerID, OfficialTile: &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Variant: AnyCandidate2}},
		&CandidateTile{OtherPlayerID: NoPlayerID, OfficialTile: &OfficialTile{PlayerID: NoPlayerID, TempID: NoPlayerID, Variant: AnyCandidate3}},
	}

	// Shuffle first three candidates
	for i := 0; i < 3; i++ {
		ri := i + sn.MyRand.Intn(3-i)
		g.Candidates[i], g.Candidates[ri] = g.Candidates[ri], g.Candidates[i]
	}
}

func (m *Ministry) MarkerCount() int {
	count := 0
	for _, official := range m.Officials {
		if official.Bribed() {
			count += 1
		}
	}
	return count
}

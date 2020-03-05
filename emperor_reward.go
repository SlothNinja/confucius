package confucius

import (
	"encoding/gob"
	"fmt"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.takeCashEntry", new(takeCashEntry))
	gob.RegisterName("*game.takeGiftEntry", new(takeGiftEntry))
	gob.RegisterName("*game.takeArmyEntry", new(takeArmyEntry))
	gob.RegisterName("*game.takeExtraActionEntry", new(takeExtraActionEntry))
	gob.RegisterName("*game.avengeEmperorEntry", new(avengeEmperorEntry))
	gob.RegisterName("*game.takeBriberyRewardEntry", new(takeBriberyRewardEntry))
}

func (g *Game) EnableEmperorReward(c *gin.Context) bool {
	return g.CUserIsCPlayerOrAdmin(c) && g.CurrentPlayer().canEmperorReward()
}

func (p *Player) canEmperorReward() bool {
	switch {
	case p.Game().Phase != Actions:
		return false
	case p.Game().ExtraAction:
		return false
	case len(p.EmperorHand) < 1:
		return false
	}
	return true
}

func (g *Game) takeCash(c *gin.Context) (tmpl string, a game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var cd *EmperorCard
	if cd, err = g.validateTakeCash(c); err != nil {
		a = game.None
		return
	}

	// Perform Take Cash Action
	cp := g.CurrentPlayer()
	cp.ConCardHand.Append(g.DrawConCard(), g.DrawConCard(), g.DrawConCard(), g.DrawConCard())
	cp.EmperorHand.Remove(cd)
	cp.PerformedAction = true

	// Discard Played Card
	g.EmperorDiscard.Append(cd)

	// Create Action Object for logging
	e := g.NewTakeCashEntry(cp)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	a = game.Cache
	return
}

type takeCashEntry struct {
	*Entry
}

func (g *Game) NewTakeCashEntry(p *Player) *takeCashEntry {
	e := new(takeCashEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *takeCashEntry) HTML() template.HTML {
	return restful.HTML("%s played Emperor's Reward card to take four Confucius cards.", e.Player().Name())
}

func (g *Game) validateTakeCash(c *gin.Context) (cd *EmperorCard, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err = g.validatePlayerAction(c); err != nil {
		return
	}

	if cd, err = g.getRewardCard(c); err != nil {
		return
	}

	if !cd.hasType(Cash) {
		err = sn.NewVError("You did not play the correct emperor's reward card for the selected action.")
	}
	return
}

func (g *Game) takeGift(c *gin.Context) (tmpl string, a game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var (
		cd *EmperorCard
		gc *GiftCard
	)

	if cd, gc, err = g.validateTakeGift(c); err != nil {
		a = game.None
		return
	}

	// Remove Gift From GiftCardHand
	cp := g.CurrentPlayer()
	cp.GiftCardHand.Remove(gc)

	// Place Gift With Those Bought
	cp.GiftsBought.Append(gc)
	cp.PerformedAction = true

	// Remove Played Card From Hand
	cp.EmperorHand.Remove(cd)

	// Discard Played Card
	g.EmperorDiscard.Append(cd)

	// Create Action Object for logging
	e := g.NewTakeGiftEntry(cp)
	e.Gift = gc

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	a = game.Cache
	return
}

type takeGiftEntry struct {
	*Entry
	Gift *GiftCard
}

func (g *Game) NewTakeGiftEntry(p *Player) *takeGiftEntry {
	e := new(takeGiftEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *takeGiftEntry) HTML() template.HTML {
	return restful.HTML("%s used Emperor's Reward card to take %d value gift (%s).",
		e.Player().Name(), e.Gift.Value, e.Gift.Name())
}

func (g *Game) validateTakeGift(c *gin.Context) (cd *EmperorCard, gc *GiftCard, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err = g.validatePlayerAction(c); err != nil {
		return
	}

	if cd, err = g.getRewardCard(c); err != nil {
		return
	}

	if !cd.hasType(FreeGift) {
		err = sn.NewVError("You did not play the correct emperor's reward card for the selected action.")
		return
	}

	var gv GiftCardValue
	if gv, err = g.getGiftValue(c, "take-gift"); err != nil {
		return
	}

	if gc = g.CurrentPlayer().GetGift(gv); gc == nil {
		err = sn.NewVError("Selected gift card is not available.")
	}
	return
}

func (g *Game) takeArmy(c *gin.Context) (tmpl string, a game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var cd *EmperorCard

	if cd, err = g.validateTakeArmy(c); err != nil {
		a = game.None
		return
	}

	// Recruit Army
	cp := g.CurrentPlayer()
	cp.Armies -= 1
	cp.RecruitedArmies += 1
	cp.PerformedAction = true

	// Remove Played Card From Hand
	cp.EmperorHand.Remove(cd)

	// Discard Played Card
	g.EmperorDiscard.Append(cd)

	// Create Action Object for logging
	e := g.NewTakeArmyEntry(cp)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	a = game.Cache
	return
}

type takeArmyEntry struct {
	*Entry
}

func (g *Game) NewTakeArmyEntry(p *Player) *takeArmyEntry {
	e := new(takeArmyEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *takeArmyEntry) HTML() template.HTML {
	return restful.HTML("%s played Emperor's Reward card to recruit an army.", e.Player().Name())
}

func (g *Game) validateTakeArmy(c *gin.Context) (cd *EmperorCard, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err = g.validatePlayerAction(c); err != nil {
		return
	}

	if cd, err = g.getRewardCard(c); err != nil {
		return
	}

	if !cd.hasType(RecruitFreeArmy) {
		err = sn.NewVError("You did not play the correct emperor's reward card for the selected action.")
	}

	if cp := g.CurrentPlayer(); !cp.hasArmies() {
		err = sn.NewVError("You have no armies to recruit.")
	}
	return
}

func (g *Game) takeExtraAction(c *gin.Context) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var cd *EmperorCard
	if cd, err = g.validateTakeExtraAction(c); err != nil {
		act = game.None
		return
	}

	// Setup For Extra Action
	g.ExtraAction = true

	// Remove Played Card From Hand
	cp := g.CurrentPlayer()
	cp.EmperorHand.Remove(cd)

	// Discard Played Card
	g.EmperorDiscard.Append(cd)

	// Create Action Object for logging
	e := g.NewTakeExtraActionEntry(cp)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	act = game.Cache
	return
}

type takeExtraActionEntry struct {
	*Entry
}

func (g *Game) NewTakeExtraActionEntry(p *Player) *takeExtraActionEntry {
	e := new(takeExtraActionEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *takeExtraActionEntry) HTML() template.HTML {
	return restful.HTML("%s played Emperor's Reward card to perform action without paying an action cube.", e.Player().Name())
}

func (g *Game) validateTakeExtraAction(c *gin.Context) (ec *EmperorCard, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err = g.validatePlayerAction(c); err != nil {
		return
	}

	if ec, err = g.getRewardCard(c); err != nil {
		return
	}

	if !ec.hasType(ExtraAction) {
		err = sn.NewVError("You did not play the correct emperor's reward card for the selected action.")
	}
	return
}

func (g *Game) avengeEmperor(c *gin.Context) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var eCard *EmperorCard
	if eCard, err = g.validateAvengeEmperor(c); err != nil {
		act = game.None
		return
	}

	// Commit Recruited Army and Score Points
	cp := g.CurrentPlayer()
	cp.RecruitedArmies -= 1
	cp.Score += 2
	g.SetAvenger(cp)
	cp.PerformedAction = true

	// Remove Played Card From Hand
	cp.EmperorHand.Remove(eCard)

	// Discard Played Card
	g.EmperorDiscard.Append(eCard)

	// Create Action Object for logging
	e := g.NewAvengeEmperorEntry(cp)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	act = game.Cache
	return
}

type avengeEmperorEntry struct {
	*Entry
}

func (g *Game) NewAvengeEmperorEntry(p *Player) *avengeEmperorEntry {
	e := new(avengeEmperorEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *avengeEmperorEntry) HTML() template.HTML {
	return restful.HTML("%s used Emperor's Reward card and army to avenge emperor.", e.Player().Name())
}

func (g *Game) validateAvengeEmperor(c *gin.Context) (eCard *EmperorCard, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err = g.validatePlayerAction(c); err != nil {
		return nil, err
	}

	if eCard, err = g.getRewardCard(c); err != nil {
		return
	}

	if !eCard.hasType(EmperorInsulted) {
		err = sn.NewVError("You did not play the correct emperor's reward card for the selected action.")
		return
	}

	if !g.CurrentPlayer().hasRecruitedArmies() {
		err = sn.NewVError("You have no recruited armies with which to avenge the Emperor.")
	}
	return
}

func (g *Game) takeBriberyReward(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cd, cs, ministry, o, err := g.validateBriberyReward(c)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	// Remove Played Card From Hand
	cp.EmperorHand.Remove(cd)

	// Discard Played Card
	g.EmperorDiscard.Append(cd)

	// Move played Confucius cards from hand to discard pile.
	cp.ConCardHand.Remove(cs...)
	g.ConDiscardPile.Append(cs...)

	// Update Bribed Official
	var otherPlayerID int
	if otherPlayer := o.Player(); otherPlayer != nil {
		otherPlayerID = otherPlayer.ID()
	} else {
		otherPlayerID = NoPlayerID
	}
	o.setPlayer(cp)

	// Create Action Object for logging
	e := g.NewTakeBriberyRewardEntry(cp)
	e.MinistryName = ministry.Name()
	e.Seniority = o.Seniority
	e.OtherPlayerID = otherPlayerID
	e.Played = cs

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, err
}

type takeBriberyRewardEntry struct {
	*Entry
	MinistryName string
	Seniority    Seniority
	Played       ConCards
}

func (g *Game) NewTakeBriberyRewardEntry(p *Player) *takeBriberyRewardEntry {
	e := new(takeBriberyRewardEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *takeBriberyRewardEntry) HTML() template.HTML {
	if e.OtherPlayer() == nil {
		return restful.HTML("%s used Emperor's Reward card to place unsecured marker on %s official having %d seniority.", e.Player().Name(), e.MinistryName, e.Seniority)
	}
	length := len(e.Played)
	return restful.HTML("%s used Emperor's Reward card and %d Confucius %s having %d coins to replace unsecured marker of %s on %s official having %d seniority.", e.Player().Name(), length, pluralize("card", length), e.Played.Coins(), e.OtherPlayer().Name(), e.MinistryName, e.Seniority)
}

func (g *Game) validateBriberyReward(c *gin.Context) (*EmperorCard, ConCards, *Ministry, *OfficialTile, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if _, err := g.validatePlayerAction(c); err != nil {
		return nil, nil, nil, nil, err
	}

	card, err := g.getRewardCard(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cards, err := g.getConCards(c, "take-bribery-reward")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	ministry, o, err := g.getMinistryAndOfficial(c, fmt.Sprintf("take-bribery-reward-official-%d", card.Type))
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cp := g.CurrentPlayer()
	validMininstry := false
	for _, m := range g.emperorsRewardMinistriesFor(card) {
		if ministry.Name() == m.Name() {
			validMininstry = true
			break
		}
	}

	if !validMininstry {
		return nil, nil, nil, nil, sn.NewVError("You must select a valid ministry for the selected card.")
	}

	switch {
	case o.Secured:
		return nil, nil, nil, nil, sn.NewVError("You must select an official that doesn't have a secured marker.")
	case cp.Equal(o.Player()):
		return nil, nil, nil, nil, sn.NewVError("You must select an official that doesn't have your marker.")
	case o.Bribed() && !cp.canAffordToBribe(o):
		return nil, nil, nil, nil, sn.NewVError("You selected cards having %d total coins, but you need %d coins to bribe the selected official.", cards.Coins(), cp.CostFor(o))
	}
	return card, cards, ministry, o, nil
}

func (p *Player) canEmperorRewardBribeIn(m *Ministry) bool {
	return m != nil && !m.Resolved && len(m.unbribedUnsecuredSpotsFor(p)) > 0
}

func (m *Ministry) unbribedUnsecuredSpotsFor(p *Player) []*OfficialTile {
	os := []*OfficialTile{}
	for _, o := range m.Officials {
		if o.NotBribed() || (o.Player().NotEqual(p) && !o.Secured) {
			os = append(os, o)
		}
	}
	return os
}

func (g *Game) emperorsRewardMinistriesFor(card *EmperorCard) Ministries {
	cp := g.CurrentPlayer()
	switch t := card.Type; {
	case t == BingbuBribery && cp.canEmperorRewardBribeIn(g.Ministries[Bingbu]):
		return Ministries{Bingbu: g.Ministries[Bingbu]}
	case t == HubuBribery && cp.canEmperorRewardBribeIn(g.Ministries[Hubu]):
		return Ministries{Hubu: g.Ministries[Hubu]}
	case t == GongbuBribery && cp.canEmperorRewardBribeIn(g.Ministries[Gongbu]):
		return Ministries{Gongbu: g.Ministries[Gongbu]}
	}
	ms := Ministries{}
	for _, m := range g.Ministries {
		if cp.canEmperorRewardBribeIn(m) {
			ms[m.ID] = m
		}
	}
	return ms
}

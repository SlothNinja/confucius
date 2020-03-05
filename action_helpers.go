package confucius

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (g *Game) getForeignLandBox(c *gin.Context, formValue string) (*ForeignLandBox, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	landParams := strings.Split(c.PostForm(formValue), "-")
	if landParams[0] == "None" {
		return nil, sn.NewVError("You must select a land from which to redeploy an army.")
	}

	landIndex, err1 := strconv.Atoi(landParams[0])
	if err1 != nil {
		return nil, sn.NewVError("Invalid value received for %q tile.", formValue)
	}

	boxIndex, err1 := strconv.Atoi(landParams[1])
	if err1 != nil {
		return nil, sn.NewVError("Invalid value received for %q box.", formValue)
	}

	if landIndex < 0 || landIndex > 2 {
		return nil, sn.NewVError("Invalid value recieved for %q tile.", formValue)
	}

	land := g.ForeignLand(landIndex)
	if boxIndex < 0 || boxIndex >= len(g.ForeignLand(landIndex).Boxes) {
		return nil, sn.NewVError("Invalid value recieved for %q box: boxIndex: %d, Boxes length: %d.", formValue, boxIndex, len(g.ForeignLand(landIndex).Boxes))
	}

	return land.Box(boxIndex), nil
}

func (g *Game) getSpaceID(c *gin.Context) (id SpaceID, err error) {
	switch a := c.PostForm("action"); a {
	case "bribe-official", "secure-official":
		id = BribeSecureSpace
	case "nominate-student":
		id = NominateSpace
	case "force-exam":
		id = ForceSpace
	case "buy-junks", "start-voyage":
		id = JunksVoyageSpace
	case "recruit-army", "invade-land":
		id = RecruitArmySpace
	case "buy-gift":
		id = BuyGiftSpace
	case "give-gift":
		id = GiveGiftSpace
	case "commercial":
		id = CommercialSpace
	case "tax-income":
		id = TaxIncomeSpace
	case "no-action":
		id = NoActionSpace
	case "move-junks", "replace-student",
		"swap-officials", "redeploy-army",
		"replace-influence":
		id = PetitionSpace
	case "pass", "take-gift", "take-cash", "take-extra-action", "take-bribery-reward",
		"avenge-emperor", "take-army", "transfer-influence":
		id = NoSpace
	default:
		err = sn.NewVError("%q is an invalid action.", a)
	}
	return
}

func (g *Game) getMinistryAndSeniority(c *gin.Context, formValue string) (m *Ministry, s Seniority, err error) {
	var (
		i  int
		ss []string
	)

	if param := c.PostForm(formValue); param == "None" {
		err = sn.NewVError("You must select an official.")
		return
	} else if ss = strings.SplitN(param, "-", 2); len(ss) != 2 {
		err = sn.NewVError("Invalid format for ministry/seniority param.")
		return
	} else if i, err = strconv.Atoi(ss[1]); err != nil {
		err = sn.NewVError("Invalid Official Seniority Provided.")
		return
	}

	switch s = Seniority(i); ss[0] {
	case "Bingbu":
		m = g.Ministries[Bingbu]
	case "Hubu":
		m = g.Ministries[Hubu]
	case "Gongbu":
		m = g.Ministries[Gongbu]
	default:
		err = sn.NewVError("Invalid Ministry Provided.")
	}
	return
}

func (g *Game) getMinistryAndOfficial(c *gin.Context, formValue string) (m *Ministry, o *OfficialTile, err error) {
	var (
		s  Seniority
		ok bool
	)

	if m, s, err = g.getMinistryAndSeniority(c, formValue); err != nil {
		return
	}

	if o, ok = m.Officials[s]; !ok {
		err = sn.NewVError("Invalid official selected.")
	}
	return
}

func (g *Game) getPlayer(c *gin.Context, formValue string) (*Player, error) {
	sid := c.PostForm(formValue)

	switch sid {
	case "":
		return nil, sn.NewVError("Player form value %q not found.", formValue)
	case "none":
		return nil, sn.NewVError("You must select a player.")
	}

	p := g.PlayerBySID(sid)
	if p == nil {
		return nil, sn.NewVError("You must select a player.")
	}

	return p, nil
}

func (g *Game) getGiftValue(c *gin.Context, formValue string) (gv GiftCardValue, err error) {
	var gi int

	if gi, err = strconv.Atoi(c.PostForm(formValue)); err != nil {
		err = sn.NewVError("You must select an gift card.")
	} else {
		gv = GiftCardValue(gi)
	}
	return
}

func (g *Game) getRewardCard(c *gin.Context) (cd *EmperorCard, err error) {

	var t int
	if t, err = strconv.Atoi(c.PostForm("reward-card")); err != nil {
		err = sn.NewVError("You must select an Emperor's Reward card.")
		return
	}

	cp := g.CurrentPlayer()
	if cd = cp.GetEmperorCard(EmperorCardType(t)); cd == nil {
		err = sn.NewVError("You don't have the selected Emperor's Reward card.")
	}
	return
}

func (g *Game) getConCards(c *gin.Context, formValue string) (cds ConCards, err error) {
	var c1, c2, c3, c1Cnt, c2Cnt, c3Cnt int

	cp := g.CurrentPlayer()

	if c1, err = strconv.Atoi(c.PostForm(formValue + "-coins1")); err != nil || c1 < 0 {
		err = sn.NewVError("Invalid value for Coin 1 cards received.")
	} else if c2, err = strconv.Atoi(c.PostForm(formValue + "-coins2")); err != nil || c2 < 0 {
		err = sn.NewVError("Invalid value for Coin 2 cards received.")
	} else if c3, err = strconv.Atoi(c.PostForm(formValue + "-coins3")); err != nil || c3 < 0 {
		err = sn.NewVError("Invalid value for Coin 3 cards received.")
	} else if c1Cnt = cp.CardCount(1); c1 > c1Cnt {
		err = sn.NewVError("You selected %d cards with one coin, but only have %d of such cards.", c1, c1Cnt)
	} else if c2Cnt = cp.CardCount(2); c2 > c2Cnt {
		err = sn.NewVError("You selected %d cards with two coins, but only have %d of such cards.", c2, c2Cnt)
	} else if c3Cnt = cp.CardCount(3); c3 > c3Cnt {
		err = sn.NewVError("You selected %d cards with three coins, but only have %d of such cards.", c3, c3Cnt)
	} else {
		cds = ConCards{}.AppendN(1, c1).AppendN(2, c2).AppendN(3, c3)
	}
	return
}

func (g *Game) CoinOptions(prefix string) template.HTML {
	var s string
	for i := 1; i <= 3; i++ {
		s += fmt.Sprintf(`
                <div>
                        <label for="%s-coins%d">%d Coin Cards:</label>
		        <select id="%s-coins%d" name="%s-coins%d">`, prefix, i, i, prefix, i, prefix, i)
		for c := 0; c <= g.CurrentPlayer().ConCardHand.Count(i); c++ {
			s += fmt.Sprintf(`
                                <option value="%d">%d</option>`, c, c)
		}
		s += `
                        </select>
                </div>`
	}
	return template.HTML(s)
}

func (g *Game) RecruitArmyOptions() template.HTML {
	var result string
	cp := g.CurrentPlayer()
	licenses := cp.ConCardHand.Licenses()
	cost := cp.armyCost()

	if licenses >= cost {
		result += fmt.Sprintf("<option value='%d'>%d licenses</option>\n", cost, cost)
	}

	return template.HTML(result)
}

func (g *Game) BuyJunksOptions() template.HTML {
	var result string
	coins := g.CurrentPlayer().ConCardHand.Coins()
	for i := 1; i < 5; i++ {
		cost := g.CurrentPlayer().junkCostFor(i)
		if coins >= cost {
			coinText := "coins"
			if cost == 1 {
				coinText = "coin"
			}
			result += fmt.Sprintf("<option value='%d'>%d (%d %s)</option>\n", i, i, cost, coinText)
		}
	}
	return template.HTML(result)
}

func (g *Game) LicenseOptions(prefix string) template.HTML {
	var s string
	for licenses := 1; licenses <= 3; licenses++ {
		coins := 4 - licenses
		s += fmt.Sprintf(`
                <div>
                        <label for="%s-coins%d">%d License Cards:</label>
		        <select id="%s-coins%d" name="%s-coins%d">`,
			prefix, coins, licenses, prefix, coins, prefix, coins)
		for c := 0; c <= g.CurrentPlayer().ConCardHand.Count(coins); c++ {
			s += fmt.Sprintf(`
                                <option value="%d">%d</option>`, c, c)
		}
		s += `
                        </select>
                </div>`
	}
	return template.HTML(s)
}

type OfficialTest func(*OfficialTile) bool

func (g *Game) BriberyRewardOfficialOptions(id, label string, card *EmperorCard) template.HTML {
	s := fmt.Sprintf(`
        <div>
	        <label for=%q>%s</label>
	        <select id=%q name=%q>
	                <option value="None">None</option>`, id, label, id, id)
	ministries := g.emperorsRewardMinistriesFor(card)
	for _, ministryID := range g.MinistryIDS() {
		ministry := ministries[ministryID]
		if ministry == nil {
			continue
		}
		officials := make(OfficialTiles)
		for _, official := range ministry.Officials {
			if official.NotBribed() || official.Bribed() && !official.Secured && !official.Player().IsCurrentPlayer() {
				officials[official.Seniority] = official
			}
		}
		if len(officials) > 0 {
			cp := g.CurrentPlayer()
			s += fmt.Sprintf(`
                        <optgroup label=%q>`, ministry.Name())
			for _, seniority := range []Seniority{1, 2, 3, 4, 5, 6, 7} {
				if official := officials[seniority]; official != nil {
					var cost int
					if official.Bribed() {
						cost = cp.CostFor(official)
					}
					s += fmt.Sprintf(`
                                <option value="%s-%d">
                                        Official %d (%d Coins)
                                </option>`, ministry.Name(), seniority, seniority, cost)
				}
			}
			s += `
                        </optgroup>`
		}
	}
	s += `
                </select>
        </div>`
	return template.HTML(s)
}

func (g *Game) officialOptions(id, label string, test OfficialTest) template.HTML {
	s := fmt.Sprintf(`
        <div>
	        <label for=%q>%s</label>
                <select id=%q name=%q>
                        <option value="None">None</option>`, id, label, id, id)
	for _, ministryID := range g.MinistryIDS() {
		ministry := g.Ministries[ministryID]
		officials := make(OfficialTiles)
		for _, official := range ministry.Officials {
			if test(official) {
				officials[official.Seniority] = official
			}
		}
		if len(officials) > 0 {
			cp := g.CurrentPlayer()
			s += fmt.Sprintf(`
                        <optgroup label=%q>`, ministry.Name())
			for _, seniority := range []Seniority{1, 2, 3, 4, 5, 6, 7} {
				if official := officials[seniority]; official != nil {
					s += fmt.Sprintf(`
                                <option value="%s-%d">
                                        Official %d (%d Coins)
                                </option>`, ministry.Name(), seniority, seniority, cp.CostFor(official))
				}
			}
			s += `
                        </optgroup>`
		}
	}
	s += `
                </select>
        </div>`
	return template.HTML(s)
}

func (g *Game) BribedOfficialOptions(id, label string) template.HTML {
	return g.officialOptions(id, label, func(official *OfficialTile) bool {
		return !official.ministry.Resolved && official.NotBribed()

	})
}

func (g *Game) SecuredOfficialOptions(id, label string) template.HTML {
	return g.officialOptions(id, label, func(official *OfficialTile) bool {
		return !official.ministry.Resolved && official.Bribed() && official.Player().IsCurrentPlayer()

	})
}

func (g *Game) YourOfficialOptions(id, label string) template.HTML {
	return g.officialOptions(id, label, func(official *OfficialTile) bool {
		return !official.ministry.Resolved && official.Bribed() && official.Player().IsCurrentPlayer()

	})
}

func (g *Game) OtherOfficialOptions(id, label string) template.HTML {
	return g.officialOptions(id, label, func(official *OfficialTile) bool {
		return !official.ministry.Resolved && official.Bribed() && !official.Player().IsCurrentPlayer()
	})
}

func (g *Game) ReplaceInfluenceOfficialOptions(id, label string) template.HTML {
	return g.officialOptions(id, label, func(official *OfficialTile) bool {
		return !official.ministry.Resolved && official.Bribed() && !official.Secured

	})
}

func (g *Game) PlaceStudentOptions() template.HTML {
	var s string
	if len(g.Candidates) == 0 {
		return template.HTML(s)
	}
	s += `
        <div> 
                <label for="place-student-official">Official:</label> 
                <select id="place-student-official" name="official">
                        <option value="None">None</option>`
	for _, ministry := range g.MinistriesFor(g.Candidate()) {
		emptySpots := ministry.emptyCandidateSpots()
		if len(emptySpots) == 0 {
			emptySpots = ministry.unbribedUnsecuredCandidateSpots()
		}
		s += fmt.Sprintf(`
                        <optgroup label=%q>`, ministry.Name())
		for _, seniority := range emptySpots {
			s += fmt.Sprintf(`
                                <option value="%s-%d">Seniority Spot %d</option>`,
				ministry.Name(), seniority, seniority)
		}
		s += `
                        </optgroup>`
	}
	s += `
                </select> 
        </div>`
	return template.HTML(s)
}

type BoxTest func(*ForeignLandBox) bool

func (g *Game) boxOptions(id, label string, test BoxTest) template.HTML {
	s := fmt.Sprintf(`
        <div> 
                <label for=%q>%s</label> 
                <select id=%q name=%q>
                        <option value="None">None</option>`, id, label, id, id)
	for i, land := range g.ForeignLands {
		haveBoxes := false
		for _, box := range land.Boxes {
			if test(box) {
				haveBoxes = true
				break
			}
		}

		if haveBoxes {
			s += fmt.Sprintf(`
                        <optgroup label="%s (%d coins)">`, land.Name(), land.Cost())
			for j, box := range land.Boxes {
				if test(box) {
					s += fmt.Sprintf(`
                                <option value="%d-%d">`, i, j)
					if box.AwardCard {
						s += fmt.Sprintf(`
                                        VP: %d*`, box.Points)
					} else {
						s += fmt.Sprintf(`
                                        VP: %d`, box.Points)
					}
					s += `
                                </option>`
				}
			}
			s += `
                        </optgroug>`
		}
	}
	s += `
                </select>
        </div>`
	return template.HTML(s)
}

func (g *Game) FromLandOptions(id, label string) template.HTML {
	return g.boxOptions(id, label, func(box *ForeignLandBox) bool {
		return !box.land.Resolved && box.Invaded() && box.Player().IsCurrentPlayer()
	})
}

func (g *Game) ToLandOptions(id, label string) template.HTML {
	return g.boxOptions(id, label, func(box *ForeignLandBox) bool {
		return !box.land.Resolved && box.NotInvaded()
	})
}

func (g *Game) InvadeLandOptions(id, label string) template.HTML {
	return g.boxOptions(id, label, func(box *ForeignLandBox) bool {
		return g.CurrentPlayer().ConCardHand.Coins() >= box.land.Cost() && !box.land.Resolved && box.NotInvaded()
	})
}

func (g *Game) PetitionGiftOptions(cp *Player) template.HTML {
	s := `
        <label for="petition-gift">Gift:</label> 
        <select id="petition-gift" name="petition-gift">
                <option value="none">None</option>`
	for _, card := range cp.GiftsBought {
		if card.Value != Hanging {
			s += fmt.Sprintf(`
                <option value="%d">%s (%d Coins)</option>`, card.Value, card.Value, card.Value)
		}
	}
	s += `
        </select>`
	return template.HTML(s)
}

func (g *Game) MoveJunkPlayerOptions(prefix, label string) template.HTML {
	return g.playerOptions(prefix, label, func(player *Player) (result bool) {
		switch {
		case player.Junks < 1:
		default:
			result = true
		}
		return
	})
}

type PlayerTest func(*Player) bool

func (g *Game) OtherPlayerOptions(id, label string) template.HTML {
	return g.playerOptions(id, label, func(player *Player) bool {
		return !player.IsCurrentPlayer()
	})
}

func (g *Game) PlayerOptions(id, label string) template.HTML {
	return g.playerOptions(id, label, func(player *Player) bool {
		return true
	})
}

func (g *Game) ReplaceStudentOptions(id, label string) template.HTML {
	return g.playerOptions(id, label, func(p *Player) bool {
		c := g.Candidate()
		cp := g.CurrentPlayer()
		switch {
		case c == nil:
			return false
		case cp == nil:
			return false
		case c.Player().Equal(p) && c.Player().NotEqual(cp):
			return true
		case c.OtherPlayer().Equal(p) && c.OtherPlayer().NotEqual(cp):
			return true
		}
		return false
	})
}

func (g *Game) playerOptions(id, label string, test PlayerTest) template.HTML {
	s := fmt.Sprintf(`
        <div>
	        <label for=%q>%s</label>
	        <select id=%q name=%q>
                        <option value="none">None</option>`, id, label, id, id)
	for _, p := range g.Players() {
		if test == nil || test(p) {
			s += fmt.Sprintf(`
                        <option value="%d">%s (%s)</option>`, p.ID(), g.NameFor(p), p.Color())
		}
	}
	s += `
                </select>
        </div>`
	return template.HTML(s)
}

func (g *Game) GiftCardHandOptions(id, label string) template.HTML {
	return g.giftOptions(g.CurrentPlayer().GiftCardHand, id, label)
}

func (g *Game) GiveGiftCardOptions(id, label string) template.HTML {
	return g.giftOptions(g.CurrentPlayer().GiftsBought, id, label)
}

func (g *Game) giftOptions(cards GiftCards, id, label string) template.HTML {
	s := fmt.Sprintf(`
        <div>
                <label for=%q>%s</label>
                <select id=%q name=%q>`, id, label, id, id)
	for _, card := range cards {
		s += fmt.Sprintf(`
                        <option value="%d">%s (%d)</option>`, card.Value, card.Name(), card.Value)
	}
	s += `
                </select>
        </div>`
	return template.HTML(s)
}

func (g *Game) GiftTracker() template.HTML {
	s := `
        <div id="gift-tracker">
	        <div class="content">`
	for _, player := range g.Players() {
		for _, value := range g.GiftCardValues() {
			s += fmt.Sprintf(`
                        <div class="%s-%d">`, player.Color(), value)
			s += `
                                <div class="content">`
			for count, gift := range player.GiftsReceived.OfValue(value) {
				giver := gift.Player()
				s += fmt.Sprintf(`
                                        <div class="marker-%d">`, count)
				s += fmt.Sprintf(`
                                                <img src="/images/confucius/%s-barrel-shadowed.png" alt="%s Marker" />`,
					giver.Color(), giver.Color())
				s += `
                                        </div>`
			}
			s += `
                                </div>
                        </div>`
		}
	}
	s += `
                </div>
        </div>`
	return template.HTML(s)
}

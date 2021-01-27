package confucius

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (g *Game) invokeInvadePhase(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.invasionPhase(c)
	return "", game.Cache, nil
}

func (g *Game) adminHeader(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h := struct {
		Title         string           `form:"title"`
		Turn          int              `form:"turn" binding:"min=0"`
		Phase         game.Phase       `form:"phase" binding:"min=0"`
		SubPhase      game.SubPhase    `form:"sub-phase" binding:"min=0"`
		Round         int              `form:"round" binding:"min=0"`
		NumPlayers    int              `form:"num-players" binding"min=0,max=5"`
		Password      string           `form:"password"`
		CreatorID     int64            `form:"creator-id"`
		CreatorSID    string           `form:"creator-sid"`
		CreatorName   string           `form:"creator-name"`
		UserIDS       []int64          `form:"user-ids"`
		UserSIDS      []string         `form:"user-sids"`
		UserNames     []string         `form:"user-names"`
		UserEmails    []string         `form:"user-emails"`
		OrderIDS      game.UserIndices `form:"order-ids"`
		CPUserIndices game.UserIndices `form:"cp-user-indices"`
		WinnerIDS     game.UserIndices `form:"winner-ids"`
		Status        game.Status      `form:"status"`
		Progress      string           `form:"progress"`
		Options       []string         `form:"options"`
		OptString     string           `form:"opt-string"`
		CreatedAt     time.Time        `form:"created-at"`
		UpdatedAt     time.Time        `form:"updated-at"`
	}{}

	err := c.ShouldBind(&h)
	if err != nil {
		return "", game.None, err
	}

	s := struct {
		Junks           int  `form:"junks"`
		ChiefMinisterID int  `form:"chief-minister-id"`
		AdmiralID       int  `form:"admiral-id"`
		GeneralID       int  `form:"general-id"`
		AvengerID       int  `form:"avenger-id"`
		Wall            int  `form:"wall"`
		ExtraAction     bool `form:"extra-action"`
		BasicGame       bool `form:"basic-game"`
		AdmiralVariant  bool `form:"admiral-variant"`
	}{}
	err = c.ShouldBind(&s)
	if err != nil {
		return "", game.None, err
	}

	// h := game.NewHeader(c, nil, 0)
	// if err = restful.BindWith(c, h, binding.FormPost); err != nil {
	// 	act = game.None
	// 	return
	// }

	// s := new(State)
	// if err = restful.BindWith(c, s, binding.FormPost); err != nil {
	// 	act = game.None
	// 	return
	// }

	g.Title = h.Title
	g.Turn = h.Turn
	g.Phase = h.Phase
	g.SubPhase = h.SubPhase
	g.Round = h.Round
	g.NumPlayers = h.NumPlayers
	g.Password = h.Password
	g.CreatorID = h.CreatorID
	g.UserIDS = h.UserIDS
	g.OrderIDS = h.OrderIDS
	g.CPUserIndices = h.CPUserIndices
	g.WinnerIDS = h.WinnerIDS
	g.Status = h.Status

	g.Junks = s.Junks
	g.ChiefMinisterID = s.ChiefMinisterID
	g.AdmiralID = s.AdmiralID
	g.GeneralID = s.GeneralID
	g.Wall = s.Wall

	return "", game.Save, nil
}

//func (g *Game) adminState(ctx context.Context) (string, game.ActionType, error) {
//	log.Debugf(ctx, msgEnter)
//	defer log.Debugf(ctx, msgExit)
//
//	h := game.NewHeader(ctx, nil)
//	if err := restful.BindWith(ctx, h, binding.FormPost); err != nil {
//		return "", game.None, err
//	}
//
//	log.Debugf(ctx, "h: %#v", h)
//
//	g.UserIDS = h.UserIDS
//	g.Title = h.Title
//	g.Phase = h.Phase
//	g.Round = h.Round
//	g.NumPlayers = h.NumPlayers
//	g.Password = h.Password
//	g.CreatorID = h.CreatorID
//	if !(len(h.CPUserIndices) == 1 && h.CPUserIndices[0] == -1) {
//		g.CPUserIndices = h.CPUserIndices
//	}
//	if !(len(h.WinnerIDS) == 1 && h.WinnerIDS[0] == -1) {
//		g.WinnerIDS = h.WinnerIDS
//	}
//	g.Status = h.Status
//	return "", game.Save, nil
//}

//func adminState(g *Game, form url.Values) (string, game.ActionType, error) {
//	ctx := g.CTX()
//	values := ctx.Req.Form
//
//	for key, value := range values {
//		params := strings.Split(key, "-")
//		switch params[0] {
//		case "user":
//			index, err := strconv.Atoi(params[1])
//			if err != nil {
//				return "", game.None, err
//			}
//
//			uid, err := strconv.ParseInt(value[0], 10, 64)
//			if err != nil {
//				return "", game.None, err
//			}
//			u := user.New(ctx)
//			u.ID = uid
//			if err := datastore.Get(ctx, u); err != nil {
//				return "", game.None, err
//			}
//			g.Users[index] = u
//			g.UserIDS[index] = uid
//		case "title":
//			g.Title = value[0]
//		case "phase":
//			phaseInt, err := strconv.Atoi(value[0])
//			if err != nil {
//				return "", game.None, err
//			}
//			g.Phase = game.Phase(phaseInt)
//		case "round":
//			round, err := strconv.Atoi(value[0])
//			if err != nil {
//				return "", game.None, err
//			}
//			g.Round = round
//		case "junks":
//			junks, err := strconv.Atoi(value[0])
//			if err != nil {
//				return "", game.None, err
//			}
//			g.Junks = junks
//		case "wall":
//			wall, err := strconv.Atoi(value[0])
//			if err != nil {
//				return "", game.None, err
//			}
//			g.Wall = wall
//		case "numplayers":
//			numplayers, err := strconv.Atoi(value[0])
//			if err != nil {
//				return "", game.None, err
//			}
//			g.NumPlayers = numplayers
//		case "password":
//			g.Password = value[0]
//		case "creatorid":
//			intID, err := strconv.ParseInt(value[0], 10, 64)
//			if err != nil {
//				return "", game.None, err
//			}
//			g.CreatorID = intID
//		case "currentplayerid":
//			pids := game.UserIndices{}
//			if value[0] != "none" {
//				for _, v := range value {
//					pid, err := strconv.Atoi(v)
//					if err != nil {
//						return "", game.None, err
//					}
//					pids = append(pids, pid)
//				}
//			}
//			g.CPUserIndices = pids
//		case "chiefministerid":
//			if value[0] == "none" {
//				g.ChiefMinisterID = NoPlayerID
//			} else {
//				pid, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				g.ChiefMinisterID = pid
//			}
//		case "admiraliD":
//			if value[0] == "none" {
//				g.AdmiralID = NoPlayerID
//			} else {
//				pid, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				g.AdmiralID = pid
//			}
//		case "generalid":
//			if value[0] == "none" {
//				g.GeneralID = NoPlayerID
//			} else {
//				pid, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				g.GeneralID = pid
//			}
//		case "winnerid":
//			pids := game.UserIndices{}
//			if value[0] != "none" {
//				for _, v := range value {
//					pid, err := strconv.Atoi(v)
//					if err != nil {
//						return "", game.None, err
//					}
//					pids = append(pids, pid)
//				}
//			}
//			g.WinnerIDS = pids
//		case "status":
//			statusInt, err := strconv.Atoi(value[0])
//			if err != nil {
//				return "", game.None, err
//			}
//			g.Status = game.Status(statusInt)
//		case "student_player":
//			switch value[0] {
//			case "none":
//				g.Candidate().PlayerID = NoPlayerID
//			default:
//				pid, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				g.Candidate().PlayerID = pid
//			}
//		case "student_otherplayer":
//			switch value[0] {
//			case "none":
//				g.Candidate().OtherPlayerID = NoPlayerID
//			default:
//				pid, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				g.Candidate().OtherPlayerID = pid
//			}
//		case "removecandidate":
//			switch value[0] {
//			case "none":
//			default:
//				i, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				if len(g.Candidates) > 0 {
//					g.Candidates = append(g.Candidates[:i], g.Candidates[i+1:]...)
//				}
//			}
//		case "addemperorcard":
//			switch value[0] {
//			case "none":
//			default:
//				cardType, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				card := new(EmperorCard)
//				card.Type = EmperorCardType(cardType)
//				card.Revealed = true
//				g.EmperorDeck.Append(card)
//			}
//		case "removeemperorcard":
//			switch value[0] {
//			case "none":
//			default:
//				index, err := strconv.Atoi(value[0])
//				if err != nil {
//					return "", game.None, err
//				}
//				g.EmperorDeck = append(g.EmperorDeck[:index], g.EmperorDeck[index+1:]...)
//			}
//		}
//	}
//	return "", game.Save, nil
//}

func (g *Game) adminPlayer(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	pid, err := strconv.Atoi(c.PostForm("pid"))
	if err != nil {
		return "", game.None, err
	}

	if pid < 0 || pid >= g.NumPlayers {
		err = errors.New("Player not found.")
		return "", game.None, err
	}

	player := g.PlayerByID(pid)

	req := c.Request
	req.ParseForm()
	for key, value := range req.PostForm {
		switch key {
		case "id":
			var id int
			if id, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.SetID(id)
		case "score":
			var score int
			if score, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.Score = score
		case "actioncubes":
			var actioncubes int
			if actioncubes, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.ActionCubes = actioncubes
		case "junks":
			var junks int
			if junks, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.Junks = junks
		case "onvoyage":
			var onvoyage int
			if onvoyage, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.OnVoyage = onvoyage
		case "armies":
			var armies int
			if armies, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.Armies = armies
		case "recruitedarmies":
			var recruitedarmies int
			if recruitedarmies, err = strconv.Atoi(value[0]); err != nil {
				return "", game.None, err
			}
			player.RecruitedArmies = recruitedarmies
		case "performedaction":
			var performedaction bool
			if performedaction, err = strconv.ParseBool(value[0]); err != nil {
				return "", game.None, err
			}
			player.PerformedAction = performedaction
		case "takencommercial":
			var takencommercial bool
			if takencommercial, err = strconv.ParseBool(value[0]); err != nil {
				return "", game.None, err
			}
			player.TakenCommercial = takencommercial
		case "passed":
			var passed bool
			if passed, err = strconv.ParseBool(value[0]); err != nil {
				return "", game.None, err
			}
			player.Passed = passed
		case "addemperorcard":
			switch value[0] {
			case "none":
			default:
				var cardType int
				if cardType, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				card := &EmperorCard{Type: EmperorCardType(cardType), Revealed: true}
				player.EmperorHand = append(player.EmperorHand, card)
			}
		case "removeemperorcard":
			switch value[0] {
			case "none":
			default:
				var index int
				if index, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				player.EmperorHand = append(player.EmperorHand[:index], player.EmperorHand[index+1:]...)
			}
		case "addavailablegift":
			switch value[0] {
			case "none":
			default:
				var cardValue int
				if cardValue, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				card := &GiftCard{Value: GiftCardValue(cardValue + 1), PlayerID: player.ID()}
				player.GiftCardHand.Append(card)
			}
		case "removeavailablegift":
			switch value[0] {
			case "none":
			default:
				var i int
				if i, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				player.GiftCardHand = append(player.GiftCardHand[:i], player.GiftCardHand[i+1:]...)
			}
		case "addreceivedgift":
			if params := strings.Split(value[0], "-"); len(params) == 2 {
				switch params[0] {
				case "none":
				default:
					var cardValue int
					if cardValue, err = strconv.Atoi(params[0]); err != nil {
						return "", game.None, err
					}
					var pid int
					if pid, err = strconv.Atoi(params[1]); err != nil {
						return "", game.None, err
					}
					card := &GiftCard{Value: GiftCardValue(cardValue + 1), PlayerID: pid}
					player.GiftsReceived.Append(card)
				}
			}
		case "removereceivedgift":
			switch value[0] {
			case "none":
			default:
				var i int
				if i, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				player.GiftsReceived = append(player.GiftsReceived[:i], player.GiftsReceived[i+1:]...)
			}
		case "addboughtgift":
			switch value[0] {
			case "none":
			default:
				var cardValue int
				if cardValue, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				card := &GiftCard{Value: GiftCardValue(cardValue), PlayerID: player.ID()}
				player.GiftsBought.Append(card)
			}
		case "removeboughtgift":
			switch value[0] {
			case "none":
			default:
				var i int
				if i, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				player.GiftsBought = append(player.GiftsBought[:i], player.GiftsBought[i+1:]...)
			}
		case "addconcard":
			switch value[0] {
			case "none":
			default:
				var coins int
				if coins, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				player.ConCardHand.Append(&ConCard{Coins: coins, Revealed: true})
			}
		case "removeconcard":
			switch value[0] {
			case "none":
			default:
				var i int
				if i, err = strconv.Atoi(value[0]); err != nil {
					return "", game.None, err
				}
				player.ConCardHand = append(player.ConCardHand[:i], player.ConCardHand[i+1:]...)
			}
		}
	}
	return "", game.Save, nil
}

//func adminMinstry(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	mid, err := strconv.Atoi(values.Get("mid"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	ministry := g.Ministries[MinistryID(mid)]
//
//	for key, value := range values {
//		params := strings.Split(key, "-")
//		switch params[0] {
//		case "minister":
//			switch params[1] {
//			case "value":
//				var v int
//				if v, err = strconv.Atoi(value[0]); err != nil {
//					return "", game.None, err
//				}
//				ministry.MinisterChit = MinistryChit(v)
//			case "playerid":
//				switch playersid := value[0]; playersid {
//				case "none":
//					ministry.MinisterID = NoPlayerID
//				default:
//					var playerid int
//					if playerid, err = strconv.Atoi(playersid); err != nil {
//						return "", game.None, err
//					}
//					ministry.MinisterID = playerid
//				}
//			}
//		case "secretary":
//			switch params[1] {
//			case "value":
//				var v int
//				if v, err = strconv.Atoi(value[0]); err != nil {
//					return "", game.None, err
//				}
//				ministry.SecretaryChit = MinistryChit(v)
//			case "playerid":
//				switch playersid := value[0]; playersid {
//				case "none":
//					ministry.SecretaryID = NoPlayerID
//				default:
//					var playerid int
//					if playerid, err = strconv.Atoi(playersid); err != nil {
//						return "", game.None, err
//					}
//					ministry.SecretaryID = playerid
//				}
//			}
//		case "resolved":
//			var resolved bool
//			if resolved, err = strconv.ParseBool(value[0]); err != nil {
//				return "", game.None, err
//			}
//			ministry.Resolved = resolved
//		case "inprogress":
//			var inprogress bool
//			if inprogress, err = strconv.ParseBool(value[0]); err != nil {
//				return "", game.None, err
//			}
//			ministry.InProgress = inprogress
//		}
//	}
//	return "", game.Save, nil
//}
//
//func adminMinstryOfficial(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	mid, err := strconv.Atoi(values.Get("mid"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	ministry := g.Ministries[MinistryID(mid)]
//
//	oldSeniorityI, err := strconv.Atoi(values.Get("seniority"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	oldSeniority := Seniority(oldSeniorityI)
//	official := ministry.Officials[oldSeniority]
//
//	for key, value := range values {
//		switch key {
//		case "cost":
//			var cost int
//			if cost, err = strconv.Atoi(value[0]); err != nil {
//				return "", game.None, err
//			}
//			official.Cost = cost
//		case "variant":
//			var variant int
//			if variant, err = strconv.Atoi(value[0]); err != nil {
//				return "", game.None, err
//			}
//			official.Variant = VariantID(variant)
//		case "playerid":
//			switch playersid := value[0]; playersid {
//			case "none":
//				official.PlayerID = NoPlayerID
//			default:
//				var playerid int
//				if playerid, err = strconv.Atoi(playersid); err != nil {
//					return "", game.None, err
//				}
//				official.PlayerID = playerid
//			}
//		case "tempid":
//			switch tempsid := value[0]; tempsid {
//			case "none":
//				official.TempID = NoPlayerID
//			default:
//				var tempid int
//				if tempid, err = strconv.Atoi(tempsid); err != nil {
//					return "", game.None, err
//				}
//				official.TempID = tempid
//			}
//		case "secured":
//			var secured bool
//			if secured, err = strconv.ParseBool(value[0]); err != nil {
//				return "", game.None, err
//			}
//			official.Secured = secured
//		case "new-seniority":
//			var newSeniorityI int
//			if newSeniorityI, err = strconv.Atoi(value[0]); err != nil {
//				return "", game.None, err
//			}
//			newSeniority := Seniority(newSeniorityI)
//			if newSeniorityI != oldSeniorityI {
//				official.Seniority = newSeniority
//				ministry.Officials[newSeniority] = official
//				delete(ministry.Officials, oldSeniority)
//			}
//		}
//	}
//	return "", game.Save, nil
//}
//
//func adminCandidate(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	for key, value := range values {
//		params := strings.Split(key, "-")
//		switch params[0] {
//		case "student_player":
//			switch value[0] {
//			case "none":
//				g.Candidate().PlayerID = NoPlayerID
//			default:
//				var pid int
//				if pid, err = strconv.Atoi(value[0]); err != nil {
//					return "", game.None, err
//				}
//				g.Candidate().PlayerID = pid
//			}
//		case "student_otherplayer":
//			switch value[0] {
//			case "none":
//				g.Candidate().OtherPlayerID = NoPlayerID
//			default:
//				var pid int
//				if pid, err = strconv.Atoi(value[0]); err != nil {
//					return "", game.None, err
//				}
//				g.Candidate().OtherPlayerID = pid
//			}
//		case "removecandidate":
//			switch value[0] {
//			case "none":
//			default:
//				var i int
//				if i, err = strconv.Atoi(value[0]); err != nil {
//					return "", game.None, err
//				}
//				if len(g.Candidates) > 0 {
//					g.Candidates = append(g.Candidates[:i], g.Candidates[i+1:]...)
//				}
//			}
//		}
//	}
//	return "", game.Save, nil
//}
//
//func adminForeignLand(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	index, err := strconv.Atoi(values.Get("lid"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	land := g.ForeignLands[index]
//
//	for key, value := range values {
//		switch key {
//		case "resolved":
//			var resolved bool
//			if resolved, err = strconv.ParseBool(value[0]); err != nil {
//				return "", game.None, err
//			}
//			land.Resolved = resolved
//		}
//	}
//	return "", game.Save, nil
//}
//
//func adminForeignLandBox(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	lindex, err := strconv.Atoi(values.Get("lid"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	land := g.ForeignLands[lindex]
//
//	bindex, err := strconv.Atoi(values.Get("bid"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	box := land.Boxes[bindex]
//
//	for key, value := range values {
//		switch key {
//		case "position":
//			var position int
//			if position, err = strconv.Atoi(value[0]); err != nil {
//				return "", game.None, err
//			}
//			box.Position = position
//		case "playerid":
//			switch playersid := value[0]; playersid {
//			case "none":
//				box.PlayerID = NoPlayerID
//			default:
//				var playerid int
//				if playerid, err = strconv.Atoi(playersid); err != nil {
//					return "", game.None, err
//				}
//				box.PlayerID = playerid
//			}
//		case "points":
//			var points int
//			if points, err = strconv.Atoi(value[0]); err != nil {
//				return "", game.None, err
//			}
//			box.Points = points
//		case "card":
//			var card bool
//			if card, err = strconv.ParseBool(value[0]); err != nil {
//				return "", game.None, err
//			}
//			box.AwardCard = card
//		}
//	}
//	return "", game.Save, nil
//}
//
//func adminActionSpace(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	index, err := strconv.Atoi(values.Get("sid"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	space := g.ActionSpaces[SpaceID(index)]
//
//	for _, player := range g.Players() {
//		key := fmt.Sprintf("player-%d-cubes", player.ID())
//		value := values.Get(key)
//
//		var cubes int
//		if cubes, err = strconv.Atoi(value); err != nil {
//			return "", game.None, err
//		}
//		space.Cubes[player.ID()] = cubes
//	}
//	return "", game.Save, nil
//}
//
//func adminDistantLand(g *Game, form url.Values) (string, game.ActionType, error) {
//	values, err := g.getValues()
//	if err != nil {
//		return "", game.None, err
//	}
//
//	index, err := strconv.Atoi(values.Get("lindex"))
//	if err != nil {
//		return "", game.None, err
//	}
//
//	land := g.DistantLands[index]
//
//	for key, value := range values {
//		switch key {
//		case "chit":
//			if value[0] == "none" {
//				land.Chit = NoChit
//			} else {
//				var chit int
//				if chit, err = strconv.Atoi(value[0]); err != nil {
//					return "", game.None, err
//				}
//				land.Chit = DistantLandChit(chit)
//			}
//		case "playerids":
//			pids := game.UserIndices{}
//			if value[0] != "none" {
//				for _, v := range value {
//					pid, err := strconv.Atoi(v)
//					if err != nil {
//						return "", game.None, err
//					}
//					pids = append(pids, pid)
//				}
//			}
//			land.PlayerIDS = pids
//		}
//	}
//	return "", game.Save, nil
//}

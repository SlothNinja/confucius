package confucius

import (
	"encoding/gob"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(make(restful.Notices, 0))
	gob.Register(make(restful.Errors, 0))
}

const (
	gameKey   = "Game"
	homePath  = "/"
	jsonKey   = "JSON"
	statusKey = "Status"
	msgEnter  = "Entering"
	msgExit   = "Exiting"
)

var (
	ErrInvalidID = errors.New("invalid identifier")
)

func gameFrom(c *gin.Context) *Game {
	g, _ := c.Value(gameKey).(*Game)
	return g
}

func withGame(c *gin.Context, g *Game) *gin.Context {
	c.Set(gameKey, g)
	return c
}

func jsonFrom(c *gin.Context) *Game {
	g, _ := c.Value(jsonKey).(*Game)
	return g
}

func withJSON(c *gin.Context, g *Game) *gin.Context {
	c.Set(jsonKey, g)
	return c
}

func (g *Game) Update(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	switch a := c.PostForm("action"); a {
	case "bribe-official":
		return g.bribeOfficial(c, cu)
	case "secure-official":
		return g.secureOfficial(c, cu)
	case "buy-gift":
		return g.buyGift(c, cu)
	case "give-gift":
		return g.giveGift(c, cu)
	case "nominate-student":
		return g.nominateStudent(c, cu)
	case "force-exam":
		return g.forceExam(c, cu)
	case "transfer-influence":
		return g.transferInfluence(c, cu)
	case "temp-transfer-influence":
		return g.tempTransfer(c, cu)
	case "move-junks":
		return g.moveJunks(c, cu)
	case "replace-student":
		return g.replaceStudent(c, cu)
	case "swap-officials":
		return g.swapOfficials(c, cu)
	case "redeploy-army":
		return g.redeployArmy(c, cu)
	case "replace-influence":
		return g.replaceInfluence(c, cu)
	case "place-student":
		return g.placeStudent(c, cu)
	case "buy-junks":
		return g.buyJunks(c, cu)
	case "start-voyage":
		return g.startVoyage(c, cu)
	case "commercial":
		return g.commercial(c, cu)
	case "tax-income":
		return g.taxIncome(c, cu)
	case "recruit-army":
		return g.recruitArmy(c, cu)
	case "invade-land":
		return g.invadeLand(c, cu)
	case "no-action":
		return g.noAction(c, cu)
	case "pass":
		return g.pass(c, cu)
	case "take-cash":
		return g.takeCash(c, cu)
	case "take-gift":
		return g.takeGift(c, cu)
	case "take-extra-action":
		return g.takeExtraAction(c, cu)
	case "take-bribery-reward":
		return g.takeBriberyReward(c, cu)
	case "avenge-emperor":
		return g.avengeEmperor(c, cu)
	case "take-army":
		return g.takeArmy(c, cu)
	case "discard":
		return g.discard(c, cu)
	case "choose-chief-minister":
		return g.chooseChiefMinister(c, cu)
	case "tutor-student":
		return g.tutorStudent(c, cu)
	case "reset":
		return g.resetTurn(c, cu)
	case "game-state":
		return g.adminHeader(c, cu)
	case "player":
		return g.adminPlayer(c, cu)
		//	case "ministry":
		//		return g.adminMinstry(c)
		//	case "official":
		//		return g.adminMinstryOfficial(c)
		//	case "candidate":
		//		return g.adminCandidate(c)
		//	case "foreign-land":
		//		return g.adminForeignLand(c)
		//	case "foreign-land-box":
		//		return g.adminForeignLandBox(c)
		//	case "action-space":
		//		return g.adminActionSpace(c)
		//	case "invoke-invade-phase":
		//		return g.invokeInvadePhase(c)
		//	case "distant-land":
		//		return g.adminDistantLand(c)
	default:
		return "confucius/flash_notice", game.None, sn.NewVError("%v is not a valid action.", a)
	}
}

// gets any notices and errors from flash and clears flash
func getFlashes(c *gin.Context) (restful.Notices, restful.Errors, error) {
	session := sessions.Default(c)

	var ns, es []template.HTML
	notices := session.Flashes("_notices")
	session.AddFlash("", "_notices")
	for i := range notices {
		n, ok := notices[i].(restful.Notices)
		if ok {
			ns = append(es, n...)
		}
	}

	errors := session.Flashes("_errors")
	session.AddFlash("", "_errors")
	for i := range errors {
		e, ok := errors[i].(restful.Errors)
		if ok {
			es = append(es, e...)
		}
	}

	err := session.Save()
	return ns, es, err
}

func (client *Client) show(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		ml, err := client.MLog.Get(c, id)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}

		notices, errors, err := getFlashes(c)
		if err != nil {
			client.Log.Errorf(err.Error())
		}

		c.HTML(http.StatusOK, prefix+"/show", gin.H{
			"Context":    c,
			"VersionID":  sn.VersionID(),
			"CUser":      cu,
			"Game":       gameFrom(c),
			"IsAdmin":    cu.IsAdmin(),
			"Admin":      game.AdminFrom(c),
			"MessageLog": ml,
			"ColorMap":   color.MapFrom(c),
			"Notices":    notices,
			"Errors":     errors,
		})
	}
}

func (client *Client) addMessage(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
			return
		}

		ml, err := client.MLog.Get(c, id)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		m := ml.AddMessage(cu, c.PostForm("message"))

		_, err = client.MLog.Put(c, id, ml)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		c.HTML(http.StatusOK, "shared/message", gin.H{
			"message": m,
			"ctx":     c,
			"map":     gameFrom(c).ColorMapFor(cu),
			"link":    cu.Link(),
		})
	}
}

func (client *Client) update(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := gameFrom(c)
		if g == nil {
			client.Log.Errorf("Controller#Update Game Not Found")
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}

		template, actionType, err := g.Update(c, cu)
		client.Log.Debugf("template: %v actionType: %v err: %v", template, actionType, err)
		switch {
		case err != nil && sn.IsVError(err):
			restful.AddErrorf(c, "%v", err)
			withJSON(c, g)
		case err != nil:
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		case actionType == game.Cache:
			client.Cache.SetDefault(g.UndoKey(cu), g)
		case actionType == game.Save:
			err := client.save(c, g, cu)
			if err != nil {
				client.Log.Errorf("%s", err)
				restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(c, prefix))
				return
			}
		case actionType == game.Undo:
			mkey := g.UndoKey(cu)
			client.Cache.Delete(mkey)
		}

		switch jData := jsonFrom(c); {
		case jData != nil && template == "json":
			c.JSON(http.StatusOK, jData)
		case template == "":
			notices := restful.NoticesFrom(c)
			errors := restful.ErrorsFrom(c)

			client.Log.Debugf("template: %s", template)
			client.Log.Debugf("Notices: %v", notices)
			client.Log.Debugf("Errors: %v", errors)

			if len(notices) < 1 && len(errors) < 1 {
				c.Redirect(http.StatusSeeOther, showPath(c, prefix))
				return
			}

			session := sessions.Default(c)
			if len(notices) > 0 {
				session.AddFlash(notices, "_notices")
			}
			if len(errors) > 0 {
				session.AddFlash(errors, "_errors")
			}
			err := session.Save()
			if err != nil {
				client.Log.Errorf(err.Error())
			}
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
		default:
			client.Log.Debugf("template: %s", template)
			client.Log.Debugf("Notices: %v", restful.NoticesFrom(c))
			client.Log.Debugf("Errors: %v", restful.ErrorsFrom(c))
			cu, err := client.User.Current(c)
			if err != nil {
				client.Log.Debugf(err.Error())
			}

			d := gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Game":      g,
				"IsAdmin":   cu.IsAdmin(),
				"Notices":   restful.NoticesFrom(c),
				"Errors":    restful.ErrorsFrom(c),
			}
			c.HTML(http.StatusOK, template, d)
		}
	}
}

func (client *Client) save(c *gin.Context, g *Game, cu *user.User) error {
	_, err := client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		oldG := New(c, g.ID())
		err := tx.Get(oldG.Key, oldG.Header)
		if err != nil {
			return err
		}

		if oldG.UpdatedAt != g.UpdatedAt {
			return fmt.Errorf("Game state changed unexpectantly.  Try again.")
		}

		err = g.encode(c)
		if err != nil {
			return err
		}

		_, err = tx.Put(g.Key, g.Header)
		if err != nil {
			return err
		}

		client.Cache.Delete(g.UndoKey(cu))
		return nil
	})
	return err
}

func (client *Client) saveWith(c *gin.Context, g *Game, cu *user.User, ks []*datastore.Key, es []interface{}) error {
	_, err := client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		oldG := New(c, g.ID())
		err := tx.Get(oldG.Key, oldG.Header)
		if err != nil {
			return err
		}

		if oldG.UpdatedAt != g.UpdatedAt {
			return fmt.Errorf("Game state changed unexpectantly.  Try again.")
		}

		err = g.encode(c)
		if err != nil {
			return err
		}

		ks = append(ks, g.Key)
		es = append(es, g.Header)

		_, err = tx.PutMulti(ks, es)
		if err != nil {
			return err
		}

		client.Cache.Delete(g.UndoKey(cu))
		return nil
	})
	return err
}

// Playerers game.Playerers
// Log       game.GameLog
// Junks     int `form:"junks"`

// ChiefMinisterID int `form:"chief-minister-id"`
// AdmiralID       int `form:"admiral-id"`
// GeneralID       int `form:"general-id"`
// AvengerID       int `form:"avenger-id"`

// ActionSpaces ActionSpaces

// Candidates     CandidateTiles
// OfficialsDeck  OfficialsDeck
// ConDeck        ConCards
// ConDiscardPile ConCards
// EmperorDeck    EmperorCards
// EmperorDiscard EmperorCards

// DistantLands DistantLands
// ForeignLands ForeignLands

// Ministries Ministries

// Wall        int  `form:"wall"`
// ExtraAction bool `form:"extra-action"`

// BasicGame      bool `form:"basic-game"`
// AdmiralVariant bool `form:"admiral-variant"`

func (g *Game) encode(cx *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	log.Debugf("state: %#v", g.State)
	encoded, err := codec.Encode(g.State)
	if err != nil {
		return err
	}
	g.SavedState = encoded
	g.updateHeader()

	return nil
}

// func (g *Game) cache(c *gin.Context) error {
// 	item := &memcache.Item{
// 		Key:        g.UndoKey(c),
// 		Expiration: time.Minute * 30,
// 	}
// 	v, err := codec.Encode(g)
// 	if err != nil {
// 		return err
// 	}
// 	item.Value = v
// 	return memcache.Set(c, item)
// }

func wrap(s *user.Stats, cs []*contest.Contest) ([]*datastore.Key, []interface{}) {
	l := len(cs) + 1
	es := make([]interface{}, l)
	ks := make([]*datastore.Key, l)
	es[0] = s
	ks[0] = s.Key
	for i, c := range cs {
		es[i+1] = c
		ks[i+1] = c.Key
	}
	return ks, es
}

func showPath(c *gin.Context, prefix string) string {
	return fmt.Sprintf("/%s/game/show/%s", prefix, c.Param("hid"))
}

func recruitingPath(prefix string) string {
	return fmt.Sprintf("/%s/games/recruiting", prefix)
}

func newPath(prefix string) string {
	return fmt.Sprintf("/%s/game/new", prefix)
}

func newGamer(c *gin.Context) game.Gamer {
	return New(c, 0)
}

func (client *Client) undo(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)
		c.Redirect(http.StatusSeeOther, showPath(c, prefix))

		g := gameFrom(c)
		if g == nil {
			client.Log.Errorf("Controller#Update Game Not Found")
			return
		}
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}
		mkey := g.UndoKey(cu)
		client.Cache.Delete(mkey)
	}
}

func (client *Client) endRound(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := gameFrom(c)
		if g == nil {
			client.Log.Errorf("game not found")
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		_, err := client.endOfRoundPhase(c, g)
		if err != nil {
			client.Log.Errorf("cache error: %s", err.Error())
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
			return
		}

		err = client.save(c, g, cu)
		if err != nil {
			client.Log.Errorf("cache error: %s", err.Error())
		}
		c.Redirect(http.StatusSeeOther, showPath(c, prefix))
		return
	}
}

func (client *Client) index(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		gs := game.GamersFrom(c)
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}
		switch status := game.StatusFrom(c); status {
		case game.Recruiting:
			c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Games":     gs,
				"Type":      gtype.Confucius.String(),
			})
		default:
			c.HTML(http.StatusOK, "shared/games_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Games":     gs,
				"Type":      gtype.Confucius.String(),
				"Status":    status,
			})
		}
	}
}
func (client *Client) newAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := New(c, 0)
		withGame(c, g)
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}

		if err := g.FromParams(c, cu, gtype.GOT); err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		c.HTML(http.StatusOK, prefix+"/new", gin.H{
			"Context":   c,
			"VersionID": sn.VersionID(),
			"CUser":     cu,
			"Game":      g,
		})
	}
}

func (client *Client) create(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := New(c, 0)
		withGame(c, g)

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.FromParams(c, cu, g.Type)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.fromForm(c, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.encode(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		ks, err := client.DS.AllocateIDs(c, []*datastore.Key{g.Key})
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		k := ks[0]

		_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
			m := mlog.New(k.ID)
			ks = []*datastore.Key{m.Key, k}
			es := []interface{}{m, g.Header}

			_, err = tx.PutMulti(ks, es)
			return err
		})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		restful.AddNoticef(c, "<div>%s created.</div>", g.Title)
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
}

func (client *Client) accept(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := gameFrom(c)
		if g == nil {
			client.Log.Errorf("game not found")
			restful.AddErrorf(c, "game not found")
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}
		start, err := g.Accept(c, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		if start {
			err = g.Start(c)
			if err != nil {
				client.Log.Errorf(err.Error())
				restful.AddErrorf(c, err.Error())
				c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
				return
			}
		}

		err = client.save(c, g, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		if start {
			err = g.SendTurnNotificationsTo(c, g.CurrentPlayer())
			if err != nil {
				client.Log.Warningf(err.Error())
			}
		}
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
}

func (client *Client) drop(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := gameFrom(c)
		if g == nil {
			client.Log.Errorf("game not found")
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}

		err = g.Drop(cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = client.save(c, g, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
		}
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
}

func (client *Client) fetch(c *gin.Context) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	// create Gamer
	id, err := strconv.ParseInt(c.Param("hid"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	g := New(c, id)

	switch action := c.PostForm("action"); {
	case action == "reset":
		// pull from cache/datastore
		// same as undo & !MultiUndo
		fallthrough
	case action == "undo":
		// pull from cache/datastore
		err = client.dsGet(c, g)
		if err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	default:
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}
		if cu != nil {
			// pull from cache and return if successful; otherwise pull from datastore
			err = client.mcGet(c, g, cu)
			if err == nil {
				return
			}
		}
		err = client.dsGet(c, g)
		if err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	}
}

// pull temporary game state from cache.  Note may be different from value stored in datastore.
func (client *Client) mcGet(c *gin.Context, g *Game, cu *user.User) error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	mkey := g.GetHeader().UndoKey(cu)
	item, found := client.Cache.Get(mkey)
	if !found {
		return fmt.Errorf("game not found")
	}

	g2, ok := item.(*Game)
	if !ok {
		return fmt.Errorf("item not a *Game")
	}
	g2.SetCTX(c)

	cu, err := client.User.Current(c)
	if err != nil {
		client.Log.Debugf(err.Error())
	}
	g = g2
	color.WithMap(withGame(c, g), g.ColorMapFor(cu))
	return nil
}

// pull game state from cache/datastore.  returned memcache should be same as datastore.
func (client *Client) dsGet(c *gin.Context, g *Game) error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.DS.Get(c, g.Key, g.Header)
	switch {
	case err != nil:
		restful.AddErrorf(c, err.Error())
		return err
	case g == nil:
		err = fmt.Errorf("Unable to get game for id: %v", g.ID)
		restful.AddErrorf(c, err.Error())
		return err
	}

	s := newState()
	err = codec.Decode(&s, g.SavedState)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}
	g.State = s

	err = client.init(c, g)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}
	cu, err := client.User.Current(c)
	if err != nil {
		client.Log.Debugf(err.Error())
	}

	cm := g.ColorMapFor(cu)
	color.WithMap(withGame(c, g), cm)
	return nil
}

func JSON(c *gin.Context) {
	c.JSON(http.StatusOK, gameFrom(c))
}

func (client *Client) jsonIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Game.JSONIndexAction(c)
	}
}

func (g *Game) updateHeader() {
	g.OptString = g.options()
	switch g.Phase {
	case GameOver:
		g.Progress = g.PhaseName()
	default:
		g.Progress = fmt.Sprintf("<div>Round: %d</div><div>Phase: %s</div>", g.Round, g.PhaseName())
	}
	// if u := g.Creator; u != nil {
	// 	g.CreatorSID = user.GenID(u.GoogleID)
	// 	g.CreatorName = u.Name
	// }

	// if l := len(g.Users); l > 0 {
	// 	g.UserSIDS = make([]string, l)
	// 	g.UserNames = make([]string, l)
	// 	g.UserEmails = make([]string, l)
	// 	for i, u := range g.Users {
	// 		g.UserSIDS[i] = user.GenID(u.GoogleID)
	// 		g.UserNames[i] = u.Name
	// 		g.UserEmails[i] = u.Email
	// 	}
	// }
}
func getID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("hid"), 10, 64)
	if err != nil {
		return -1, ErrInvalidID
	}
	return id, nil
}

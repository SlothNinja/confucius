package confucius

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/memcache"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

//type Action func(*Game, url.Values) (string, game.ActionType, error)
//
//var actionMap = map[string]Action{
//	"bribe-official":          bribeOfficial,
//	"secure-official":         secureOfficial,
//	"buy-gift":                buyGift,
//	"give-gift":               giveGift,
//	"nominate-student":        nominateStudent,
//	"force-exam":              forceExam,
//	"transfer-influence":      transferInfluence,
//	"temp-transfer-influence": tempTransfer,
//	"move-junks":              moveJunks,
//	"replace-student":         replaceStudent,
//	"swap-officials":          swapOfficials,
//	"redeploy-army":           redeployArmy,
//	"replace-influence":       replaceInfluence,
//	"place-student":           placeStudent,
//	"buy-junks":               buyJunks,
//	"start-voyage":            startVoyage,
//	"commercial":              commercial,
//	"tax-income":              taxIncome,
//	"recruit-army":            recruitArmy,
//	"invade-land":             invadeLand,
//	"no-action":               noAction,
//	"pass":                    pass,
//	"take-cash":               takeCash,
//	"take-gift":               takeGift,
//	"take-extra-action":       takeExtraAction,
//	"take-bribery-reward":     takeBriberyReward,
//	"avenge-emperor":          avengeEmperor,
//	"take-army":               takeArmy,
//	"discard":                 discard,
//	"choose-chief-minister":   chooseChiefMinister,
//	"tutor-student":           tutorStudent,
//	"reset":                   resetTurn,
//	"finish":                  finishTurn,
//	"game-state":              adminState,
//	"player":                  adminPlayer,
//	"ministry":                adminMinstry,
//	"official":                adminMinstryOfficial,
//	"candidate":               adminCandidate,
//	"foreign-land":            adminForeignLand,
//	"foreign-land-box":        adminForeignLandBox,
//	"action-space":            adminActionSpace,
//	"invoke-invade-phase":     invokeInvadePhase,
//	"distant-land":            adminDistantLand,
//}

const (
	gameKey   = "Game"
	homePath  = "/"
	jsonKey   = "JSON"
	statusKey = "Status"
)

func gameFrom(c *gin.Context) (g *Game) {
	g, _ = c.Value(gameKey).(*Game)
	return
}

func withGame(c *gin.Context, g *Game) *gin.Context {
	c.Set(gameKey, g)
	return c
}

func jsonFrom(c *gin.Context) (g *Game) {
	g, _ = c.Value(jsonKey).(*Game)
	return
}

func withJSON(c *gin.Context, g *Game) *gin.Context {
	c.Set(jsonKey, g)
	return c
}

func (g *Game) Update(c *gin.Context) (tmpl string, t game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch a := c.PostForm("action"); a {
	case "bribe-official":
		return g.bribeOfficial(c)
	case "secure-official":
		return g.secureOfficial(c)
	case "buy-gift":
		return g.buyGift(c)
	case "give-gift":
		return g.giveGift(c)
	case "nominate-student":
		return g.nominateStudent(c)
	case "force-exam":
		return g.forceExam(c)
	case "transfer-influence":
		return g.transferInfluence(c)
	case "temp-transfer-influence":
		return g.tempTransfer(c)
	case "move-junks":
		return g.moveJunks(c)
	case "replace-student":
		return g.replaceStudent(c)
	case "swap-officials":
		return g.swapOfficials(c)
	case "redeploy-army":
		return g.redeployArmy(c)
	case "replace-influence":
		return g.replaceInfluence(c)
	case "place-student":
		return g.placeStudent(c)
	case "buy-junks":
		return g.buyJunks(c)
	case "start-voyage":
		return g.startVoyage(c)
	case "commercial":
		return g.commercial(c)
	case "tax-income":
		return g.taxIncome(c)
	case "recruit-army":
		return g.recruitArmy(c)
	case "invade-land":
		return g.invadeLand(c)
	case "no-action":
		return g.noAction(c)
	case "pass":
		return g.pass(c)
	case "take-cash":
		return g.takeCash(c)
	case "take-gift":
		return g.takeGift(c)
	case "take-extra-action":
		return g.takeExtraAction(c)
	case "take-bribery-reward":
		return g.takeBriberyReward(c)
	case "avenge-emperor":
		return g.avengeEmperor(c)
	case "take-army":
		return g.takeArmy(c)
	case "discard":
		return g.discard(c)
	case "choose-chief-minister":
		return g.chooseChiefMinister(c)
	case "tutor-student":
		return g.tutorStudent(c)
	case "reset":
		return g.resetTurn(c)
	case "game-state":
		return g.adminHeader(c)
	case "player":
		return g.adminPlayer(c)
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

func Show(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		cu := user.CurrentFrom(c)
		c.HTML(http.StatusOK, prefix+"/show", gin.H{
			"Context":    c,
			"VersionID":  sn.VersionID(),
			"CUser":      cu,
			"Game":       g,
			"IsAdmin":    user.IsAdmin(c),
			"Admin":      game.AdminFrom(c),
			"MessageLog": mlog.From(c),
			"ColorMap":   color.MapFrom(c),
			"Notices":    restful.NoticesFrom(c),
			"Errors":     restful.ErrorsFrom(c),
		})
	}
}

func Update(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		if g == nil {
			log.Errorf("Controller#Update Game Not Found")
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
		template, actionType, err := g.Update(c)
		log.Debugf("template: %v actionType: %v err: %v", template, actionType, err)
		switch {
		case err != nil && sn.IsVError(err):
			restful.AddErrorf(c, "%v", err)
			withJSON(c, g)
		case err != nil:
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		case actionType == game.Cache:
			if err := g.cache(c); err != nil {
				restful.AddErrorf(c, "%v", err)
			}
		case actionType == game.Save:
			if err := g.save(c); err != nil {
				log.Errorf("%s", err)
				restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(c, prefix))
				return
			}
		case actionType == game.Undo:
			mkey := g.UndoKey(c)
			if err := memcache.Delete(c, mkey); err != nil && err != memcache.ErrCacheMiss {
				log.Errorf("memcache.Delete error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(c, prefix))
				return
			}
		}

		switch jData := jsonFrom(c); {
		case jData != nil && template == "json":
			c.JSON(http.StatusOK, jData)
		case template == "":
			c.Redirect(http.StatusSeeOther, showPath(c, prefix))
		default:
			cu := user.CurrentFrom(c)

			d := gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Game":      g,
				"IsAdmin":   user.IsAdmin(c),
				"Notices":   restful.NoticesFrom(c),
				"Errors":    restful.ErrorsFrom(c),
			}
			c.HTML(http.StatusOK, template, d)
		}
	}
}

func (g *Game) save(c *gin.Context) error {
	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

		err = memcache.Delete(c, g.UndoKey(c))
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return err
	})
	return err
}

func (g *Game) saveWith(c *gin.Context, ks []*datastore.Key, es []interface{}) error {
	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

		err = memcache.Delete(c, g.UndoKey(c))
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return err
	})
	return err
}

func (g *Game) encode(cx *gin.Context) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var encoded []byte
	if encoded, err = codec.Encode(g.State); err != nil {
		return
	}
	g.SavedState = encoded
	g.updateHeader()

	return
}

func (g *Game) cache(c *gin.Context) error {
	item := &memcache.Item{
		Key:        g.UndoKey(c),
		Expiration: time.Minute * 30,
	}
	v, err := codec.Encode(g)
	if err != nil {
		return err
	}
	item.Value = v
	return memcache.Set(c, item)
}

func wrap(s *stats.Stats, cs contest.Contests) ([]*datastore.Key, []interface{}) {
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

func Undo(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		c.Redirect(http.StatusSeeOther, showPath(c, prefix))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("Controller#Update Game Not Found")
			return
		}
		mkey := g.UndoKey(c)
		if err := memcache.Delete(c, mkey); err != nil && err != memcache.ErrCacheMiss {
			log.Errorf("Controller#Undo Error: %s", err)
		}
	}
}

func EndRound(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		c.Redirect(http.StatusSeeOther, showPath(c, prefix))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			return
		}
		g.endOfRoundPhase(c)
		if err := g.save(c); err != nil {
			log.Errorf("cache error: %s", err.Error())
		}
		return
	}
}

func Index(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		gs := game.GamersFrom(c)
		switch status := game.StatusFrom(c); status {
		case game.Recruiting:
			c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     user.CurrentFrom(c),
				"Games":     gs,
				"Type":      gtype.Confucius.String(),
			})
		default:
			c.HTML(http.StatusOK, "shared/games_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     user.CurrentFrom(c),
				"Games":     gs,
				"Type":      gtype.Confucius.String(),
				"Status":    status,
			})
		}
	}
}
func NewAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := New(c, 0)
		withGame(c, g)
		if err := g.FromParams(c, gtype.GOT); err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		c.HTML(http.StatusOK, prefix+"/new", gin.H{
			"Context":   c,
			"VersionID": sn.VersionID(),
			"CUser":     user.CurrentFrom(c),
			"Game":      g,
		})
	}
}

func Create(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		dsClient, err := datastore.NewClient(c, "")
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		g := New(c, 0)
		withGame(c, g)

		err = g.FromParams(c, g.Type)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.fromForm(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.encode(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		ks, err := dsClient.AllocateIDs(c, []*datastore.Key{g.Key})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		k := ks[0]

		_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

func Accept(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			return
		}

		var (
			start bool
			err   error
		)

		u := user.CurrentFrom(c)
		if start, err = g.Accept(c, u); err == nil && start {
			err = g.Start(c)
		}

		if err == nil {
			err = g.save(c)
		}

		if err == nil && start {
			g.SendTurnNotificationsTo(c, g.CurrentPlayer())
		}

		if err != nil {
			log.Errorf(err.Error())
		}

	}
}

func Drop(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			return
		}

		var err error

		u := user.CurrentFrom(c)
		if err = g.Drop(u); err == nil {
			err = g.save(c)
		}

		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
		}

	}
}

func Fetch(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var (
		id  int64
		err error
	)

	// create Gamer
	if id, err = strconv.ParseInt(c.Param("hid"), 10, 64); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	g := New(c, id)

	switch action := c.PostForm("action"); {
	case action == "reset":
		// pull from memcache/datastore
		// same as undo & !MultiUndo
		fallthrough
	case action == "undo":
		// pull from memcache/datastore
		if err = dsGet(c, g); err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	default:
		if user.CurrentFrom(c) != nil {
			// pull from memcache and return if successful; otherwise pull from datastore
			if err := mcGet(c, g); err == nil {
				return
			}
		}
		if err = dsGet(c, g); err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	}
}

// pull temporary game state from memcache.  Note may be different from value stored in datastore.
func mcGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	mkey := g.GetHeader().UndoKey(c)
	item, err := memcache.Get(c, mkey)
	if err != nil {
		return err
	}

	err = codec.Decode(g, item.Value)
	if err != nil {
		return err
	}

	err = g.AfterCache()
	if err != nil {
		return err
	}

	color.WithMap(withGame(c, g), g.ColorMapFor(user.CurrentFrom(c)))
	return nil
}

// pull game state from memcache/datastore.  returned memcache should be same as datastore.
func dsGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	switch err = dsClient.Get(c, g.Key, g.Header); {
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

	err = g.init(c)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}

	cm := g.ColorMapFor(user.CurrentFrom(c))
	color.WithMap(withGame(c, g), cm)
	return nil
}

func JSON(c *gin.Context) {
	c.JSON(http.StatusOK, gameFrom(c))
}

func JSONIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		game.JSONIndexAction(c)
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
	if u := g.Creator; u != nil {
		g.CreatorSID = user.GenID(u.GoogleID)
		g.CreatorName = u.Name
	}

	if l := len(g.Users); l > 0 {
		g.UserSIDS = make([]string, l)
		g.UserNames = make([]string, l)
		g.UserEmails = make([]string, l)
		for i, u := range g.Users {
			g.UserSIDS[i] = user.GenID(u.GoogleID)
			g.UserNames[i] = u.Name
			g.UserEmails[i] = u.Email
		}
	}
}

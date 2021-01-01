package confucius

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("*game.moveJunksEntry", new(moveJunksEntry))
	gob.RegisterName("*game.replaceStudentEntry", new(replaceStudentEntry))
	gob.RegisterName("*game.swapOfficialsEntry", new(swapOfficialsEntry))
	gob.RegisterName("*game.redeployArmyEntry", new(redeployArmyEntry))
	gob.RegisterName("*game.replaceInfluenceEntry", new(replaceInfluenceEntry))
}

func (g *Game) EnablePetitionEmperor(cu *user.User) bool {
	cp := g.CurrentPlayer()
	requiredCubes := cp.RequiredCubesFor(PetitionSpace)
	return g.IsCurrentPlayer(cu) && cp.ActionCubes >= requiredCubes && cp.hasPetitionGift() && !g.BasicGame
}

func (p *Player) hasPetitionGift() bool {
	for _, gift := range p.GiftsBought {
		if gift.Value > 1 {
			return true
		}
	}
	return false
}

func (g *Game) PetitionDirections() []string {
	return []string{
		"Select gift with which to petition the Emperor.",
		"Petitioning Emperor costs two action cubes.",
	}
}

func (g *GiftCard) PetitionDirections() []string {
	switch g.Value {
	case Tile:
		return []string{
			"Move 2 junks from any player's shipyard to any other player's shipyard, including your own.",
			"Does not affect junks in ocean spaces."}
	case Vase:
		return []string{
			"Remove your choice of marker from a student space in the Imperial Examinations Box and put one of your own markers in its place.",
			"If you already have a student, this will result in your family having both students."}
	case Coat:
		return []string{
			"Swap an official tile with your marker on it with another official tile of equal or lower seniority (7 is the lowest).",
			"Both officials must be in unresolved ministries and can be from the same or different ones."}
	case Necklace:
		return []string{
			"Move any single army piece on an unresolved Foreign Land to another box on any unresolved Foreign Land tile, including the one it is currently on."}
	case Junk:
		return []string{
			"Replace any unsecured marker on an official with a secured marker from any player's supply, including your own."}
	}
	return nil
}

func (g *Game) moveJunks(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	fromPlayer, toPlayer, junks, cubes, err := g.validateMoveJunks(c, cu)
	if err != nil {
		return "", game.None, err
	}

	// Place Action Cubes
	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.PlaceCubesIn(PetitionSpace, cubes)

	// Move Junks
	fromPlayer.Junks -= junks
	toPlayer.Junks += junks

	// Remove Tile Gift
	cp.GiftsBought.Remove(cp.GetBoughtGift(Tile))

	// Create Action Object for logging
	e := cp.newMoveJunksEntry(junks, fromPlayer, toPlayer)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type moveJunksEntry struct {
	*Entry
	Junks        int
	FromPlayerID int
	ToPlayerID   int
}

func (p *Player) newMoveJunksEntry(j int, fp, tp *Player) *moveJunksEntry {
	g := p.Game()
	e := new(moveJunksEntry)
	e.Entry = p.newEntry()
	e.Junks = j
	e.FromPlayerID = fp.ID()
	e.OtherPlayerID = tp.ID()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *moveJunksEntry) HTML() template.HTML {
	g := e.Game().(*Game)
	return restful.HTML("%s used value 2 gift (Tile) to petition Emperor and move %d %s from %s to %s.",
		g.NameByPID(e.PlayerID), e.Junks, pluralize("junk", e.Junks), g.NameByPID(e.FromPlayerID), g.NameByPID(e.ToPlayerID))
}

func (g *Game) validateMoveJunks(c *gin.Context, cu *user.User) (*Player, *Player, int, int, error) {
	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	fromPlayer := g.PlayerBySID(c.PostForm("move-junks-from-player"))
	toPlayer := g.PlayerBySID(c.PostForm("move-junks-to-player"))

	cp := g.CurrentPlayer()

	switch {
	case g.BasicGame:
		return nil, nil, 0, 0, sn.NewVError("You cannot petition the emperor in the basic game.")
	case cp.GetBoughtGift(Tile) == nil:
		return nil, nil, 0, 0, sn.NewVError("You don't have a value 2 (Tile) gift with which to petition the Emperor.")
	case fromPlayer == nil:
		return nil, nil, 0, 0, sn.NewVError("You must select a player from which to move junks.")
	case toPlayer == nil:
		return nil, nil, 0, 0, sn.NewVError("You must select a player to which to move junks.")
	case !fromPlayer.hasJunks():
		return nil, nil, 0, 0, sn.NewVError("%s has no junks to move.", g.NameFor(fromPlayer))
	case fromPlayer.Junks == 1:
		return fromPlayer, toPlayer, 1, cubes, nil
	}
	return fromPlayer, toPlayer, 2, cubes, nil
}

func (g *Game) replaceStudent(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	p, cubes, err := g.validateReplaceStudent(c, cu)
	if err != nil {
		return "", game.None, err
	}

	// Place Action Cubes
	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.PlaceCubesIn(PetitionSpace, cubes)

	// Replace Student
	if g.Candidate().Player() != nil && p.Equal(g.Candidate().Player()) {
		g.Candidate().setPlayer(cp)
	} else {
		g.Candidate().setOtherPlayer(cp)
	}

	// Remove Vase Gift
	cp.GiftsBought.Remove(cp.GetBoughtGift(Vase))

	// Create Action Object for logging
	e := g.NewReplaceStudentEntry(cp)
	e.OtherPlayerID = p.ID()

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type replaceStudentEntry struct {
	*Entry
}

func (g *Game) NewReplaceStudentEntry(p *Player) *replaceStudentEntry {
	e := new(replaceStudentEntry)
	e.Entry = p.newEntry()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *replaceStudentEntry) HTML() template.HTML {
	return restful.HTML("%s used value 3 gift (Vase) to petition Emperor and replace student of %s with own student.",
		e.Player().Name(), e.OtherPlayer().Name())
}

func (g *Game) validateReplaceStudent(c *gin.Context, cu *user.User) (*Player, int, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cbs, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, 0, err
	}

	p := g.PlayerBySID(c.PostForm("replace-student-player"))
	cp := g.CurrentPlayer()

	switch {
	case p == nil:
		return nil, 0, sn.NewVError("Selected player not found.")
	case g.BasicGame:
		return nil, 0, sn.NewVError("You cannot petition the emperor in the basic game.")
	case cp.Equal(p):
		return nil, 0, sn.NewVError("You did not select a marker of another player.")
	case g.Candidate().Player() == nil && g.Candidate().OtherPlayer() == nil:
		return nil, 0, sn.NewVError("There is no student to replace.")
	case cp.GetBoughtGift(Vase) == nil:
		return nil, 0, sn.NewVError("You don't have a value 3 (Vase) gift with which to petition the Emperor.")
	case p.Equal(g.Candidate().Player()):
		return p, cbs, nil
	case p.Equal(g.Candidate().OtherPlayer()):
		return p, cbs, nil
	}
	return nil, 0, sn.NewVError("Selected player does not have a student.")
}

func (g *Game) swapOfficials(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	ministry1, ministry2, official1, official2, cubes, err := g.validateSwapOfficials(c, cu)
	if err != nil {
		return "", game.None, err
	}

	// Place Action Cubes
	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.PlaceCubesIn(PetitionSpace, cubes)

	// Remove Coat Gift
	cp.GiftsBought.Remove(cp.GetBoughtGift(Coat))

	// Create Action Object for logging
	e := cp.newSwapOfficialsEntry(ministry1, ministry2, official1, official2)

	official1.Seniority, official2.Seniority = official2.Seniority, official1.Seniority
	ministry1.Officials[official2.Seniority], ministry2.Officials[official1.Seniority] = official2, official1

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type swapOfficialsEntry struct {
	*Entry
	MinistryName1 string
	MinistryName2 string
	Seniority1    Seniority
	Seniority2    Seniority
}

func (p *Player) newSwapOfficialsEntry(m1, m2 *Ministry, o1, o2 *OfficialTile) *swapOfficialsEntry {
	g := p.Game()
	e := new(swapOfficialsEntry)
	e.Entry = p.newEntry()
	e.MinistryName1 = m1.Name()
	e.Seniority1 = o1.Seniority
	e.MinistryName2 = m2.Name()
	e.Seniority2 = o2.Seniority
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (g *swapOfficialsEntry) HTML() template.HTML {
	return restful.HTML("%s used value 4 gift (Coat) to swap %s official with %d seniority with %s official with %d seniority.",
		g.Player().Name(), g.MinistryName1, g.Seniority1, g.MinistryName2, g.Seniority2)
}

func (g *Game) validateSwapOfficials(c *gin.Context, cu *user.User) (*Ministry, *Ministry, *OfficialTile, *OfficialTile, int, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	ministry1, official1, err := g.getMinistryAndOfficial(c, "swap-your-official")
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	ministry2, official2, err := g.getMinistryAndOfficial(c, "swap-other-official")
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	cp := g.CurrentPlayer()
	switch {
	case g.BasicGame:
		return nil, nil, nil, nil, 0, sn.NewVError("You cannot petition the emperor in the basic game.")
	case ministry1.Resolved || ministry2.Resolved:
		return nil, nil, nil, nil, 0, sn.NewVError("You selected an official from a resolved ministry.")
	case cp.GetBoughtGift(Coat) == nil:
		return nil, nil, nil, nil, 0, sn.NewVError("You don't have a value 4 (Coat) gift with which to petition the Emperor.")
	case official1.Variant == NoOfficial || official2.Variant == NoOfficial:
		return nil, nil, nil, nil, 0, sn.NewVError("You selected a space without an official.")
	case official1.NotBribed():
		return nil, nil, nil, nil, 0, sn.NewVError("You selected an official without a marker.")
	case official1.Player().NotEqual(cp):
		return nil, nil, nil, nil, 0, sn.NewVError("You did not select one of your officials to swap.")
	case official2.Seniority < official1.Seniority:
		return nil, nil, nil, nil, 0, sn.NewVError("You selected an official of another player with a higher seniority.")
	}
	return ministry1, ministry2, official1, official2, cubes, err
}

func (g *Game) redeployArmy(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	fromBox, toBox, cubes, err := g.validateRedeployArmy(c, cu)
	if err != nil {
		return "", game.None, err
	}

	// Place Action Cubes
	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.PlaceCubesIn(PetitionSpace, cubes)

	// Redeploy Army
	fromBox.setPlayer(nil)
	toBox.setPlayer(cp)

	// Remove Necklace Gift
	cp.GiftsBought.Remove(cp.GetBoughtGift(Necklace))

	// Create Action Object for logging
	e := cp.newRedeployArmyEntry(fromBox, toBox)

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type redeployArmyEntry struct {
	*Entry
	FromForeignLandName string
	ToForeignLandName   string
	FromBox             int
	ToBox               int
}

func (p *Player) newRedeployArmyEntry(fromBox, toBox *ForeignLandBox) *redeployArmyEntry {
	g := p.Game()
	e := new(redeployArmyEntry)
	e.Entry = p.newEntry()
	e.FromForeignLandName = fromBox.land.Name()
	e.ToForeignLandName = toBox.land.Name()
	e.FromBox = fromBox.Points
	e.ToBox = toBox.Points
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *redeployArmyEntry) HTML() template.HTML {
	return restful.HTML("%s used value 5 gift (Necklace) to redeploy army from %d point box of %s to %d point box of %s.", e.Player().Name(), e.FromBox, e.FromForeignLandName, e.ToBox, e.ToForeignLandName)
}

func (g *Game) validateRedeployArmy(c *gin.Context, cu *user.User) (*ForeignLandBox, *ForeignLandBox, int, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, nil, 0, err
	}

	fromBox, err := g.getForeignLandBox(c, "from-land")
	if err != nil {
		return nil, nil, 0, err
	}

	toBox, err := g.getForeignLandBox(c, "to-land")
	if err != nil {
		return nil, nil, 0, err
	}

	switch {
	case g.BasicGame:
		return nil, nil, 0, sn.NewVError("You cannot petition the emperor in the basic game.")
	case g.CurrentPlayer().GetBoughtGift(Necklace) == nil:
		return nil, nil, 0, sn.NewVError("You don't have a value 5 (Necklace) gift with which to petition the Emperor.")
	case fromBox == nil:
		return nil, nil, 0, sn.NewVError("You must select a land box from which to redeploy an army.")
	case fromBox.Player() == nil || !fromBox.Player().IsCurrentPlayer():
		return nil, nil, 0, sn.NewVError("You don't have an army in the selected box.")
	case fromBox.land.Resolved:
		return nil, nil, 0, sn.NewVError("You can't redeploy an army from a resolved foreign land tile.")
	case toBox == nil:
		return nil, nil, 0, sn.NewVError("You must select a land box to which to redeploy an army.")
	case toBox.Player() != nil:
		return nil, nil, 0, sn.NewVError("The selected land box already has an army.")
	case toBox.land.Resolved:
		return nil, nil, 0, sn.NewVError("You can't redeploy an army to a resolved foreign land tile.")
	}

	return fromBox, toBox, cubes, nil
}

func (g *Game) replaceInfluence(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	ministry, official, player, cubes, err := g.validateReplaceInfluence(c, cu)
	if err != nil {
		return "", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true

	// Create Action Object for logging
	e := cp.newReplaceInfluenceEntry(ministry, official, player)

	// Replace Influence
	official.setPlayer(player)
	official.Secured = true

	// Place Action Cubes
	cp.PlaceCubesIn(PetitionSpace, cubes)

	// Remove Junk Gift
	cp.GiftsBought.Remove(cp.GetBoughtGift(Junk))

	// Set flash message
	restful.AddNoticef(c, string(e.HTML()))
	return "", game.Cache, nil
}

type replaceInfluenceEntry struct {
	*Entry
	MinistryName string
	Seniority    Seniority
	FromPlayerID int
	ToPlayerID   int
}

func (p *Player) newReplaceInfluenceEntry(m *Ministry, o *OfficialTile, tp *Player) *replaceInfluenceEntry {
	g := p.Game()
	e := new(replaceInfluenceEntry)
	e.Entry = p.newEntry()
	e.MinistryName = m.Name()
	e.Seniority = o.Seniority
	e.FromPlayerID = o.Player().ID()
	e.ToPlayerID = tp.ID()
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *replaceInfluenceEntry) HTML() template.HTML {
	g := e.Game().(*Game)
	return restful.HTML("%s used value 6 gift (Junk) to replace unsecured marker of %s on %s official with %d seniority with a secured marker of %s.",
		g.NameByPID(e.PlayerID), g.NameByPID(e.FromPlayerID), e.MinistryName, e.Seniority, g.NameByPID(e.ToPlayerID))
}

func (g *Game) validateReplaceInfluence(c *gin.Context, cu *user.User) (*Ministry, *OfficialTile, *Player, int, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cubes, err := g.validatePlayerAction(c, cu)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	ministry, official, err := g.getMinistryAndOfficial(c, "replace-influence-official")
	if err != nil {
		return nil, nil, nil, 0, err
	}

	cp := g.CurrentPlayer()
	player := g.PlayerBySID(c.PostForm("replace-influence-player"))

	switch {
	case g.BasicGame:
		return nil, nil, nil, 0, sn.NewVError("You cannot petition the emperor in the basic game.")
	case player == nil:
		return nil, nil, nil, 0, sn.NewVError("Selected player not found.")
	case ministry.Resolved:
		return nil, nil, nil, 0, sn.NewVError("You selected an official from a resolved ministry.")
	case official.Secured:
		return nil, nil, nil, 0, sn.NewVError("You selected a secured official.")
	case cp.GetBoughtGift(Junk) == nil:
		return nil, nil, nil, 0, sn.NewVError("You don't have a value 6 (Junk) gift with which to petition the Emperor.")
	case official.Variant == NoOfficial:
		return nil, nil, nil, 0, sn.NewVError("You selected a space without an official.")
	case official.NotBribed():
		return nil, nil, nil, 0, sn.NewVError("You selected an official without a marker.")
	}

	return ministry, official, player, cubes, nil
}

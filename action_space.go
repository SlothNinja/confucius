package confucius

type ActionSpace struct {
	ID    SpaceID
	Cubes Cubes
}

type Cubes map[int]int
type ActionSpaces map[SpaceID]*ActionSpace

type SpaceID int

const (
	NoSpace SpaceID = iota
	BribeSecureSpace
	NominateSpace
	ForceSpace
	JunksVoyageSpace
	RecruitArmySpace
	BuyGiftSpace
	GiveGiftSpace
	PetitionSpace
	CommercialSpace
	TaxIncomeSpace
	NoActionSpace
	ImperialFavourSpace
)

var spaceIDStrings = map[SpaceID]string{
	BribeSecureSpace:    "BribeSecureSpace",
	NominateSpace:       "NominateSpace",
	ForceSpace:          "ForceSpace",
	JunksVoyageSpace:    "JunksVoyageSpace",
	RecruitArmySpace:    "RecruitArmySpace",
	BuyGiftSpace:        "BuyGiftSpace",
	GiveGiftSpace:       "GiveGiftSpace",
	PetitionSpace:       "PetitionSpace",
	CommercialSpace:     "CommercialSpace",
	TaxIncomeSpace:      "TaxIncomeSpace",
	NoActionSpace:       "NoActionSpace",
	ImperialFavourSpace: "ImperialFavourSpace",
}

func (this SpaceID) String() string {
	return spaceIDStrings[this]
}

func (this SpaceID) Int() int {
	return int(this)
}

func (this *ActionSpace) Name() string {
	return this.ID.String()
}

func (this *Player) CubesIn(space *ActionSpace) int {
	return space.Cubes[this.ID()]
}

func (p *Player) PlaceCubesIn(id SpaceID, cubes int) {
	space := p.Game().ActionSpaces[id]
	p.ActionCubes -= cubes
	space.Cubes[p.ID()] += cubes
}

func (p *Player) RequiredCubesFor(id SpaceID) int {
	switch space := p.Game().ActionSpaces[id]; {
	case space == nil:
		return 0
	case p.Game().Phase == Actions:
		switch {
		case p.Game().ExtraAction:
			return 0
		case id == NoActionSpace:
			return 1
		case space.ID == PetitionSpace, space.Cubes[p.ID()] > 0:
			return 2
		default:
			return 1
		}
	case p.Game().Phase == ImperialFavour:
		return 1
	}
	return 1
}

func (p *Player) hasEnoughCubesFor(id SpaceID) bool {
	return p.ActionCubes >= p.RequiredCubesFor(id)
}

func (p *Player) hasConCards() bool {
	return len(p.ConCardHand) > 0
}

package confucius

import "github.com/SlothNinja/game"

const (
	NoPhase game.Phase = iota
	Actions
	ImperialFavour
	EndOfRound
	EndGameScoring
	GameOver
	AwardChiefMinister
	AwardAdmiral
	AwardGeneral
	AnnounceWinners
	Discard
	ChooseChiefMinister
	ExaminationResolution
	MinistryResolution
	FinalMinistryResolution
	ImperialExamination
	Invasion
	CountGifts
	Setup
	StartGame
	ReturnCubesPhase
	BuildWall
	EndOfGamePhase
	PlaceNewOfficial
	NewRound
	ResolveMinistry
	TempTransfer
	AutoTempTransfer
	InitMinistryResolution
	AutoTutor
	AwaitPlayerInput
)

var PhaseNames = game.PhaseNameMap{
	NoPhase:                 "None",
	Actions:                 "Actions",
	ImperialFavour:          "Imperial Favour",
	EndOfRound:              "End Of Round",
	EndGameScoring:          "End Game Scoring",
	GameOver:                "Game Over",
	AwardChiefMinister:      "Award Chief Minister",
	AwardAdmiral:            "Award Admiral",
	AwardGeneral:            "Award General",
	AnnounceWinners:         "Announce Winners",
	Discard:                 "Discard",
	ChooseChiefMinister:     "Choose Chief Minister",
	ExaminationResolution:   "Examination Resolution",
	MinistryResolution:      "Ministry Resolution",
	FinalMinistryResolution: "Final Ministry Resolution",
	ImperialExamination:     "Imperial Examination",
	Invasion:                "Invasion",
	CountGifts:              "Count Gifts",
	Setup:                   "Setup",
	StartGame:               "Start Game",
	ReturnCubesPhase:        "Return Cubes",
	BuildWall:               "Build Wall",
	EndOfGamePhase:          "End Of Game",
	PlaceNewOfficial:        "Place New Official",
	NewRound:                "New Round Phase",
	ResolveMinistry:         "Resolve Ministry",
	TempTransfer:            "Transfer Influence",
	AutoTempTransfer:        "Auto Temp Transfer",
	InitMinistryResolution:  "Initialize Ministry Resolution",
	AutoTutor:               "Auto-Tutor Student",
	AwaitPlayerInput:        "Awaiting Player Input",
}

func (g *Game) PhaseName() (n string) {
	n, _ = PhaseNames[g.Phase]
	return
}

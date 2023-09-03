package main

import "log"

type StatusState string
type MediaListStatus string

var (
	AnnictWatchingStatus   = StatusState("WATCHING")
	AnnictWatchedStatus    = StatusState("WATCHED")
	AnnictWannaWatchStatus = StatusState("WANNA_WATCH")
	AnnictOnHoldStatus     = StatusState("ON_HOLD")
	AnnictStopWatching     = StatusState("STOP_WATCHING")
	AnnictNoState          = StatusState("NO_STATE")
	AniListCurrentStatus   = MediaListStatus("CURRENT")
	AniListCompletedStatus = MediaListStatus("COMPLETED")
	AniListPlanningStatus  = MediaListStatus("PLANNING")
	AniListPausedStatus    = MediaListStatus("PAUSED")
	AniListDroppedStatus   = MediaListStatus("DROPPED")
	AniListRepeatingStatus = MediaListStatus("REPEATING")
)

func IsSameListStatus(annict StatusState, aniList MediaListStatus) bool {
	return annict.ToAniListStatus() == aniList && annict == aniList.ToAnnictStatus()
}

func (s StatusState) ToAniListStatus() MediaListStatus {
	switch s {
	case AnnictWatchingStatus:
		return AniListCurrentStatus
	case AnnictWatchedStatus:
		return AniListCompletedStatus
	case AnnictWannaWatchStatus:
		return AniListPlanningStatus
	case AnnictOnHoldStatus:
		return AniListPausedStatus
	case AnnictStopWatching:
		return AniListDroppedStatus
	default:
		log.Fatalf("unexpected StatusState: %s", s)
		panic("unreachable")
	}
}

func (s MediaListStatus) ToAnnictStatus() StatusState {
	switch s {
	case AniListCurrentStatus:
		return AnnictWatchingStatus
	case AniListCompletedStatus:
		return AnnictWatchedStatus
	case AniListPlanningStatus:
		return AnnictWannaWatchStatus
	case AniListPausedStatus:
		return AnnictOnHoldStatus
	case AniListDroppedStatus:
		return AnnictStopWatching
	case AniListRepeatingStatus:
		// Repeating は Watching 扱いとする
		return AnnictWatchingStatus
	default:
		log.Fatalf("unexpected MediaListStatus: %s", s)
		panic("unreachable")
	}
}

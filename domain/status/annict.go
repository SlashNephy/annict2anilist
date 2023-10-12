package status

import "fmt"

type AnnictStatusState string

const (
	AnnictWatching     = AnnictStatusState("WATCHING")
	AnnictWatched      = AnnictStatusState("WATCHED")
	AnnictWannaWatch   = AnnictStatusState("WANNA_WATCH")
	AnnictOnHold       = AnnictStatusState("ON_HOLD")
	AnnictStopWatching = AnnictStatusState("STOP_WATCHING")
)

func (s AnnictStatusState) ToAniListStatus() AniListMediaListStatus {
	switch s {
	case AnnictWatching:
		return AniListCurrent
	case AnnictWatched:
		return AniListCompleted
	case AnnictWannaWatch:
		return AniListPlanning
	case AnnictOnHold:
		return AniListPaused
	case AnnictStopWatching:
		return AniListDropped
	default:
		panic(fmt.Sprintf("unexpected status: %s", s))
	}
}

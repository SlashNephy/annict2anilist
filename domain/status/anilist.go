package status

import "fmt"

type AniListMediaListStatus string

var (
	AniListCurrent   = AniListMediaListStatus("CURRENT")
	AniListCompleted = AniListMediaListStatus("COMPLETED")
	AniListPlanning  = AniListMediaListStatus("PLANNING")
	AniListPaused    = AniListMediaListStatus("PAUSED")
	AniListDropped   = AniListMediaListStatus("DROPPED")
	AniListRepeating = AniListMediaListStatus("REPEATING")
)

func (s AniListMediaListStatus) ToAnnictStatus() AnnictStatusState {
	switch s {
	case AniListCurrent:
		return AnnictWatching
	case AniListCompleted:
		return AnnictWatched
	case AniListPlanning:
		return AnnictWannaWatch
	case AniListPaused:
		return AnnictOnHold
	case AniListDropped:
		return AnnictStopWatching
	case AniListRepeating:
		// Repeating は Watching 扱いとする
		return AnnictWatching
	default:
		panic(fmt.Sprintf("unexpected status: %s", s))
	}
}

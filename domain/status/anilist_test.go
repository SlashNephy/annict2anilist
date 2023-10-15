package status

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAniListMediaListStatus_ToAnnictStatus(t *testing.T) {
	t.Run("ステータスを相互変換できる", func(t *testing.T) {
		tests := []struct {
			anilist AniListMediaListStatus
			annict  AnnictStatusState
		}{
			{
				anilist: AniListCurrent,
				annict:  AnnictWatching,
			},
			{
				anilist: AniListCompleted,
				annict:  AnnictWatched,
			},
			{
				anilist: AniListPlanning,
				annict:  AnnictWannaWatch,
			},
			{
				anilist: AniListPaused,
				annict:  AnnictOnHold,
			},
			{
				anilist: AniListDropped,
				annict:  AnnictStopWatching,
			},
			{
				anilist: AniListRepeating,
				annict:  AnnictWatching,
			},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s は %s と等価である", tt.anilist, tt.annict), func(t *testing.T) {
				actual := tt.anilist.ToAnnictStatus()
				assert.Equal(t, tt.annict, actual)
			})
		}
	})

	t.Run("未知のステータスで panic する", func(t *testing.T) {
		status := AniListMediaListStatus("UNKNOWN")
		assert.Panics(t, func() {
			status.ToAnnictStatus()
		})
	})
}

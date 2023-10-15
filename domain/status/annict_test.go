package status

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnictStatusState_ToAniListStatus(t *testing.T) {
	t.Run("ステータスを相互変換できる", func(t *testing.T) {
		tests := []struct {
			annict  AnnictStatusState
			anilist AniListMediaListStatus
		}{
			{
				annict:  AnnictWatching,
				anilist: AniListCurrent,
			},
			{
				annict:  AnnictWatched,
				anilist: AniListCompleted,
			},
			{
				annict:  AnnictWannaWatch,
				anilist: AniListPlanning,
			},
			{
				annict:  AnnictOnHold,
				anilist: AniListPaused,
			},
			{
				annict:  AnnictStopWatching,
				anilist: AniListDropped,
			},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s は %s と等価である", tt.annict, tt.anilist), func(t *testing.T) {
				actual := tt.annict.ToAniListStatus()
				assert.Equal(t, tt.anilist, actual)
			})
		}
	})

	t.Run("未知のステータスで panic する", func(t *testing.T) {
		status := AnnictStatusState("UNKNOWN")
		assert.Panics(t, func() {
			status.ToAniListStatus()
		})
	})
}

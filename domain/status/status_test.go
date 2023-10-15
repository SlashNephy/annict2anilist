package status

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSameListStatus(t *testing.T) {
	t.Run("ステータスを相互変換できる", func(t *testing.T) {
		tests := []struct {
			annict  AnnictStatusState
			aniList AniListMediaListStatus
		}{
			{
				annict:  AnnictWatching,
				aniList: AniListCurrent,
			},
			{
				annict:  AnnictWatched,
				aniList: AniListCompleted,
			},
			{
				annict:  AnnictWannaWatch,
				aniList: AniListPlanning,
			},
			{
				annict:  AnnictOnHold,
				aniList: AniListPaused,
			},
			{
				annict:  AnnictStopWatching,
				aniList: AniListDropped,
			},
			{
				annict:  AnnictWatching,
				aniList: AniListRepeating,
			},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s と %s は等価である", tt.annict, tt.aniList), func(t *testing.T) {
				assert.True(t, IsSameListStatus(tt.annict, tt.aniList))
			})
		}
	})

	t.Run("未知のステータスで panic する", func(t *testing.T) {
		t.Run("Annict 側のステータスが未知", func(t *testing.T) {
			assert.Panics(t, func() {
				IsSameListStatus("UNKNOWN", AniListCurrent)
			})
		})

		t.Run("AniList 側のステータスが未知", func(t *testing.T) {
			assert.Panics(t, func() {
				IsSameListStatus(AnnictWatching, "UNKNOWN")
			})
		})
	})
}

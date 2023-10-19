package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SlashNephy/annict2anilist/domain/status"
	"github.com/SlashNephy/annict2anilist/external/anilist"
	"github.com/SlashNephy/annict2anilist/external/annict"
	"github.com/SlashNephy/annict2anilist/external/arm"
)

const (
	dummyAnnictID  = 1
	dummyAniListID = 2
)

func createEpisodeConnection(numWatched int) annict.EpisodeConnection {
	var edges []annict.EpisodeEdge
	for i := 0; i < numWatched; i++ {
		edges = append(edges, annict.EpisodeEdge{
			Node: annict.Episode{
				ViewerDidTrack: true,
			},
		})
	}

	return annict.EpisodeConnection{
		Edges: edges,
	}
}

func TestCalculateDiff(t *testing.T) {
	t.Run("Annict ID から AniList ID を参照できない", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID: dummyAnnictID,
					Title:    "葬送のフリーレン",
				},
			},
			[]anilist.LibraryEntry{},
			&arm.ArmDatabase{},
		)

		assert.Len(t, actual.AniListUpdates, 0)
		assert.Len(t, actual.Untethered, 1)
		assert.Equal(t, "Annict", actual.Untethered[0].Source)
		assert.Equal(t, dummyAnnictID, actual.Untethered[0].ID)
		assert.Equal(t, "葬送のフリーレン", actual.Untethered[0].Title)
	})

	t.Run("差分が存在せず、更新の必要はない", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID:          dummyAnnictID,
					Title:             "陰の実力者になりたくて！ 2nd season",
					ViewerStatusState: status.AnnictWatching,
					NoEpisodes:        false,
					Episodes:          createEpisodeConnection(3),
				},
			},
			[]anilist.LibraryEntry{
				{
					Status:   status.AniListCurrent,
					Progress: 3,
					Media: anilist.Media{
						ID: dummyAniListID,
						Title: anilist.Title{
							Native: "陰の実力者になりたくて！ 2nd season",
						},
					},
				},
			},
			&arm.ArmDatabase{
				Entries: []arm.ArmEntry{
					{
						AnnictID:  dummyAnnictID,
						AniListID: dummyAniListID,
					},
				},
			})

		assert.Len(t, actual.AniListUpdates, 0)
		assert.Len(t, actual.Untethered, 0)
	})

	t.Run("劇場版などエピソード区分がないものは視聴済みのエピソード数を 1 とする", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID:          dummyAnnictID,
					Title:             "かがみの孤城",
					ViewerStatusState: status.AnnictWatched,
					NoEpisodes:        true,
					Episodes:          createEpisodeConnection(1),
				},
			},
			[]anilist.LibraryEntry{
				{
					Status:   status.AniListCompleted,
					Progress: 1,
					Media: anilist.Media{
						ID: dummyAniListID,
						Title: anilist.Title{
							Native: "かがみの孤城",
						},
					},
				},
			},
			&arm.ArmDatabase{
				Entries: []arm.ArmEntry{
					{
						AnnictID:  dummyAnnictID,
						AniListID: dummyAniListID,
					},
				},
			})

		assert.Len(t, actual.AniListUpdates, 0)
		assert.Len(t, actual.Untethered, 0)
	})

	t.Run("作品が終了していて、どちらのステータスも Completed になっている場合は Progress の更新を行わない", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID:          dummyAnnictID,
					Title:             "ぼっち・ざ・ろっく！",
					ViewerStatusState: status.AnnictWatched,
					NoEpisodes:        false,
					Episodes:          createEpisodeConnection(13),
				},
			},
			[]anilist.LibraryEntry{
				{
					Status:   status.AniListCompleted,
					Progress: 12,
					Media: anilist.Media{
						ID: dummyAniListID,
						Title: anilist.Title{
							Native: "ぼっち・ざ・ろっく！",
						},
						Status: anilist.MediaStatusFinished,
					},
				},
			},
			&arm.ArmDatabase{
				Entries: []arm.ArmEntry{
					{
						AnnictID:  dummyAnnictID,
						AniListID: dummyAniListID,
					},
				},
			})

		assert.Len(t, actual.AniListUpdates, 0)
		assert.Len(t, actual.Untethered, 0)
	})

	t.Run("差分が存在するためエントリーを更新する (Status)", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID:          dummyAnnictID,
					Title:             "お隣の天使様にいつの間にか駄目人間にされていた件",
					ViewerStatusState: status.AnnictWatched,
					NoEpisodes:        false,
					Episodes:          createEpisodeConnection(12),
				},
			},
			[]anilist.LibraryEntry{
				{
					Status:   status.AniListCurrent,
					Progress: 12,
					Media: anilist.Media{
						ID: dummyAniListID,
						Title: anilist.Title{
							Native: "お隣の天使様にいつの間にか駄目人間にされていた件",
						},
					},
				},
			},
			&arm.ArmDatabase{
				Entries: []arm.ArmEntry{
					{
						AnnictID:  dummyAnnictID,
						AniListID: dummyAniListID,
					},
				},
			})

		assert.Len(t, actual.AniListUpdates, 1)
		assert.Equal(t, dummyAniListID, actual.AniListUpdates[0].MediaID)
		assert.Equal(t, status.AniListCompleted, actual.AniListUpdates[0].Status)
		assert.Equal(t, 12, actual.AniListUpdates[0].Progress)
		assert.Len(t, actual.Untethered, 0)
	})

	t.Run("差分が存在するためエントリーを更新する (Progress)", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID:          dummyAnnictID,
					Title:             "ダークギャザリング",
					ViewerStatusState: status.AnnictWatching,
					NoEpisodes:        false,
					Episodes:          createEpisodeConnection(3),
				},
			},
			[]anilist.LibraryEntry{
				{
					Status:   status.AniListCurrent,
					Progress: 2,
					Media: anilist.Media{
						ID: dummyAniListID,
						Title: anilist.Title{
							Native: "ダークギャザリング",
						},
					},
				},
			},
			&arm.ArmDatabase{
				Entries: []arm.ArmEntry{
					{
						AnnictID:  dummyAnnictID,
						AniListID: dummyAniListID,
					},
				},
			})

		assert.Len(t, actual.AniListUpdates, 1)
		assert.Equal(t, dummyAniListID, actual.AniListUpdates[0].MediaID)
		assert.Equal(t, status.AniListCurrent, actual.AniListUpdates[0].Status)
		assert.Equal(t, 3, actual.AniListUpdates[0].Progress)
		assert.Len(t, actual.Untethered, 0)
	})

	t.Run("AniList に視聴記録がないためエントリーを作成する", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{
				{
					AnnictID:          dummyAnnictID,
					Title:             "現実主義勇者の王国再建記 第二部",
					ViewerStatusState: status.AnnictWatching,
					NoEpisodes:        false,
					Episodes:          createEpisodeConnection(10),
				},
			},
			[]anilist.LibraryEntry{},
			&arm.ArmDatabase{
				Entries: []arm.ArmEntry{
					{
						AnnictID:  dummyAnnictID,
						AniListID: dummyAniListID,
					},
				},
			})

		assert.Len(t, actual.AniListUpdates, 1)
		assert.Equal(t, dummyAniListID, actual.AniListUpdates[0].MediaID)
		assert.Equal(t, status.AniListCurrent, actual.AniListUpdates[0].Status)
		assert.Equal(t, 10, actual.AniListUpdates[0].Progress)
		assert.Len(t, actual.Untethered, 0)
	})

	t.Run("AniList ID から Annict ID を参照できない", func(t *testing.T) {
		actual := CalculateDiff(
			[]annict.Work{},
			[]anilist.LibraryEntry{
				{
					Media: anilist.Media{
						ID: dummyAniListID,
						Title: anilist.Title{
							Native: "江戸前エルフ",
						},
					},
				},
			},
			&arm.ArmDatabase{},
		)

		assert.Len(t, actual.AniListUpdates, 0)
		assert.Len(t, actual.Untethered, 1)
		assert.Equal(t, "AniList", actual.Untethered[0].Source)
		assert.Equal(t, dummyAniListID, actual.Untethered[0].ID)
		assert.Equal(t, "江戸前エルフ", actual.Untethered[0].Title)
	})
}

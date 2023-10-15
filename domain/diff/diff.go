package diff

import (
	"log/slog"

	"github.com/samber/lo"

	"github.com/SlashNephy/annict2anilist/domain/status"
	"github.com/SlashNephy/annict2anilist/external/anilist"
	"github.com/SlashNephy/annict2anilist/external/annict"
	"github.com/SlashNephy/annict2anilist/external/arm"
)

type Diff struct {
	AniListUpdates []*anilist.MediaListEntryUpdate
	Untethered     []*UntetheredEntry
}

type UntetheredEntry struct {
	Source string `json:"source"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
}

func CalculateDiff(works []annict.Work, entries []anilist.LibraryEntry, armDatabase *arm.ArmDatabase) Diff {
	var diff Diff
	for _, work := range works {
		// arm を参照して作品 ID を相互変換する
		arm, found := armDatabase.FindForAniList(work.AnnictID, work.MALAnimeID, work.SyobocalTID)

		// Annict ID から AniList ID を参照できない
		if !found || arm.AniListID == 0 {
			slog.Debug("arm does not have AniList relation",
				slog.Int("annict_id", work.AnnictID),
				slog.String("annict_title", work.Title),
			)

			// 紐付けられなかったものを記録
			diff.Untethered = append(diff.Untethered, &UntetheredEntry{
				Source: "Annict",
				ID:     work.AnnictID,
				Title:  work.Title,
			})
			continue
		}

		// Annict の視聴記録と一致する AniList の視聴記録を探す
		entry, found := lo.Find(entries, func(x anilist.LibraryEntry) bool {
			return x.Media.ID == arm.AniListID
		})
		annictProgress := detectAnnictProgress(work)

		// 差分が存在せず、更新の必要はない
		if found && status.IsSameListStatus(work.ViewerStatusState, entry.Status) && entry.Progress == annictProgress {
			continue
		}

		// 作品が終了していて、どちらのステータスも Completed になっている場合は Progress の更新を行わない
		// AniList は Completed にした作品の Progress を自動的に更新する
		// Annict と AniList ではエピソードの追加基準が異なる (例えば特番を Annict に含めることがあるが、AniList はそのようなエピソードを認めていないためずれが起こることがある)
		if found && entry.Media.Status == anilist.MediaStatusFinished && work.ViewerStatusState == status.AnnictWatched && entry.Status == status.AniListCompleted {
			slog.Debug("already completed",
				slog.String("annict_title", work.Title),
				slog.String("annict_state", string(work.ViewerStatusState)),
				slog.Int("annict_progress", annictProgress),
				slog.Int("annict_id", work.AnnictID),
				slog.String("anilist_title", entry.Media.Title.Native),
				slog.String("anilist_state", string(entry.Status)),
				slog.Int("anilist_progress", entry.Progress),
				slog.Int("anilist_id", entry.Media.ID),
			)
			continue
		}

		// 差分が存在するためエントリーを更新する
		if found {
			slog.Info(
				"Annict -> AniList",
				slog.String("media_status", string(entry.Media.Status)),
				slog.String("annict_title", work.Title),
				slog.String("annict_state", string(work.ViewerStatusState)),
				slog.Int("annict_progress", annictProgress),
				slog.Int("annict_id", work.AnnictID),
				slog.String("anilist_title", entry.Media.Title.Native),
				slog.String("anilist_state", string(entry.Status)),
				slog.Int("anilist_progress", entry.Progress),
				slog.Int("anilist_id", entry.Media.ID),
			)
		} else {
			// AniList に視聴記録がないためエントリーを作成する
			slog.Info(
				"Annict -> nil",
				slog.String("annict_title", work.Title),
				slog.String("annict_state", string(work.ViewerStatusState)),
				slog.Int("annict_progress", annictProgress),
				slog.Int("annict_id", work.AnnictID),
			)
		}

		// AniList にエントリーを作成 or 更新する
		diff.AniListUpdates = append(diff.AniListUpdates, &anilist.MediaListEntryUpdate{
			MediaID:  arm.AniListID,
			Status:   work.ViewerStatusState.ToAniListStatus(),
			Progress: annictProgress,
		})
	}

	for _, entry := range entries {
		// arm を参照して作品 ID を相互変換する
		arm, found := armDatabase.FindForAnnict(entry.Media.ID, entry.Media.IDMal)

		// AniList ID から Annict ID を参照できない
		if !found || arm.AnnictID == 0 {
			slog.Debug("arm does not have Annict relation",
				slog.Int("anilist_id", entry.Media.ID),
				slog.String("anilist_title", entry.Media.Title.Native),
			)

			diff.Untethered = append(diff.Untethered, &UntetheredEntry{
				Source: "AniList",
				ID:     entry.Media.ID,
				Title:  entry.Media.Title.Native,
			})
			continue
		}

		// AniList の視聴記録と一致する Annict の視聴記録を探す
		_, found = lo.Find(works, func(x annict.Work) bool {
			return x.AnnictID == arm.AnnictID
		})

		if !found {
			// AniList のみに含まれている
			slog.Info(
				"nil -> AniList",
				slog.String("anilist_title", entry.Media.Title.Native),
				slog.String("anilist_state", string(entry.Status)),
				slog.Int("anilist_progress", entry.Progress),
				slog.Int("anilist_id", entry.Media.ID),
			)

			// ひとまず AniList 側から削除することはない
		}
	}

	return diff
}

func detectAnnictProgress(work annict.Work) int {
	// 劇場版などエピソード区分がないものは視聴済みのエピソード数を 1 とする
	if work.NoEpisodes {
		if work.ViewerStatusState == status.AnnictWatched {
			return 1
		}

		return 0
	}

	// 記録済みのエピソード数を数える
	return lo.CountBy(work.Episodes.Nodes, func(episode annict.Episode) bool {
		return episode.ViewerDidTrack
	})
}

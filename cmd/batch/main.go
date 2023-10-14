package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/goccy/go-json"
	"github.com/samber/lo"

	"github.com/SlashNephy/annict2anilist/config"
	"github.com/SlashNephy/annict2anilist/domain/status"
	"github.com/SlashNephy/annict2anilist/external/anilist"
	"github.com/SlashNephy/annict2anilist/external/annict"
	"github.com/SlashNephy/annict2anilist/external/arm"
	_ "github.com/SlashNephy/annict2anilist/logger"
)

func main() {
	ctx := context.Background()

	if err := do(ctx); err != nil {
		panic(err)
	}
}

func do(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	annict, err := annict.NewClient(ctx, cfg)
	if err != nil {
		return err
	}

	annictViewer, err := annict.FetchViewer(ctx)
	if err != nil {
		return err
	}
	slog.Info("Annict user", slog.String("name", annictViewer.Viewer.Name), slog.String("username", annictViewer.Viewer.Username))

	aniList, err := anilist.NewClient(ctx, cfg)
	if err != nil {
		return err
	}

	aniListViewer, err := aniList.FetchViewer(ctx)
	if err != nil {
		return err
	}
	slog.Info("AniList user", slog.String("name", aniListViewer.Viewer.Name), slog.Int("id", aniListViewer.Viewer.ID))

	if cfg.DryRun {
		slog.Info("running in dry run mode")
	}

	return doLoop(ctx, cfg, annict, aniList, aniListViewer.Viewer.ID)
}

func doLoop(ctx context.Context, cfg *config.Config, annict *annict.Client, aniList *anilist.Client, aniListUserID int) error {
	arm, err := arm.FetchArmDatabase(ctx)
	if err != nil {
		return err
	}
	slog.Info("arm-supplementary entries", slog.Int("len", len(arm.Entries)))

	annictWorks, err := annict.FetchAllWorks(ctx)
	if err != nil {
		return err
	}
	slog.Info("Annict user works", slog.Int("len", len(annictWorks)))

	aniListEntries, err := aniList.FetchAllEntries(ctx, aniListUserID)
	if err != nil {
		return err
	}
	slog.Info("AniList user entries", slog.Int("len", len(aniListEntries)))

	return ExecuteUpdate(ctx, annictWorks, aniListEntries, arm, aniList, cfg)
}

type UntetheredEntry struct {
	Source string `json:"source"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
}

func ExecuteUpdate(ctx context.Context, works []annict.Work, entries []anilist.LibraryEntry, arm *arm.ArmDatabase, aniList *anilist.Client, cfg *config.Config) error {
	var untethered []UntetheredEntry
	for _, w := range works {
		a, found := arm.FindForAniList(w.AnnictID, w.MALAnimeID, w.SyobocalTID)
		if !found || a.AniListID == 0 {
			slog.Debug("arm does not have AniList relation", slog.Int("annict_id", w.AnnictID), slog.String("annict_title", w.Title))

			untethered = append(untethered, UntetheredEntry{
				Source: "Annict",
				ID:     w.AnnictID,
				Title:  w.Title,
			})
			continue
		}

		e, found := lo.Find(entries, func(x anilist.LibraryEntry) bool { return x.Media.ID == a.AniListID })

		var annictProgress int
		if w.NoEpisodes && w.ViewerStatusState == status.AnnictWatched {
			// 劇場版などエピソード区分がないものは視聴済みの本数を 1 とする
			annictProgress = 1
		} else {
			annictProgress = lo.CountBy(w.Episodes.Nodes, func(x annict.Episode) bool { return x.ViewerDidTrack })
		}

		if found {
			// 差分が存在
			if !status.IsSameListStatus(w.ViewerStatusState, e.Status) || e.Progress != annictProgress {
				// 作品が終了していて、どちらのステータスも Completed になっている場合は Progress の更新を行わない
				// AniList は Completed にした作品の Progress を自動的に更新する
				if e.Media.Status == anilist.MediaStatusFinished && w.ViewerStatusState == status.AnnictWatched && e.Status == status.AniListCompleted {
					slog.Debug("already completed",
						slog.String("annict_title", w.Title),
						slog.String("annict_state", string(w.ViewerStatusState)),
						slog.Int("annict_progress", annictProgress),
						slog.Int("annict_id", w.AnnictID),
						slog.String("anilist_title", e.Media.Title.Native),
						slog.String("anilist_state", string(e.Status)),
						slog.Int("anilist_progress", e.Progress),
						slog.Int("anilist_id", e.Media.ID),
					)
					continue
				}

				// Annict および AniList に含まれている
				slog.Info(
					"Annict -> AniList",
					slog.String("media_status", string(e.Media.Status)),
					slog.String("annict_title", w.Title),
					slog.String("annict_state", string(w.ViewerStatusState)),
					slog.Int("annict_progress", annictProgress),
					slog.Int("annict_id", w.AnnictID),
					slog.String("anilist_title", e.Media.Title.Native),
					slog.String("anilist_state", string(e.Status)),
					slog.Int("anilist_progress", e.Progress),
					slog.Int("anilist_id", e.Media.ID),
				)

				// AniList のエントリーを更新する
				if !cfg.DryRun {
					if err := aniList.SaveMediaListEntry(ctx, e.Media.ID, w.ViewerStatusState.ToAniListStatus(), annictProgress); err != nil {
						slog.Error("failed to save AniList entry", slog.Any("err", err))
						continue
					}
				}
			}
		} else {
			// Annict のみに含まれている
			slog.Info(
				"Annict -> nil",
				slog.String("annict_title", w.Title),
				slog.String("annict_state", string(w.ViewerStatusState)),
				slog.Int("annict_progress", annictProgress),
				slog.Int("annict_id", w.AnnictID),
			)

			// AniList にエントリーを作成する
			if !cfg.DryRun {
				if err := aniList.SaveMediaListEntry(ctx, a.AniListID, w.ViewerStatusState.ToAniListStatus(), annictProgress); err != nil {
					slog.Error("failed to save AniList entry", slog.Any("err", err))
					continue
				}
			}
		}
	}

	for _, e := range entries {
		a, found := arm.FindForAnnict(e.Media.ID, e.Media.IDMal)
		if !found || a.AnnictID == 0 {
			slog.Debug("arm does not have Annict relation", slog.Int("anilist_id", e.Media.ID), slog.String("anilist_title", e.Media.Title.Native))

			untethered = append(untethered, UntetheredEntry{
				Source: "AniList",
				ID:     e.Media.ID,
				Title:  e.Media.Title.Native,
			})
			continue
		}

		if !slices.ContainsFunc(works, func(x annict.Work) bool { return x.AnnictID == a.AnnictID }) {
			// AniList のみに含まれている
			slog.Info(
				"nil -> AniList",
				slog.String("anilist_title", e.Media.Title.Native),
				slog.String("anilist_state", string(e.Status)),
				slog.Int("anilist_progress", e.Progress),
				slog.Int("anilist_id", e.Media.ID),
			)

			// ひとまず AniList 側から削除することはない
		}
	}

	content, err := json.MarshalIndent(untethered, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(cfg.TokenDirectory, "untethered.json")
	return os.WriteFile(path, content, 0666)
}

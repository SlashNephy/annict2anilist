package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	if err := do(ctx); err != nil {
		logger.Error("do", zap.Error(err))
	}
}

func do(ctx context.Context) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	annict, err := NewAnnictClient(ctx, cfg, "token-annict.json")
	if err != nil {
		return err
	}

	annictViewer, err := annict.FetchViewer(ctx)
	if err != nil {
		return err
	}
	logger.Info("Annict user", zap.String("name", annictViewer.Viewer.Name), zap.String("username", annictViewer.Viewer.Username))

	aniList, err := NewAniListClient(ctx, cfg, "token-anilist.json")
	if err != nil {
		return err
	}

	aniListViewer, err := aniList.FetchViewer(ctx)
	if err != nil {
		return err
	}
	logger.Info("AniList user", zap.String("name", aniListViewer.Viewer.Name), zap.Int("id", aniListViewer.Viewer.ID))

	if cfg.DryRun {
		logger.Info("running in dry run mode")
	}

	return doLoop(ctx, cfg, annict, aniList, aniListViewer.Viewer.ID)
}

func doLoop(ctx context.Context, cfg *Config, annict *AnnictClient, aniList *AniListClient, aniListUserID int) error {
	arm, err := FetchArmDatabase(ctx)
	if err != nil {
		return err
	}
	logger.Info("kawaiioverflow/arm entries", zap.Int("len", len(arm.Entries)))

	annictWorks, err := annict.FetchAllWorks(ctx)
	if err != nil {
		return err
	}
	logger.Info("Annict user works", zap.Int("len", len(annictWorks)))

	aniListEntries, err := aniList.FetchAllEntries(ctx, aniListUserID)
	if err != nil {
		return err
	}
	logger.Info("AniList user entries", zap.Int("len", len(aniListEntries)))

	if err = ExecuteUpdate(ctx, annictWorks, aniListEntries, arm, aniList, cfg); err != nil {
		return err
	}

	if cfg.IntervalMinutes == 0 {
		return nil
	}

	duration := time.Duration(cfg.IntervalMinutes) * time.Minute
	logger.Info("sleep", zap.String("duration", duration.String()))
	time.Sleep(duration)

	return doLoop(ctx, cfg, annict, aniList, aniListUserID)
}

type UntetheredEntry struct {
	Source string `json:"source"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
}

func ExecuteUpdate(ctx context.Context, works []AnnictWork, entries []AniListLibraryEntry, arm *ArmDatabase, aniList *AniListClient, cfg *Config) error {
	var untethered []UntetheredEntry
	for _, w := range works {
		a, found := arm.FindForAniList(w.AnnictID, w.MALAnimeID, w.SyobocalTID)
		if !found || a.AniListID == 0 {
			logger.Debug("arm does not have AniList relation", zap.Int("annict_id", w.AnnictID), zap.String("annict_title", w.Title))

			untethered = append(untethered, UntetheredEntry{
				Source: "Annict",
				ID:     w.AnnictID,
				Title:  w.Title,
			})
			continue
		}

		e, found := FindByKey(entries, func(x AniListLibraryEntry) bool { return x.Media.ID == a.AniListID })

		var annictProgress int
		if w.NoEpisodes && w.ViewerStatusState == AnnictWatchedStatus {
			// ?????????????????????????????????????????????????????????????????????????????? 1 ?????????
			annictProgress = 1
		} else {
			annictProgress = CountByKey(w.Episodes.Nodes, func(x AnnictEpisode) bool { return x.ViewerDidTrack })
		}

		if found {
			// ???????????????
			if !IsSameListStatus(w.ViewerStatusState, e.Status) || e.Progress != annictProgress {
				// Annict ????????? AniList ?????????????????????
				logger.Info(
					"Annict -> AniList",
					zap.String("annict_title", w.Title),
					zap.String("annict_state", string(w.ViewerStatusState)),
					zap.Int("annict_progress", annictProgress),
					zap.String("anilist_title", e.Media.Title.Native),
					zap.String("anilist_state", string(e.Status)),
					zap.Int("anilist_progress", e.Progress),
				)

				// AniList ?????????????????????????????????
				if !cfg.DryRun {
					if err := aniList.UpdateMediaStatus(ctx, e.ID, w.ViewerStatusState.ToAniListStatus(), annictProgress); err != nil {
						return err
					}
				}
			}
		} else {
			// Annict ???????????????????????????
			logger.Info(
				"Annict -> nil",
				zap.String("annict_title", w.Title),
				zap.String("annict_state", string(w.ViewerStatusState)),
				zap.Int("annict_progress", annictProgress),
			)

			// AniList ?????????????????????????????????
			if !cfg.DryRun {
				if err := aniList.CreateMediaStatus(ctx, a.AniListID, w.ViewerStatusState.ToAniListStatus(), annictProgress); err != nil {
					return err
				}
			}
		}
	}

	for _, e := range entries {
		a, found := arm.FindForAnnict(e.Media.ID, e.Media.IDMal)
		if !found || a.AnnictID == 0 {
			logger.Debug("arm does not have Annict relation", zap.Int("anilist_id", e.Media.ID), zap.String("anilist_title", e.Media.Title.Native))

			untethered = append(untethered, UntetheredEntry{
				Source: "AniList",
				ID:     e.Media.ID,
				Title:  e.Media.Title.Native,
			})
			continue
		}

		if !Contains(works, func(x AnnictWork) bool { return x.AnnictID == a.AnnictID }) {
			// AniList ???????????????????????????
			logger.Info(
				"nil -> AniList",
				zap.String("anilist_title", e.Media.Title.Native),
				zap.String("anilist_state", string(e.Status)),
				zap.Int("anilist_progress", e.Progress),
			)

			// ???????????? AniList ????????????????????????????????????
		}
	}

	content, err := json.MarshalIndent(untethered, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(cfg.TokenDirectory, "untethered.json")
	if err = os.WriteFile(path, content, 0666); err != nil {
		return err
	}

	return nil
}

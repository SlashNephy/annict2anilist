package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"

	"github.com/SlashNephy/annict2anilist/config"
	"github.com/SlashNephy/annict2anilist/domain/diff"
	"github.com/SlashNephy/annict2anilist/external/anilist"
	"github.com/SlashNephy/annict2anilist/external/annict"
	"github.com/SlashNephy/annict2anilist/external/arm"
	_ "github.com/SlashNephy/annict2anilist/logger"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.Any("err", err))
		panic(err)
	}

	annict, err := annict.NewClient(ctx, cfg)
	if err != nil {
		slog.Error("failed to create Annict client", slog.Any("err", err))
		panic(err)
	}

	annictViewer, err := annict.FetchViewer(ctx)
	if err != nil {
		slog.Error("failed to fetch Annict viewer", slog.Any("err", err))
		panic(err)
	}
	slog.Info("connected to Annict",
		slog.String("username", annictViewer.Viewer.Username),
		slog.String("nickname", annictViewer.Viewer.Name),
	)

	aniList, err := anilist.NewClient(ctx, cfg)
	if err != nil {
		slog.Error("failed to create AniList client", slog.Any("err", err))
		panic(err)
	}

	aniListViewer, err := aniList.FetchViewer(ctx)
	if err != nil {
		slog.Error("failed to fetch AniList viewer", slog.Any("err", err))
		panic(err)
	}
	slog.Info("connected to AniList",
		slog.String("nickname", aniListViewer.Viewer.Name),
		slog.Int("user_id", aniListViewer.Viewer.ID),
	)

	armDatabase, err := arm.FetchArmDatabase(ctx)
	if err != nil {
		slog.Error("failed to fetch arm-supplementary database", slog.Any("err", err))
		panic(err)
	}
	slog.Info("fetched arm-supplementary entries", slog.Int("length", len(armDatabase.Entries)))

	annictWorks, err := annict.FetchAllWorks(ctx)
	if err != nil {
		slog.Error("failed to fetch Annict works", slog.Any("err", err))
		panic(err)
	}
	slog.Info("fetched Annict user works", slog.Int("length", len(annictWorks)))

	aniListEntries, err := aniList.FetchAllEntries(ctx, aniListViewer.Viewer.ID)
	if err != nil {
		slog.Error("failed to fetch AniList entries", slog.Any("err", err))
		panic(err)
	}
	slog.Info("fetched AniList user entries", slog.Int("length", len(aniListEntries)))

	diff := diff.CalculateDiff(annictWorks, aniListEntries, armDatabase)
	if len(diff.AniListUpdates) > 0 {
		if cfg.DryRun {
			slog.Info("running in dry run mode")
		} else {
			if err = aniList.BatchSaveMediaListEntry(ctx, diff.AniListUpdates); err != nil {
				slog.Error("failed to save AniList entry", slog.Any("err", err))
				panic(err)
			}
		}
	}

	content, err := json.MarshalIndent(diff.Untethered, "", "  ")
	if err != nil {
		slog.Error("failed to marshal untethered.json", slog.Any("err", err))
		panic(err)
	}

	path := filepath.Join(cfg.TokenDirectory, "untethered.json")
	if err = os.WriteFile(path, content, 0600); err != nil {
		slog.Error("failed to write untethered.json", slog.Any("err", err))
		panic(err)
	}

	return
}

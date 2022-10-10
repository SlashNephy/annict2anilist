package main

import (
	"context"
	"log"
	"time"

	"github.com/hasura/go-graphql-client"
)

func main() {
	ctx := context.Background()

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	annict, err := NewAnnictClient(ctx, cfg, "token-annict.json")
	if err != nil {
		log.Fatal(err)
	}

	aniList, err := NewAniListClient(ctx, cfg, "token-anilist.json")
	if err != nil {
		log.Fatal(err)
	}

	aniListViewer, err := FetchAniListViewer(ctx, aniList)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("AniList User = %s (%d)\n", aniListViewer.Viewer.Name, aniListViewer.Viewer.ID)

	if err = doLoop(ctx, cfg, annict, aniList, aniListViewer.Viewer.ID); err != nil {
		log.Fatal(err)
	}
}

func doLoop(ctx context.Context, cfg *Config, annict, aniList *graphql.Client, aniListUserID int) error {
	aniListEntries, err := FetchAllAniListEntries(ctx, aniList, aniListUserID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("AniList User has %d entries\n", len(aniListEntries))

	// TODO

	if cfg.IntervalMinutes == 0 {
		return nil
	}

	duration := time.Duration(cfg.IntervalMinutes) * time.Minute
	log.Printf("Sleep %v", duration)
	time.Sleep(duration)
	return doLoop(ctx, cfg, annict, aniList, aniListUserID)
}

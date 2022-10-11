package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
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

	annictViewer, err := annict.FetchViewer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Annict User = %s (@%s)\n", annictViewer.Viewer.Name, annictViewer.Viewer.Username)

	aniList, err := NewAniListClient(ctx, cfg, "token-anilist.json")
	if err != nil {
		log.Fatal(err)
	}

	aniListViewer, err := aniList.FetchViewer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("AniList User = %s (%d)\n", aniListViewer.Viewer.Name, aniListViewer.Viewer.ID)

	if cfg.DryRun {
		log.Printf("Running in dry run mode\n")
	}

	if err = doLoop(ctx, cfg, annict, aniList, aniListViewer.Viewer.ID); err != nil {
		log.Fatal(err)
	}
}

func doLoop(ctx context.Context, cfg *Config, annict *AnnictClient, aniList *AniListClient, aniListUserID int) error {
	arm, err := FetchArmDatabase(ctx)
	if err != nil {
		return err
	}
	log.Printf("kawaiioverflow/arm has %d entries\n", len(arm.Entries))

	annictWorks, err := annict.FetchAllWorks(ctx)
	if err != nil {
		return err
	}
	log.Printf("Annict User has %d works\n", len(annictWorks))

	aniListEntries, err := aniList.FetchAllEntries(ctx, aniListUserID)
	if err != nil {
		return err
	}
	log.Printf("AniList User has %d entries\n", len(aniListEntries))

	if err = ExecuteUpdate(ctx, annictWorks, aniListEntries, arm, aniList, cfg.DryRun); err != nil {
		return err
	}

	if cfg.IntervalMinutes == 0 {
		return nil
	}

	duration := time.Duration(cfg.IntervalMinutes) * time.Minute
	log.Printf("Sleep %v", duration)
	time.Sleep(duration)

	return doLoop(ctx, cfg, annict, aniList, aniListUserID)
}

type UntetheredEntry struct {
	Source string `json:"source"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
}

func ExecuteUpdate(ctx context.Context, works []AnnictWork, entries []AniListLibraryEntry, arm *ArmDatabase, aniList *AniListClient, dryRun bool) error {
	var untethered []UntetheredEntry
	for _, w := range works {
		a, found := arm.FindForAniList(w.AnnictID, w.MALAnimeID, w.SyobocalTID)
		if !found || a.AniListID == 0 {
			// log.Printf("arm does not have AniList ID for Annict#%d[%s]. Skipping...", w.AnnictID, w.Title)
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
			// 劇場版などエピソード区分がないものは視聴済みの本数を 1 とする
			annictProgress = 1
		} else {
			annictProgress = CountByKey(w.Episodes.Nodes, func(x AnnictEpisode) bool { return x.ViewerDidTrack })
		}

		if found {
			// 差分が存在
			if !IsSameListStatus(w.ViewerStatusState, e.Status) || e.Progress != annictProgress {
				// Annict および AniList に含まれている
				log.Printf("Annict[%s, %s, %d] -> AniList[%s, %s, %d]\n", w.Title, w.ViewerStatusState, annictProgress, e.Media.Title.Native, e.Status, e.Progress)

				// AniList のエントリーを更新する
				if !dryRun {
					if err := aniList.UpdateMediaStatus(ctx, e.ID, w.ViewerStatusState.ToAniListStatus(), annictProgress); err != nil {
						return err
					}
				}
			}
		} else {
			// Annict のみに含まれている
			log.Printf("Annict[%s, %s, %d] -> nil\n", w.Title, w.ViewerStatusState, annictProgress)

			// AniList にエントリーを作成する
			if !dryRun {
				if err := aniList.CreateMediaStatus(ctx, a.AniListID, w.ViewerStatusState.ToAniListStatus(), annictProgress); err != nil {
					return err
				}
			}
		}
	}

	for _, e := range entries {
		a, found := arm.FindForAnnict(e.Media.ID, e.Media.IDMal)
		if !found || a.AnnictID == 0 {
			// log.Printf("arm does not have Annict ID for AniList#%d[%s]. Skipping...", e.Media.ID, e.Media.Title.Native)
			untethered = append(untethered, UntetheredEntry{
				Source: "AniList",
				ID:     e.Media.ID,
				Title:  e.Media.Title.Native,
			})
			continue
		}

		if !Contains(works, func(x AnnictWork) bool { return x.AnnictID == a.AnnictID }) {
			// AniList のみに含まれている
			log.Printf("nil -> AniList[%s, %s, %d]\n", e.Media.Title.Native, e.Status, e.Progress)

			// ひとまず AniList 側から削除することはない
		}
	}

	content, err := json.MarshalIndent(untethered, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile("untethered.json", content, 0666); err != nil {
		return err
	}

	return nil
}

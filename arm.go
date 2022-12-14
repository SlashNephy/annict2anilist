package main

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"golang.org/x/exp/slices"
)

type ArmDatabase struct {
	Entries []ArmEntry
}

type ArmEntry struct {
	MalID       int `json:"mal_id"`
	AniListID   int `json:"anilist_id"`
	AnnictID    int `json:"annict_id"`
	SyobocalTID int `json:"syobocal_tid"`
}

func FetchArmDatabase(ctx context.Context) (*ArmDatabase, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	request, err := http.NewRequestWithContext(ctx, "GET", "https://raw.githubusercontent.com/SlashNephy/arm-supplementary/master/dist/arm.json", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", UserAgent)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var entries []ArmEntry
	if err = json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	return &ArmDatabase{
		Entries: entries,
	}, nil
}

func (d *ArmDatabase) FindByAnnictID(id int) (*ArmEntry, bool) {
	if id == 0 {
		return nil, false
	}

	index := slices.IndexFunc(d.Entries, func(entry ArmEntry) bool {
		return entry.AnnictID != 0 && entry.AnnictID == id
	})
	if index < 0 {
		return nil, false
	}

	return &d.Entries[index], true
}

func (d *ArmDatabase) FindByAniListID(id int) (*ArmEntry, bool) {
	if id == 0 {
		return nil, false
	}

	index := slices.IndexFunc(d.Entries, func(entry ArmEntry) bool {
		return entry.AniListID != 0 && entry.AniListID == id
	})
	if index < 0 {
		return nil, false
	}

	return &d.Entries[index], true
}

func (d *ArmDatabase) FindByMalID(id int) (*ArmEntry, bool) {
	if id == 0 {
		return nil, false
	}

	index := slices.IndexFunc(d.Entries, func(entry ArmEntry) bool {
		return entry.MalID != 0 && entry.MalID == id
	})
	if index < 0 {
		return nil, false
	}

	return &d.Entries[index], true
}

func (d *ArmDatabase) FindBySyobocalTID(tid int) (*ArmEntry, bool) {
	if tid == 0 {
		return nil, false
	}

	index := slices.IndexFunc(d.Entries, func(entry ArmEntry) bool {
		return entry.SyobocalTID != 0 && entry.SyobocalTID == tid
	})
	if index < 0 {
		return nil, false
	}

	return &d.Entries[index], true
}

func (d *ArmDatabase) FindForAniList(annictID int, malID string, syobocalID int) (*ArmEntry, bool) {
	// 1. Annict ID ????????????
	arm, found := d.FindByAnnictID(annictID)
	if found {
		return arm, found
	}

	// 2. MAL ID ????????????
	if malID != "" {
		malIntID, err := strconv.Atoi(malID)
		if err == nil {
			arm, found = d.FindByMalID(malIntID)
			if found {
				return arm, found
			}
		}
	}

	// 3. ??????????????????????????? TID ????????????
	return d.FindBySyobocalTID(syobocalID)
}

func (d *ArmDatabase) FindForAnnict(aniListID, malID int) (*ArmEntry, bool) {
	// 1. AniList ID ????????????
	arm, found := d.FindByAniListID(aniListID)
	if found {
		return arm, found
	}

	// 2. MAL ID ????????????
	return d.FindByMalID(malID)
}

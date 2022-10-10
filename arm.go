package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

type ArmEntry struct {
	MalID       int `json:"mal_id"`
	AniListID   int `json:"anilist_id"`
	AnnictID    int `json:"annict_id"`
	SyoboCalTID int `json:"syobocal_tid"`
}

func FetchArmDatabase(ctx context.Context) ([]ArmEntry, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	request, err := http.NewRequestWithContext(ctx, "GET", "https://raw.githubusercontent.com/kawaiioverflow/arm/master/arm.json", nil)
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

	return entries, nil
}

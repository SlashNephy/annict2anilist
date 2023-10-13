package arm

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"

	"github.com/SlashNephy/annict2anilist/config"
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

	request.Header.Add("User-Agent", config.UserAgent)
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

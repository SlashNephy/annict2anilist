package arm

import (
	"context"
	"io"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
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

func FetchArmDatabase(ctx context.Context, client *http.Client) (*ArmDatabase, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", "https://raw.githubusercontent.com/SlashNephy/arm-supplementary/master/dist/arm.json", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var entries []ArmEntry
	if err = json.Unmarshal(body, &entries); err != nil {
		return nil, errors.WithStack(err)
	}

	return &ArmDatabase{
		Entries: entries,
	}, nil
}

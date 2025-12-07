package spotify

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/pp-develop/music-timer-api/model"
)

func GetOembed(playlistId string) (model.Oembed, error) {
	var response model.Oembed

	endopint := "https://open.spotify.com/oEmbed"
	req, _ := http.NewRequest("GET", endopint, nil)
	q := req.URL.Query()
	q.Add("url", "https://open.spotify.com/playlist/"+playlistId)
	req.URL.RawQuery = q.Encode()

	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		slog.Error("oembed request failed", slog.Any("error", err))
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("oembed request returned non-200 status", slog.Int("status", resp.StatusCode))
		return response, err
	}

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		slog.Error("failed to unmarshal oembed response", slog.Any("error", err))
		return response, err
	}

	return response, nil
}

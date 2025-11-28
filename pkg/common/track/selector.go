package track

import (
	"github.com/pp-develop/music-timer-api/model"
)

// MakeTracks selects tracks from the given list to match the specified total playtime.
// It returns success status and the selected tracks.
func MakeTracks(allTracks []model.Track, totalPlayTimeMs int) (bool, []model.Track) {
	var tracks []model.Track
	var totalDuration int

	// Add tracks until total duration exceeds the specified playtime
	for _, v := range allTracks {
		tracks = append(tracks, v)
		totalDuration += v.DurationMs
		if totalDuration > totalPlayTimeMs {
			break
		}
	}

	// Remove the last track that caused overflow
	if len(tracks) > 0 {
		tracks = tracks[:len(tracks)-1]
	}

	// Calculate remaining time
	totalDuration = 0
	var remainingTime int
	for _, v := range tracks {
		totalDuration += v.DurationMs
	}
	remainingTime = totalPlayTimeMs - totalDuration

	// If remaining time is within allowance and playlist is long enough, no need to fill the gap
	if remainingTime <= AllowanceMs && totalPlayTimeMs >= MinPlaylistDurationForAllowanceMs {
		return true, tracks
	}

	// Try to find a track to fill the gap
	var isTrackFound bool
	getTrack := GetTrackByDuration(allTracks, remainingTime)
	if len(getTrack) > 0 {
		isTrackFound = true
		tracks = append(tracks, getTrack...)
	}

	return isTrackFound, tracks
}

// GetTrackByDuration finds a track with the specified duration in milliseconds
func GetTrackByDuration(allTracks []model.Track, durationMs int) []model.Track {
	tracks := []model.Track{}
	for _, track := range allTracks {
		if track.DurationMs == durationMs {
			tracks = append(tracks, track)
			break
		}
	}
	return tracks
}

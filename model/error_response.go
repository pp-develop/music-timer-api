package model

// ErrorResponse はAPIエラーレスポンスの構造
type ErrorResponse struct {
	Code    string                 `json:"code"`              // エラーコード（定数）
	Details map[string]interface{} `json:"details,omitempty"` // 追加情報
}

// エラーコード定数
const (
	// リソース不足
	CodeNotEnoughTracks      = "NOT_ENOUGH_TRACKS"       // 指定された再生時間に対して十分なトラックが見つからない
	CodeNoFavoriteTracks     = "NO_FAVORITE_TRACKS"      // ユーザーのお気に入りトラックが存在しない
	
	// API制限
	CodeSpotifyRateLimit     = "SPOTIFY_RATE_LIMIT"      // Spotify APIのレート制限に到達
	CodePlaylistQuotaExceeded = "PLAYLIST_QUOTA_EXCEEDED" // Spotifyアカウントのプレイリスト作成上限に到達
	
	// 処理エラー
	CodePlaylistCreationFailed = "PLAYLIST_CREATION_FAILED" // Spotify上でプレイリストの作成に失敗
	
	// 既存エラー
	CodeTracksNotFound       = "TRACKS_NOT_FOUND"       // データベースにトラックが存在しない
	CodeInternalError        = "INTERNAL_ERROR"          // その他の内部エラー
)
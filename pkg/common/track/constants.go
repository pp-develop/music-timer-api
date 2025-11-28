package track

const (
	// タイムアウト設定 15秒
	DefaultTimeoutSeconds = 15

	// 時間変換定数
	// 1分 = 60000ms
	MillisecondsPerMinute = 60000
	// 1秒 = 1000ms
	MillisecondsPerSecond = 1000

	// プレイリスト作成の許容誤差
	// 15秒 = 15000ms
	AllowanceMs = 15 * MillisecondsPerSecond

	// 許容誤差を適用する最小再生時間
	// 10分 = 600000ms
	MinPlaylistDurationForAllowanceMs = 10 * MillisecondsPerMinute
)

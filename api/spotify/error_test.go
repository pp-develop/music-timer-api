package spotify

import (
	"errors"
	"testing"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/zmb3/spotify/v2"
)

// =============================================================================
// IsAuthError 関数のテスト
// =============================================================================
// IsAuthError は、エラーが認証エラー（401 Unauthorized）かどうかを判定する関数。
// Spotify APIからのエラーレスポンスを解析し、トークンの期限切れなどを検出する。
// =============================================================================

// TestIsAuthError は、様々なエラータイプに対してIsAuthErrorが正しく判定するかをテストする。
//
// テストケース:
//   - nilエラー: falseを返す（エラーがないため認証エラーではない）
//   - 401 Spotifyエラー: trueを返す（トークン期限切れ）
//   - 429 Spotifyエラー: falseを返す（レート制限は認証エラーではない）
//   - 一般的なエラー: falseを返す（Spotifyエラーではない）
func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string // テストケースの説明
		err      error  // 入力エラー
		expected bool   // 期待される戻り値
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "401 Spotify error",
			err:      spotify.Error{Status: 401, Message: "token expired"},
			expected: true,
		},
		{
			name:     "429 Spotify error",
			err:      spotify.Error{Status: 429, Message: "rate limit"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("IsAuthError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// WrapSpotifyError 関数のテスト
// =============================================================================
// WrapSpotifyError は、Spotify APIエラーを適切なモデルエラーに変換する関数。
// HTTPステータスコードやエラーメッセージを解析し、アプリケーション固有の
// エラータイプにマッピングする。
//
// マッピングルール:
//   - 401 Unauthorized → model.ErrAccessTokenExpired
//   - 429 Too Many Requests → model.ErrSpotifyRateLimit
//   - "token expired" を含むメッセージ → model.ErrAccessTokenExpired
//   - "rate limit" を含むメッセージ → model.ErrSpotifyRateLimit
//   - "quota" を含むメッセージ → model.ErrPlaylistQuotaExceeded
//   - 不明なエラー + フォールバック指定 → フォールバックエラー
//   - 不明なエラー + フォールバックなし → 元のエラーをそのまま返す
// =============================================================================

// TestWrapSpotifyError は、様々なSpotifyエラーが正しくラップされるかをテストする。
func TestWrapSpotifyError(t *testing.T) {
	tests := []struct {
		name        string // テストケースの説明
		err         error  // 入力エラー
		fallback    error  // フォールバックエラー（オプション）
		expectedErr error  // 期待されるエラー
	}{
		{
			// nilエラーの場合はnilを返す
			name:        "nil error",
			err:         nil,
			fallback:    nil,
			expectedErr: nil,
		},
		{
			// 401エラーはトークン期限切れとして扱う
			// Spotify APIがトークンを拒否した場合に発生
			name:        "401 Spotify error",
			err:         spotify.Error{Status: 401, Message: "unauthorized"},
			fallback:    nil,
			expectedErr: model.ErrAccessTokenExpired,
		},
		{
			// 429エラーはレート制限として扱う
			// 短時間に多すぎるリクエストを送った場合に発生
			name:        "429 Spotify error",
			err:         spotify.Error{Status: 429, Message: "too many requests"},
			fallback:    nil,
			expectedErr: model.ErrSpotifyRateLimit,
		},
		{
			// エラーメッセージに "token expired" が含まれる場合
			// ステータスコードが取得できない場合のフォールバック判定
			name:        "token expired in message",
			err:         errors.New("token expired"),
			fallback:    nil,
			expectedErr: model.ErrAccessTokenExpired,
		},
		{
			// エラーメッセージに "rate limit" が含まれる場合
			name:        "rate limit in message",
			err:         errors.New("rate limit exceeded"),
			fallback:    nil,
			expectedErr: model.ErrSpotifyRateLimit,
		},
		{
			// エラーメッセージに "quota" が含まれる場合
			// ユーザーがプレイリスト作成上限に達した場合
			name:        "quota in message",
			err:         errors.New("playlist quota exceeded"),
			fallback:    nil,
			expectedErr: model.ErrPlaylistQuotaExceeded,
		},
		{
			// 不明なエラーでフォールバックが指定されている場合
			// 呼び出し元が代替エラーを指定できる
			name:        "unknown error with fallback",
			err:         errors.New("unknown error"),
			fallback:    model.ErrPlaylistCreationFailed,
			expectedErr: model.ErrPlaylistCreationFailed,
		},
		{
			// 不明なエラーでフォールバックがない場合
			// 元のエラーをそのまま返す
			name:        "unknown error without fallback",
			err:         errors.New("unknown error"),
			fallback:    nil,
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result error
			if tt.fallback != nil {
				result = WrapSpotifyError(tt.err, tt.fallback)
			} else {
				result = WrapSpotifyError(tt.err)
			}

			// nilチェック
			if tt.expectedErr == nil {
				if result != nil {
					t.Errorf("WrapSpotifyError() = %v, expected nil", result)
				}
				return
			}

			// モデルエラーの場合はerrors.Isで比較
			if errors.Is(result, tt.expectedErr) {
				return
			}

			// 不明なエラーの場合はメッセージで比較
			if result.Error() != tt.expectedErr.Error() {
				t.Errorf("WrapSpotifyError() = %v, expected %v", result, tt.expectedErr)
			}
		})
	}
}

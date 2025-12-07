package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// =============================================================================
// JWT テストの実行方法
// =============================================================================
// テストを実行する際は、以下の環境変数を設定する必要があります:
//
//   JWT_SECRET="test-secret-key-for-unit-testing-32chars"
//   ENCRYPTION_KEY="..."
//
// 実行コマンド例:
//   JWT_SECRET="..." ENCRYPTION_KEY="..." go test ./utils/...
//
// これらの環境変数は、jwtSecretパッケージ変数の初期化時に必要です。
// =============================================================================

// =============================================================================
// GenerateAccessToken 関数のテスト
// =============================================================================
// GenerateAccessToken は、短期間有効なアクセストークン（1時間）を生成する関数。
// ネイティブアプリ認証で使用され、APIリクエストの認証に使用される。
// =============================================================================

// TestGenerateAccessToken は、アクセストークンの生成が正しく動作することをテストする。
//
// テストケース:
//   - 有効な入力: トークンが正常に生成される
//   - サービスが空: エラーが返される（サービス識別が必須）
//
// アクセストークンの特徴:
//   - 有効期限: 1時間
//   - 用途: API呼び出しの認証
//   - 含まれる情報: ユーザーID、サービス名、トークンタイプ
func TestGenerateAccessToken(t *testing.T) {
	tests := []struct {
		name      string // テストケースの説明
		userID    string // ユーザーID
		service   string // サービス名（spotify, soundcloud等）
		expectErr bool   // エラーが期待されるか
	}{
		{
			// 正常系: 有効なパラメータでトークンを生成
			name:      "valid token generation",
			userID:    "user123",
			service:   "spotify",
			expectErr: false,
		},
		{
			// 異常系: サービスが空の場合はエラー
			// サービス識別は必須（どのサービスのユーザーか判別するため）
			name:      "empty service should fail",
			userID:    "user123",
			service:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateAccessToken(tt.userID, tt.service)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if token == "" {
				t.Error("Expected non-empty token")
			}
		})
	}
}

// =============================================================================
// GenerateRefreshToken 関数のテスト
// =============================================================================
// GenerateRefreshToken は、長期間有効なリフレッシュトークン（30日）を生成する関数。
// アクセストークンの再発行に使用され、JTI（JWT ID）でDB管理される。
// =============================================================================

// TestGenerateRefreshToken は、リフレッシュトークンの生成が正しく動作することをテストする。
//
// テストケース:
//   - 有効な入力: トークンが正常に生成される
//   - サービスが空: エラーが返される
//
// リフレッシュトークンの特徴:
//   - 有効期限: 30日
//   - 用途: アクセストークンの再発行
//   - JTI: 一意のID（DBで無効化を追跡するため）
func TestGenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		service   string
		jti       string // JWT ID（トークンの一意識別子）
		expectErr bool
	}{
		{
			// 正常系: 有効なパラメータでトークンを生成
			name:      "valid refresh token",
			userID:    "user123",
			service:   "spotify",
			jti:       "unique-jti-123",
			expectErr: false,
		},
		{
			// 異常系: サービスが空の場合はエラー
			name:      "empty service should fail",
			userID:    "user123",
			service:   "",
			jti:       "unique-jti-123",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshToken(tt.userID, tt.service, tt.jti)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if token == "" {
				t.Error("Expected non-empty token")
			}
		})
	}
}

// =============================================================================
// GenerateTokenPair 関数のテスト
// =============================================================================
// GenerateTokenPair は、アクセストークンとリフレッシュトークンのペアを生成する関数。
// ネイティブアプリのログイン成功時に使用される。
// =============================================================================

// TestGenerateTokenPair は、トークンペアの生成が正しく動作することをテストする。
//
// 検証項目:
//   - アクセストークンが空でない
//   - リフレッシュトークンが空でない
//   - ExpiresInが3600秒（1時間）
//   - TokenTypeがBearer
func TestGenerateTokenPair(t *testing.T) {
	tokenPair, err := GenerateTokenPair("user123", "spotify", "jti-123")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// アクセストークンの検証
	if tokenPair.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	// リフレッシュトークンの検証
	if tokenPair.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}

	// 有効期限の検証（1時間 = 3600秒）
	if tokenPair.ExpiresIn != 3600 {
		t.Errorf("Expected ExpiresIn=3600, got %d", tokenPair.ExpiresIn)
	}

	// トークンタイプの検証（OAuth 2.0標準のBearer）
	if tokenPair.TokenType != "Bearer" {
		t.Errorf("Expected TokenType=Bearer, got %s", tokenPair.TokenType)
	}
}

// =============================================================================
// ValidateJWT 関数のテスト
// =============================================================================
// ValidateJWT は、アクセストークンを検証し、ユーザーIDとサービスを返す関数。
// ミドルウェアで認証チェックに使用される。
// =============================================================================

// TestValidateJWT は、JWTの検証が正しく動作することをテストする。
//
// テストケース:
//   - 有効なトークン: ユーザーIDとサービスが正しく返される
//   - 無効なフォーマット: エラーが返される
//   - 空のトークン: エラーが返される
//
// 検証項目:
//   - 署名の検証（HS256）
//   - 有効期限のチェック
//   - トークンタイプのチェック（access）
//   - 必須フィールドの存在チェック
func TestValidateJWT(t *testing.T) {
	// テスト用の有効なトークンを生成
	validToken, _ := GenerateAccessToken("user123", "spotify")

	tests := []struct {
		name           string
		token          string
		expectedUserID string
		expectedSvc    string
		expectErr      bool
	}{
		{
			// 正常系: 有効なトークンを検証
			name:           "valid token",
			token:          validToken,
			expectedUserID: "user123",
			expectedSvc:    "spotify",
			expectErr:      false,
		},
		{
			// 異常系: 無効なトークンフォーマット
			// JWTは3つのドット区切りパートが必要
			name:      "invalid token format",
			token:     "invalid.token.here",
			expectErr: true,
		},
		{
			// 異常系: 空のトークン
			name:      "empty token",
			token:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, service, err := ValidateJWT(tt.token)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if userID != tt.expectedUserID {
				t.Errorf("UserID = %s, expected %s", userID, tt.expectedUserID)
			}
			if service != tt.expectedSvc {
				t.Errorf("Service = %s, expected %s", service, tt.expectedSvc)
			}
		})
	}
}

// TestValidateJWT_RefreshTokenShouldFail は、リフレッシュトークンを
// アクセストークンとして検証しようとした場合にエラーになることをテストする。
//
// セキュリティ上の理由:
//   - リフレッシュトークンはトークン再発行専用
//   - API認証にはアクセストークンを使用すべき
//   - トークンタイプの混同を防ぐ
func TestValidateJWT_RefreshTokenShouldFail(t *testing.T) {
	// リフレッシュトークンを生成
	refreshToken, _ := GenerateRefreshToken("user123", "spotify", "jti-123")

	// アクセストークン検証に渡すとエラーになるべき
	_, _, err := ValidateJWT(refreshToken)
	if err == nil {
		t.Error("Expected error when validating refresh token as access token")
	}
}

// =============================================================================
// ValidateRefreshToken 関数のテスト
// =============================================================================
// ValidateRefreshToken は、リフレッシュトークンを検証し、
// ユーザーID、サービス、JTIを返す関数。
// トークン再発行時に使用される。
// =============================================================================

// TestValidateRefreshToken は、リフレッシュトークンの検証が正しく動作することをテストする。
//
// 検証項目:
//   - 署名の検証
//   - トークンタイプのチェック（refresh）
//   - JTIの取得（DB照合用）
func TestValidateRefreshToken(t *testing.T) {
	// テスト用の有効なリフレッシュトークンを生成
	validToken, _ := GenerateRefreshToken("user123", "spotify", "jti-123")

	tests := []struct {
		name           string
		token          string
		expectedUserID string
		expectedSvc    string
		expectedJTI    string
		expectErr      bool
	}{
		{
			// 正常系: 有効なリフレッシュトークンを検証
			name:           "valid refresh token",
			token:          validToken,
			expectedUserID: "user123",
			expectedSvc:    "spotify",
			expectedJTI:    "jti-123",
			expectErr:      false,
		},
		{
			// 異常系: 無効なトークン
			name:      "invalid token",
			token:     "invalid.token",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, service, jti, err := ValidateRefreshToken(tt.token)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if userID != tt.expectedUserID {
				t.Errorf("UserID = %s, expected %s", userID, tt.expectedUserID)
			}
			if service != tt.expectedSvc {
				t.Errorf("Service = %s, expected %s", service, tt.expectedSvc)
			}
			if jti != tt.expectedJTI {
				t.Errorf("JTI = %s, expected %s", jti, tt.expectedJTI)
			}
		})
	}
}

// TestValidateRefreshToken_AccessTokenShouldFail は、アクセストークンを
// リフレッシュトークンとして検証しようとした場合にエラーになることをテストする。
//
// セキュリティ上の理由:
//   - トークンタイプの混同を防ぐ
//   - アクセストークンでリフレッシュを許可しない
func TestValidateRefreshToken_AccessTokenShouldFail(t *testing.T) {
	// アクセストークンを生成
	accessToken, _ := GenerateAccessToken("user123", "spotify")

	// リフレッシュトークン検証に渡すとエラーになるべき
	_, _, _, err := ValidateRefreshToken(accessToken)
	if err == nil {
		t.Error("Expected error when validating access token as refresh token")
	}
}

// =============================================================================
// 期限切れトークンのテスト
// =============================================================================

// TestValidateJWT_ExpiredToken は、期限切れのトークンが正しく拒否されることをテストする。
//
// テストシナリオ:
//   - 1時間前に期限切れになったトークンを手動で作成
//   - ValidateJWTに渡す
//   - エラーが返されることを確認
//
// 注意: 5分のクロックスキュー（時刻誤差）は許容されるが、
// 1時間の期限切れは確実に拒否される。
func TestValidateJWT_ExpiredToken(t *testing.T) {
	// 期限切れトークンを手動で作成（通常のAPIでは作れないため）
	claims := &Claims{
		UserID:  "user123",
		Service: "spotify",
		Type:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1時間前に期限切れ
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // 2時間前に発行
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	// 期限切れトークンは拒否されるべき
	_, _, err := ValidateJWT(tokenString)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

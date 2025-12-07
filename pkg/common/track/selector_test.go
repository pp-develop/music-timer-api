package track

import (
	"testing"

	"github.com/pp-develop/music-timer-api/model"
)

// =============================================================================
// MakeTracks 関数のテスト
// =============================================================================
// MakeTracks は指定された総再生時間に合うようにトラックを選択する関数。
// 以下のロジックをテストする:
// 1. トラックの合計時間が指定時間と完全一致する場合 → 成功
// 2. 10分以上のプレイリストで許容誤差（15秒）内の場合 → 成功
// 3. 10分未満のプレイリストでは許容誤差なし → 完全一致のみ成功
// 4. ギャップを埋めるトラックが見つかった場合 → 成功
// =============================================================================

// TestMakeTracks_ExactMatch は、トラックの合計時間が指定時間と完全に一致するケースをテストする。
//
// テストシナリオ:
//   - 入力: 3分 + 2分 = 5分のトラック
//   - 要求: 5分（300000ms）
//   - 期待結果: 成功（success=true）、2曲が選択される
//
// このテストは、最も基本的な「ぴったり一致」のケースを検証する。
func TestMakeTracks_ExactMatch(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 180000}, // 3分
		{Uri: "track2", DurationMs: 120000}, // 2分
	}

	success, result := MakeTracks(tracks, 300000) // 5分

	if !success {
		t.Error("Expected success=true for exact match")
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(result))
	}

	totalDuration := 0
	for _, track := range result {
		totalDuration += track.DurationMs
	}
	if totalDuration != 300000 {
		t.Errorf("Expected total duration 300000ms, got %d", totalDuration)
	}
}

// TestMakeTracks_WithAllowance は、10分以上のプレイリストで許容誤差内のケースをテストする。
//
// テストシナリオ:
//   - 入力: 5分 + 4分55秒 = 9分55秒のトラック
//   - 要求: 10分（600000ms）
//   - 誤差: 5秒（許容誤差15秒以内）
//   - 期待結果: 成功（長いプレイリストでは15秒の誤差を許容する）
//
// ビジネスルール:
//   - 10分以上のプレイリストでは、ユーザー体験への影響が小さいため
//     最大15秒の誤差を許容する
func TestMakeTracks_WithAllowance(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 300000}, // 5分
		{Uri: "track2", DurationMs: 295000}, // 4分55秒
	}

	// 10分（600000ms）を要求、実際は9分55秒（595000ms）= 5秒の誤差
	success, result := MakeTracks(tracks, 600000)

	if !success {
		t.Error("Expected success=true within allowance (15s) for playlist >= 10min")
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(result))
	}
}

// TestMakeTracks_ShortPlaylistNoAllowance は、10分未満のプレイリストでは
// 許容誤差が適用されないことをテストする。
//
// テストシナリオ:
//   - 入力: 3分 + 1分50秒 = 4分50秒のトラック
//   - 要求: 5分（300000ms）
//   - 誤差: 10秒
//   - 期待結果: 失敗（短いプレイリストでは誤差を許容しない）
//
// ビジネスルール:
//   - 短いプレイリスト（10分未満）では、誤差の影響が大きいため
//     完全一致のみを許可する
func TestMakeTracks_ShortPlaylistNoAllowance(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 180000}, // 3分
		{Uri: "track2", DurationMs: 110000}, // 1分50秒
	}

	// 5分（300000ms）を要求、実際は4分50秒 = 10秒の誤差
	success, _ := MakeTracks(tracks, 300000)

	if success {
		t.Error("Expected success=false for short playlist without exact match")
	}
}

// TestMakeTracks_EmptyTracks は、空のトラックリストが渡された場合の
// エッジケースをテストする。
//
// テストシナリオ:
//   - 入力: 空のトラックリスト
//   - 期待結果: 失敗（success=false）、空の結果
//
// このテストは、入力のバリデーションを検証する。
func TestMakeTracks_EmptyTracks(t *testing.T) {
	tracks := []model.Track{}

	success, result := MakeTracks(tracks, 300000)

	if success {
		t.Error("Expected success=false for empty tracks")
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 tracks, got %d", len(result))
	}
}

// TestMakeTracks_FindGapFiller は、ギャップを埋めるトラックが見つかる
// ケースをテストする。
//
// テストシナリオ:
//   - 入力: track1(3分), track2(1分), track3(3分20秒)
//   - 要求: 4分（240000ms）
//   - 処理フロー:
//     1. track1(3分)を選択 → 残り1分
//     2. track2(1分)でギャップを埋める
//   - 期待結果: 成功、合計4分ちょうど
//
// このテストは、メインの選択後に残り時間を埋める
// ギャップフィラーロジックを検証する。
func TestMakeTracks_FindGapFiller(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 180000}, // 3分
		{Uri: "track2", DurationMs: 60000},  // 1分 (ギャップ埋め用)
		{Uri: "track3", DurationMs: 200000}, // 3分20秒
	}

	// 4分（240000ms）を要求
	success, result := MakeTracks(tracks, 240000)

	if !success {
		t.Error("Expected success=true when gap filler found")
	}

	totalDuration := 0
	for _, track := range result {
		totalDuration += track.DurationMs
	}
	if totalDuration != 240000 {
		t.Errorf("Expected total duration 240000ms, got %d", totalDuration)
	}
}

// =============================================================================
// GetTrackByDuration 関数のテスト
// =============================================================================
// GetTrackByDuration は指定された再生時間に最も近い曲を探す関数。
// 以下のロジックをテストする:
// 1. 完全一致するトラックがある場合 → そのトラックを返す
// 2. 許容誤差内のトラックがある場合 → 最も近いトラックを返す
// 3. 短いプレイリストでは完全一致のみ
// 4. 複数の候補がある場合、最も近いものを選択
// =============================================================================

// TestGetTrackByDuration_ExactMatch は、指定した時間と完全に一致する
// トラックが見つかるケースをテストする。
//
// テストシナリオ:
//   - 入力: 3分, 2分, 1分のトラック
//   - 検索: 1分（60000ms）
//   - プレイリスト長: 10分（許容誤差あり）
//   - 期待結果: 1分のトラックが返される
func TestGetTrackByDuration_ExactMatch(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 180000}, // 3分
		{Uri: "track2", DurationMs: 120000}, // 2分
		{Uri: "track3", DurationMs: 60000},  // 1分
	}

	result := GetTrackByDuration(tracks, 60000, 600000)

	if len(result) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(result))
	}
	if result[0].DurationMs != 60000 {
		t.Errorf("Expected track with 60000ms, got %d", result[0].DurationMs)
	}
}

// TestGetTrackByDuration_WithinAllowance は、許容誤差内のトラックが
// 選択されるケースをテストする。
//
// テストシナリオ:
//   - 入力: 3分, 1分5秒のトラック
//   - 検索: 1分（60000ms）
//   - 許容誤差: 15秒（10分以上のプレイリスト）
//   - 期待結果: 1分5秒のトラックが返される（5秒の誤差は許容範囲内）
func TestGetTrackByDuration_WithinAllowance(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 180000}, // 3分
		{Uri: "track2", DurationMs: 65000},  // 1分5秒（5秒長い）
	}

	result := GetTrackByDuration(tracks, 60000, 600000)

	if len(result) != 1 {
		t.Fatalf("Expected 1 track within allowance, got %d", len(result))
	}
	if result[0].DurationMs != 65000 {
		t.Errorf("Expected track with 65000ms, got %d", result[0].DurationMs)
	}
}

// TestGetTrackByDuration_NoMatchShortPlaylist は、10分未満のプレイリストで
// 完全一致がない場合に何も返されないことをテストする。
//
// テストシナリオ:
//   - 入力: 1分5秒のトラック（5秒の誤差）
//   - 検索: 1分（60000ms）
//   - プレイリスト長: 5分（許容誤差なし）
//   - 期待結果: 空の結果（短いプレイリストでは誤差を許容しない）
func TestGetTrackByDuration_NoMatchShortPlaylist(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 65000}, // 1分5秒（5秒長い）
	}

	result := GetTrackByDuration(tracks, 60000, 300000)

	if len(result) != 0 {
		t.Error("Expected no match for short playlist without exact match")
	}
}

// TestGetTrackByDuration_SelectsBestMatch は、複数の候補がある場合に
// 最も近いトラックが選択されることをテストする。
//
// テストシナリオ:
//   - 入力: 1分10秒, 1分3秒, 55秒のトラック
//   - 検索: 1分（60000ms）
//   - 各トラックの誤差: 10秒, 3秒, 5秒
//   - 期待結果: 1分3秒のトラック（誤差3秒が最小）
//
// このテストは、「最も近い」トラックを選ぶ最適化ロジックを検証する。
func TestGetTrackByDuration_SelectsBestMatch(t *testing.T) {
	tracks := []model.Track{
		{Uri: "track1", DurationMs: 70000}, // 1分10秒（10秒長い）
		{Uri: "track2", DurationMs: 63000}, // 1分3秒（3秒長い）← ベストマッチ
		{Uri: "track3", DurationMs: 55000}, // 55秒（5秒短い）
	}

	result := GetTrackByDuration(tracks, 60000, 600000)

	if len(result) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(result))
	}
	if result[0].DurationMs != 63000 {
		t.Errorf("Expected best match (63000ms), got %d", result[0].DurationMs)
	}
}

// =============================================================================
// abs 関数のテスト
// =============================================================================

// TestAbs は、整数の絶対値を返すヘルパー関数をテストする。
//
// テストケース:
//   - 正の数: そのまま返す
//   - 負の数: 符号を反転して返す
//   - ゼロ: ゼロを返す
func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},     // 正の数はそのまま
		{-5, 5},    // 負の数は符号反転
		{0, 0},     // ゼロはゼロ
		{-100, 100}, // 大きな負の数
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%d) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

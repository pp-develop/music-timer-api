package internal

import (
	"fmt"
)

const playlistId = ""

func CreatePlaylistBySpecifyTime(ms int) (bool, string) {
	// tokenを使用してUser取得

	// トラックを取得
	isGetTracks, tracks := GetTracks(ms)
	fmt.Println(tracks)
	if !isGetTracks {
		return false, ""
	}

	// プレイリスト作成
	isCreate, playlistId := CreatePlaylist("", ms, "")
	if !isCreate {
		return false, ""
	}
	fmt.Println(isCreate)

	// プレイリストにトラックを追加
	isAddItems := AddItemsPlaylist(playlistId, tracks, "")
	fmt.Println(isAddItems)
	if !isAddItems {
		// 作成したプレイリストを削除
		return false, playlistId
	}
	return isAddItems, playlistId
}

package api

import (
	"log"
)

func CreatePlaylistBySpecifyTime(ms int) (bool, string) {
	// トラックを取得
	isGetTracks, tracks := GetTracks(ms)
	log.Println(tracks)
	if !isGetTracks {
		return false, ""
	}

	// tokenを使用してUser取得
	// プレイリスト作成
	isCreate, playlist := CreatePlaylist("", ms, "")
	if !isCreate {
		return false, ""
	}
	log.Println(isCreate)

	// プレイリストにトラックを追加
	isAddItems := AddItemsPlaylist(playlist.ID, tracks, "")
	log.Println(isAddItems)
	if !isAddItems {
		// 作成したプレイリストを削除
		return false, playlist.ID
	}
	return isAddItems, playlist.ID
}

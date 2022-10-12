package internal

import (
	"log"
)

const playlistId = ""

func CreatePlaylistBySpecifyTime(ms int) (bool, string) {
	// トラックを取得
	isGetTracks, tracks := GetTracks(ms)
	log.Println(tracks)
	if !isGetTracks {
		return false, ""
	}

	// tokenを使用してUser取得
	// プレイリスト作成
	isCreate, playlistId := CreatePlaylist("", ms, "")
	if !isCreate {
		return false, ""
	}
	log.Println(isCreate)

	// プレイリストにトラックを追加
	isAddItems := AddItemsPlaylist(playlistId, tracks, "")
	log.Println(isAddItems)
	if !isAddItems {
		// 作成したプレイリストを削除
		return false, playlistId
	}
	return isAddItems, playlistId
}

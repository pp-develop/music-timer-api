package internal

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Tracks struct {
	URI         string `json:"uri"`
	DURATION_MS int    `json:"duration_ms"`
}

func MysqlConecct() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database_name := os.Getenv("DB_NAME")

	dbconf := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + database_name + "?charset=utf8mb4"
	db, err := sql.Open("mysql", dbconf)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func getAllTracks(db *sql.DB) ([]Tracks, error) {
	var tracks []Tracks
	rows, err := db.Query("SELECT uri, duration_ms FROM tracks WHERE isrc like 'JP%' ORDER BY rand()")
	if err != nil {
		return tracks, err
	}
	defer rows.Close()

	for rows.Next() {
		var track Tracks
		if err := rows.Scan(&track.URI, &track.DURATION_MS); err != nil {
			return tracks, err
		}
		tracks = append(tracks, track)
	}
	if err = rows.Err(); err != nil {
		return tracks, err
	}
	return tracks, nil
}

func getTrackBySpecifyTime(db *sql.DB, ms int) []Tracks {
	var tracks []Tracks
	var track Tracks

	if err := db.QueryRow("SELECT uri, duration_ms FROM tracks WHERE duration_ms BETWEEN ?-30000 AND ?+30000 AND isrc LIKE 'JP%' ORDER BY rand()", ms, ms).Scan(&track.URI, &track.DURATION_MS); err != nil {
		return tracks
	}

	tracks = append(tracks, track)
	return tracks
}

func getTracksBySpecifyTime(db *sql.DB, allTracks []Tracks, specify_ms int) (bool, []Tracks) {
	var tracks []Tracks
	var sum_ms int

	// tracksの合計分数が指定された分数を超過したらループを停止
	for _, v := range allTracks {
		tracks = append(tracks, v)
		sum_ms += v.DURATION_MS
		if sum_ms > specify_ms {
			break
		}
	}

	// tracksから要素を1つ削除
	tracks = tracks[:len(tracks)-1]

	// 指定分数とtracksの合計分数の差分を求める
	sum_ms = 0
	var diff_ms int
	for _, v := range tracks {
		sum_ms += v.DURATION_MS
	}
	diff_ms = specify_ms - sum_ms

	// 誤差が30秒以内は許容
	if diff_ms < 30000 {
		return true, tracks
	}

	// 差分を埋めるtrackを取得
	var isGetTrack bool
	getTrack := getTrackBySpecifyTime(db, diff_ms)
	if len(getTrack) > 0 {
		isGetTrack = true
		tracks = append(tracks, getTrack...)
	}
	return isGetTrack, tracks
}

func GetTracks(specify_ms int) (bool, []Tracks) {
	db := MysqlConecct()
	var tracks []Tracks

	c1 := make(chan []Tracks, 1)
	go func() {
		var isGetTracks bool
		for !isGetTracks {
			allTracks, _ := getAllTracks(db)
			isGetTracks, tracks = getTracksBySpecifyTime(db, allTracks, specify_ms)
		}
		c1 <- tracks
	}()

	select {
	case <-c1:
		return true, tracks
	case <-time.After(30 * time.Second):
		return false, tracks
	}
}


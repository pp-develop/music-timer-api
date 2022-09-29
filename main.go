package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

const ONEMINUTE_TO_MSEC = 60000

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/playlist", getPlaylist)
	router.GET("/user", getUserProfile)

	router.Run(":8080")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func getPlaylist(c *gin.Context) {
	minute, _ := strconv.Atoi(c.Query("minute"))
	playlist := pkg.GetPlaylist(minute * ONEMINUTE_TO_MSEC)
	c.IndentedJSON(http.StatusOK, playlist)
}

func getUserProfile(c *gin.Context) {
	pkg.GetUserProfile()
}
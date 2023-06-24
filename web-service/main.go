package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	db "example.com/lords/stats/db"
	"github.com/gin-gonic/gin"
)

const (
	timeLayout = "2006-01-02"
)

func main() {
	file := flag.String("db", "test.db", "location of SQLite DB file")
	flag.Parse()

	statsDB, err := db.NewStats(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer statsDB.Close()

	router := gin.Default()
	router.StaticFile("/", "./static/index.html")
	router.Static("/static/", "./static/")
	router.Static("/bot/", "./bot/")
	router.GET("/api/:account/stats", getStatsGroupByUsersDate(statsDB))
	router.GET("/api/:account/players", getPlayers(statsDB))
	router.Run("localhost:8080")
}

func getStatsGroupByUsersDate(statsDB *db.StatsDB) func(*gin.Context) {
	return func(c *gin.Context) {
		account := c.Param("account")
		accountId, err := strconv.Atoi(account)
		if err != nil {
			c.IndentedJSON(http.StatusForbidden, "invalid account")
			return
		}

		userIDStr := c.DefaultQuery("user_id", "0")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid user_id filter")
			return
		}

		beginTimeStr, _isSetBeginTime := c.GetQuery("begin")
		if !_isSetBeginTime {
			beginTimeStr = time.Now().Add(time.Hour * -8).Format(timeLayout)
		}
		beginTime, err := time.Parse(timeLayout, beginTimeStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid begin time")
			return
		}

		endTimeStr, _isSetEndTime := c.GetQuery("end")
		if !_isSetEndTime {
			endTimeStr = time.Now().Add(time.Hour * -1).Format(timeLayout)
		}
		endTime, _ := time.Parse(timeLayout, endTimeStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid end time")
			return
		}

		data, err := statsDB.FindGroupByUserDate(accountId, beginTime, endTime, userID)
		if err != nil {
			log.Printf("Error: %e", err)
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, data)
	}
}

func getPlayers(statsDB *db.StatsDB) func(*gin.Context) {
	return func(c *gin.Context) {
		account := c.Param("account")
		accountId, err := strconv.Atoi(account)
		if err != nil {
			c.IndentedJSON(http.StatusForbidden, "invalid account")
			return
		}

		beginTimeStr, _isSetBeginTime := c.GetQuery("begin")
		if !_isSetBeginTime {
			beginTimeStr = time.Now().Add(time.Hour * -8).Format(timeLayout)
		}
		beginTime, err := time.Parse(timeLayout, beginTimeStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid begin time")
			return
		}

		endTimeStr, _isSetEndTime := c.GetQuery("end")
		if !_isSetEndTime {
			endTimeStr = time.Now().Add(time.Hour * -1).Format(timeLayout)
		}
		endTime, _ := time.Parse(timeLayout, endTimeStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid end time")
			return
		}

		data, err := statsDB.ListPlayers(accountId, beginTime, endTime)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, data)
	}
}

package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []struct {
		ShortDescription string `json:"shortDescription"`
		Price            string `json:"price"`
	} `json:"items"`
	Total string `json:"total"`
}

type processedReceipt struct {
	ID     string `json:"id"`
	Points int    `json:"points"`
}

type processedId struct {
	ID string `json:"id"`
}

type processedPoints struct {
	Points int `json:"points"`
}

var processedReceipts = []processedReceipt{}

func postReceipt(c *gin.Context) {
	var newPReceipt processedReceipt
	var newReceipt receipt
	var pId processedId

	if err := c.BindJSON(&newReceipt); err != nil {
		return
	}
	newPReceipt = processReceipt(newReceipt)

	processedReceipts = append(processedReceipts, newPReceipt)
	pId.ID = newPReceipt.ID
	c.IndentedJSON(http.StatusCreated, pId)
}

func getPointsById(c *gin.Context) {
	var pPoints processedPoints
	id := c.Param("id")

	for _, a := range processedReceipts {
		if a.ID == id {
			pPoints.Points = a.Points
			c.IndentedJSON(http.StatusOK, pPoints)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "processed receipt not found"})
}

func processReceipt(r receipt) processedReceipt {

	var points int
	var processedReceipt processedReceipt

	// 	One point for every alphanumeric character in the retailer name.
	for _, i := range r.Retailer {
		if unicode.IsLetter(i) || unicode.IsNumber(i) {
			points++
		}
	}

	// 50 points if the total is a round dollar amount with no cents.
	var t = strings.Split(r.Total, ".")
	i, err := strconv.Atoi(t[1])
	if i == 0 {
		points += 50
	} else {
		fmt.Println(err)
	}

	// 25 points if the total is a multiple of 0.25.
	d, err := strconv.ParseFloat(r.Total, 64)
	if (int(d)*100)%25 == 0 {
		points += 25
	} else {
		fmt.Println(err)
	}

	// 5 points for every two items on the receipt.
	var e = len(r.Items)
	points += 5 * (e / 2)

	// If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
	for _, item := range r.Items {
		var len = len(strings.Trim(item.ShortDescription, " "))
		if len%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			price = price * 0.2
			points += int(math.Round(price))

		}
	}

	// 6 points if the day in the purchase date is odd.
	var date = strings.Split(r.PurchaseDate, "-")
	dte, err := strconv.Atoi(date[2])
	if dte%2 == 1 {
		points += 6
	} else {
		fmt.Println(err)
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	var time = strings.Split(r.PurchaseTime, ":")
	tme, err := strconv.Atoi(time[0])

	if tme >= 14 && tme < 16 {
		points += 10
	} else {
		fmt.Println(err)
	}

	id := uuid.New()
	processedReceipt.Points = points
	processedReceipt.ID = id.String()

	return processedReceipt
}

func main() {
	router := gin.Default()
	router.GET("/receipts/:id/points", getPointsById)
	router.POST("/receipts/process", postReceipt)

	router.Run("localhost:8080")
}

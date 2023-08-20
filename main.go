package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []item `json:"items"`
}

type pointsResp struct {
	Points int `json:"points`
}

type idResp struct {
	Id int `json:"id"`
}

type errResp struct {
	Error string `json:"error"`
}

var receipts = map[int]receipt{}
var receiptId = 0

func calculatePoints(r receipt) int {
	var points = 0

	// One point for every alphanumeric character in the retailer name.
	for i := 0; i < len(r.Retailer); i++ {
		var c = r.Retailer[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			points++
		}
	}

	// 50 points if the total is a round dollar amount with no cents
	var s = strings.Split(r.Total, ".")
	if len(s) == 1 || s[1] == "00" {
		points += 50
	}

	// 25 points if the total is a multiple of 0.25
	if len(s) == 1 || s[1] == "00" || s[1] == "50" || s[1] == "75" {
		points += 25
	}

	// 5 points for every two items on the receipt
	points += (len(r.Items) / 2) * 5

	// If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
	for i := 0; i < len(r.Items); i++ {
		var item = r.Items[i]
		var desc = strings.Trim(item.ShortDescription, " ")
		if len(desc)%3 == 0 {
			itemPrice, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(float64(itemPrice) * 0.2))
		}
	}

	// 6 points if the day in the purchase date is odd
	date := strings.Split(r.PurchaseDate, "-")
	dd, _ := strconv.Atoi(date[2])
	if dd%2 == 1 {
		points += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm
	time := strings.Split(r.PurchaseTime, ":")
	hh, _ := strconv.Atoi(time[0])
	mm, _ := strconv.Atoi(time[1])
	if (hh > 14 && hh < 16) || (hh == 14 && mm > 0) {
		points += 10
	}

	return points
}

func getPoints(c *gin.Context) {
	receiptId := c.Param("id")
	id, _ := strconv.Atoi(receiptId)
	receipt, ok := receipts[id]

	if ok {
		var resp pointsResp
		resp.Points = calculatePoints(receipt)

		c.JSON(http.StatusOK, resp)
	} else {
		var resp errResp
		resp.Error = "Receipt ID does not exist"
		c.JSON(http.StatusBadRequest, resp)
	}

}

func uploadReceipt(c *gin.Context) {
	var newReceipt receipt
	if err := c.BindJSON(&newReceipt); err != nil {
		var resp errResp
		resp.Error = "Could not parse receipt"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	receipts[receiptId] = newReceipt
	var resp idResp
	resp.Id = receiptId
	receiptId++
	c.JSON(http.StatusCreated, resp)
}

func main() {
	router := gin.Default()
	router.GET("/receipts/:id/points", getPoints)
	router.POST("/receipts/process", uploadReceipt)
	router.Run("localhost:8080")
}

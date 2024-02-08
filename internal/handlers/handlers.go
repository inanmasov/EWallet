package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"infotecs/internal/database"

	"github.com/gin-gonic/gin"
)

func CreateWallet(c *gin.Context) {
	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	balance := 100.0
	walletID := db.CreateWalletDB(balance)

	c.JSON(http.StatusOK, gin.H{
		"id":      walletID,
		"balance": balance,
	})
}

func Send(c *gin.Context) {
	var requestData struct {
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	}

	// Читаем JSON-объект из тела запроса
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toWalletID := requestData.To
	amount := requestData.Amount

	// Здесь можно выполнить проверку наличия параметров to и amount
	if toWalletID == "" || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect parameters"})
		return
	}

	to, err := strconv.Atoi(toWalletID)
	if err != nil {
		fmt.Println("Ошибка при преобразовании строки в число:", err)
		return
	}

	walletID := c.Param("walletId")

	from, err := strconv.Atoi(walletID)
	if err != nil {
		fmt.Println("Ошибка при преобразовании строки в число:", err)
		return
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	status := db.SendWallet(from, to, amount)

	if status == -1 {
		return
	}

	c.JSON(status, nil)
}

func HistoryTransaction(c *gin.Context) {
	walletID := c.Param("walletId")

	id, err := strconv.Atoi(walletID)
	if err != nil {
		fmt.Println("Ошибка при преобразовании строки в число:", err)
		return
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	history, status := db.GetHistory(id)

	if status == -2 {
		c.JSON(http.StatusNotFound, nil)
	}

	if status == -1 {
		return
	}

	if status == 1 {
		c.JSON(http.StatusOK, "История транзакций пустая")
		return
	}

	// Распаковка массива байтов в структуру
	var historyJson []struct {
		Time         string  `json:"time"`
		FromWalletID int     `json:"from"`
		ToWalletID   int     `json:"to"`
		Amount       float64 `json:"amount"`
	}
	err = json.Unmarshal(history, &historyJson)
	if err != nil {
		fmt.Println("Ошибка при разборе JSON:", err)
		return
	}

	c.JSON(status, historyJson)
}

func GetWallet(c *gin.Context) {
	walletID := c.Param("walletId")

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	id, err := strconv.Atoi(walletID)
	if err != nil {
		fmt.Println("Ошибка при преобразовании строки в число:", err)
		return
	}

	balance, err := db.GetWalletCondition(id)

	if err != nil {
		c.JSON(http.StatusNotFound, nil)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"id":      walletID,
			"balance": balance,
		})
	}
}

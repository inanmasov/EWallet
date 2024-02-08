package service

import (
	"infotecs/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func StartService() {
	// Создаем новый экземпляр сервера Gin
	s := gin.Default()

	// Определяем маршруты для запросов на URL
	s.POST("/api/v1/wallet", handlers.CreateWallet)
	s.POST("/api/v1/wallet/:walletId/send", handlers.Send)
	s.GET("/api/v1/wallet/:walletId/history", handlers.HistoryTransaction)
	s.GET("/api/v1/wallet/:walletId", handlers.GetWallet)

	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// Запускаем сервер на порту
	s.Run(":" + viper.GetString("port"))
}

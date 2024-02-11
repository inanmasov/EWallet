package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/spf13/viper"
)

type Database struct {
	db *sql.DB
}

type transfer struct {
	Time         string  `json:"time"`
	FromWalletID int     `json:"from"`
	ToWalletID   int     `json:"to"`
	Amount       float64 `json:"amount"`
}

func Initialize() (*Database, error) {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	connect_db := "host=" + viper.GetString("db.host") + " " + "user=" + viper.GetString("db.username") + " " + "port=" + viper.GetString("db.port") + " " + "password=" + viper.GetString("db.password") + " " + "dbname=" + viper.GetString("db.dbname") + " " + "sslmode=" + viper.GetString("db.sslmode")
	db, err := sql.Open(viper.GetString("db.username"), connect_db)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &Database{db: db}, nil
}

func (db *Database) CreateWalletDB(balance float64) int {
	// Выполняем SQL запрос для вставки новой записи в таблицу Wallet
	var walletID int
	err := db.db.QueryRow("INSERT INTO Wallet (balance) VALUES ($1) RETURNING id", balance).Scan(&walletID)
	if err != nil {
		log.Fatalf("Ошибка вставки записи в таблицу Wallet: %v", err)
	}

	return walletID
}

func (db *Database) GetWalletCondition(walletID int) (float64, error) {
	// Выполняем SQL запрос для получения баланса по ID кошелька
	var balance float64
	err := db.db.QueryRow("SELECT balance FROM Wallet WHERE id = $1", walletID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Кошелек с указанным ID не найден
			return 0.0, err
		} else {
			log.Fatalf("Ошибка при запросе баланса: %v", err)
		}
	}

	return balance, err
}

func (db *Database) SendWallet(from int, to int, amount float64) int {
	balance, err := db.GetWalletCondition(from)
	if err != nil {
		return http.StatusNotFound
	}

	if balance < amount {
		return http.StatusBadRequest
	}

	_, err = db.GetWalletCondition(to)
	if err != nil {
		return http.StatusBadRequest
	}

	tx, err := db.db.Begin()
	if err != nil {
		log.Fatalf("Ошибка транзакции: %v", err)
		return -1
	}
	// Откатываем транзакцию в случае возникновения ошибки
	defer tx.Rollback()

	// Снимаем сумму с одного кошелька
	_, err = tx.Exec("UPDATE Wallet SET balance = balance - $1 WHERE id = $2", amount, from)
	if err != nil {
		log.Fatalf("Ошибка транзакции: %v", err)
		return -1
	}

	// Добавляем сумму на другой кошелек
	_, err = tx.Exec("UPDATE Wallet SET balance = balance + $1 WHERE id = $2", amount, to)
	if err != nil {
		log.Fatalf("Ошибка транзакции: %v", err)
		return -1
	}

	// Сохраняем в историю транзакций
	err = db.saveTransaction(from, to, amount)
	if err != nil {
		log.Fatalf("Ошибка сохранения транзакции: %v", err)
		return -1
	}

	// Если все SQL-запросы выполнены успешно, фиксируем транзакцию
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Ошибка транзакции: %v", err)
		return -1
	}

	return http.StatusOK
}

func (db *Database) saveTransaction(from, to int, amount float64) error {
	curTime := time.Now().Format(time.RFC3339)

	// Выполняем запрос на вставку данных о транзакции
	_, err := db.db.Exec("INSERT INTO History (id, time, from_id, to_id, amount) VALUES ($1, $2, $3, $4, $5)",
		from, curTime, from, to, amount)
	if err != nil {
		log.Fatalf("Ошибка при вставке данных в таблицу: %v", err)
		return err
	}

	// Выполняем запрос на вставку данных о транзакции
	_, err = db.db.Exec("INSERT INTO History (id, time, from_id, to_id, amount) VALUES ($1, $2, $3, $4, $5)",
		to, curTime, from, to, amount)
	if err != nil {
		log.Fatalf("Ошибка при вставке данных в таблицу: %v", err)
		return err
	}

	return err
}

func (db *Database) GetHistory(id int) ([]byte, int) {
	_, err := db.GetWalletCondition(id)
	if err != nil {
		return nil, -2
	}

	// Запрос истории транзакций
	rows, err := db.db.Query("SELECT time, from_id, to_id, amount FROM History WHERE id = $1", id)
	if err != nil {
		log.Fatalf("Ошибка запроса истории транзакций: %v", err)
		return nil, -1
	}
	defer rows.Close()

	// Создаем массив объектов Transfer
	var transfers []transfer

	// Итерируемся по результатам запроса и добавляем каждую строку в массив
	for rows.Next() {
		var transfer transfer
		var timestamp time.Time
		if err := rows.Scan(&timestamp, &transfer.FromWalletID, &transfer.ToWalletID, &transfer.Amount); err != nil {
			log.Fatalf("Ошибка запроса истории транзакций: %v", err)
			return nil, -1
		}
		transfer.Time = timestamp.Format(time.RFC3339)
		transfers = append(transfers, transfer)
	}

	// Проверяем наличие ошибок после завершения итерации по результатам запроса
	if err := rows.Err(); err != nil {
		log.Fatalf("Ошибка запроса истории транзакций: %v", err)
		return nil, -1
	}

	if len(transfers) == 0 {
		return nil, 1
	}

	// Преобразуем массив объектов Transfer в формат JSON
	jsonData, err := json.Marshal(transfers)
	if err != nil {
		log.Fatalf("Ошибка преобразования массива объектов в формат JSON: %v", err)
		return nil, -1
	}

	return jsonData, http.StatusOK
}

func (db *Database) Close() error {
	return db.db.Close()
}

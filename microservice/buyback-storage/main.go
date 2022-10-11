package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"jojonomic/microservice/buyback-storage/models"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Print("Error read environemt file", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), nil)
	if err != nil {
		log.Fatal("Error connect to database")
	}

	r := getKafkaReader(os.Getenv("KAFKA_URL"), os.Getenv("KAFKA_TOPIC"), os.Getenv("KAFKA_GROUP_ID"))
	ctx := context.Background()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			break
		}
		log.Printf("fetch message topic:%v/partition:%v/offset:%v key: %s\n", m.Topic, m.Partition, m.Offset, string(m.Key))

		if err := SaveBuyback(db, m.Value); err != nil {
			log.Println(err.Error())
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			log.Fatal("commit message error:", err)
		}
	}
}

func SaveBuyback(db *gorm.DB, data []byte) error {
	var buybackParams models.BuybackParams
	if err := json.Unmarshal(data, &buybackParams); err != nil {
		return fmt.Errorf("unmarshall data error : %s", err.Error())
	}

	conn := db.Begin()

	goldBalance := buybackParams.CurrentGoldBalance - buybackParams.GoldWeight
	transaction := models.Transaksi{
		ReffID:       buybackParams.ReffID,
		Norek:        buybackParams.Norek,
		Type:         "buyback",
		GoldWeight:   buybackParams.GoldWeight,
		GoldBalance:  goldBalance,
		HargaTopup:   buybackParams.HargaTopup,
		HargaBuyback: buybackParams.HargaBuyback,
		CreatedAt:    int(time.Now().Unix()),
	}

	if err := conn.Create(&transaction).Error; err != nil {
		conn.Rollback()
		return err
	}

	if err := conn.Model(models.Rekening{}).Where("norek = ?", buybackParams.Norek).Update("gold_balance", goldBalance).Error; err != nil {
		conn.Rollback()
		return err
	}

	return conn.Commit().Error
}

func getKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	brokers := strings.Split(kafkaURL, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 1e3,
		MaxBytes: 10e6,
	})
}
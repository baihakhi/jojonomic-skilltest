package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"jojonomic/microservice/buyback-storage/models"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
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

	r := getKafkaReader(os.Getenv("KAFKA_URL"), os.Getenv("KAFKA_GROUP_ID"))
	log.Println("topic", os.Getenv("KAFKA_TOPIC"))
	if err := r.SubscribeTopics([]string{os.Getenv("KAFKA_TOPIC")}, nil); err != nil {
		log.Println("error subscribe", err)
	}
	for {

		msg, err := r.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			log.Printf("fetch message topic:%v/partition:%v/offset:%v key: %s\n", msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset, string(msg.Key))
			if err := saveData(db, msg.Value); err != nil {
				log.Println(msg.Value)
				log.Println(err.Error())
				continue
			}
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			break
		}
	}
}

func saveData(db *gorm.DB, data []byte) error {
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

func getKafkaReader(kafkaURL, groupID string) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaURL,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}
	return c
}

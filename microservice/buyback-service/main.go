package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"jojonomic/microservice/buyback-service/models"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/teris-io/shortid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error read environemt file", err)
	}

	kafkaConnection, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_URL"), "acks": "all"})
	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}
	defer kafkaConnection.Close()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), nil)
	if err != nil {
		log.Fatal("Error connect to database")
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/buyback", HandleBuyback(kafkaConnection, db)).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Println("server start at", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func HandleBuyback(kafkaConnection *kafka.Producer, db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req models.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: err.Error(),
			})
			return
		}

		if req.Norek == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: string(errors.New("no rekening can not be empty").Error()),
			})
			return
		}
		if req.GoldWeight == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: string(errors.New("gold weight can not be empty").Error()),
			})
			return
		}
		if req.Amount == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: string(errors.New("amount can not be empty").Error()),
			})
			return
		}

		rekening, err := getRekening(db, req.Norek)
		if err != nil {
			code := http.StatusInternalServerError
			if err == gorm.ErrRecordNotFound {
				code = http.StatusNotFound
			}

			w.WriteHeader(code)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: "rekening tidak ditemukan",
			})
			return
		}

		if rekening.GoldBalance < req.GoldWeight {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: "saldo emas anda tidak mencukupi",
			})
			return
		}

		harga, err := getHarga(db, req.Amount)
		if err != nil {
			code := http.StatusInternalServerError
			if err == gorm.ErrRecordNotFound {
				code = http.StatusNotFound
			}

			w.WriteHeader(code)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: "data tidak ditemukan",
			})
			return
		}

		reffID, err := shortid.Generate()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: err.Error(),
			})
			return
		}

		params := models.Params{
			GoldWeight:         req.GoldWeight,
			Amount:             req.Amount,
			Norek:              req.Norek,
			ReffID:             reffID,
			HargaTopup:         harga.HargaTopup,
			HargaBuyback:       harga.HargaBuyback,
			CurrentGoldBalance: rekening.GoldBalance,
		}

		payloadBytes, err := json.Marshal(&params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: err.Error(),
			})
			return
		}

		go func() {
			for e := range kafkaConnection.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
					} else {
						fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
							*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
					}
				}
			}
		}()

		topic := os.Getenv("KAFKA_TOPIC")

		msg := kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(fmt.Sprintf("address-%s", r.RemoteAddr)),
			Value:          payloadBytes,
		}
		err = kafkaConnection.Produce(&msg, nil)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				ReffID:  reffID,
				Message: "Kafka not ready",
			})
			return
		}

		json.NewEncoder(w).Encode(models.Response{
			Error:  false,
			ReffID: reffID,
		})
	}
}

func getRekening(db *gorm.DB, norek string) (*models.Rekening, error) {
	rekening := models.Rekening{}
	if err := db.Model(rekening).Where("norek = ?", norek).First(&rekening).Error; err != nil {
		return nil, err
	}

	return &rekening, nil
}

func getHarga(db *gorm.DB, Amount int64) (*models.Harga, error) {
	harga := models.Harga{}
	if err := db.Model(harga).Where("harga_buyback = ?", Amount).Order("created_at desc").First(&harga).Error; err != nil {
		return nil, err
	}

	return &harga, nil
}

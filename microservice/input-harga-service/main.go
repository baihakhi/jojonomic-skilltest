package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"jojonomic/microservice/input-harga-service/models"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/teris-io/shortid"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error read environemt file ", err)
	}

	kafkaConnection, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_URL"), "acks": "all"})
	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}
	defer kafkaConnection.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/input-harga", HandleInputHarga(kafkaConnection)).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Println("server start at", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func HandleInputHarga(kafkaConnection *kafka.Producer) func(w http.ResponseWriter, r *http.Request) {
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

		if req.AdminID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: string(errors.New("AdminID can not be empty").Error()),
			})
			return
		}
		if req.HargaTopup == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: string(errors.New("harga topup can not be empty").Error()),
			})
			return
		}
		if req.HargaBuyback == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: string(errors.New("harga buyback can not be empty").Error()),
			})
			return
		}

		reffId, err := shortid.Generate()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: err.Error(),
			})
			return
		}
		req.ReffID = reffId

		payloadBytes, err := json.Marshal(&req)
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
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				ReffId:  reffId,
				Message: "Kafka not ready",
			})
			return
		}

		json.NewEncoder(w).Encode(models.Response{
			Error:  false,
			ReffId: reffId,
		})
	}
}

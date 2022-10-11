package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"jojonomic/microservice/input-harga-service/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/teris-io/shortid"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error read environemt file ", err)
	}

	kafkaConnection := createKafkaConnection(os.Getenv("KAFKA_URL"), os.Getenv("KAFKA_TOPIC"))
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

func HandleInputHarga(kafkaConnection *kafka.Conn) func(w http.ResponseWriter, r *http.Request) {
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

		kafkaConnection.SetWriteDeadline(time.Now().Add(10 * time.Second))
		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("address-%s", r.RemoteAddr)),
			Value: payloadBytes,
		}
		_, err = kafkaConnection.WriteMessages(msg)
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

func createKafkaConnection(kafkaURL, topic string) *kafka.Conn {
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, topic, 0)
	if err != nil {
		log.Fatal(err.Error())
	}

	return conn
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"jojonomic/microservice/cek-harga-service/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error read environemt file", err)
	}

	db_dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(db_dsn), nil)
	if err != nil {
		log.Fatal("Error connect to database")
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/check-harga", HandlerGetHarga(db)).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Println("server start at", srv.Addr)
	log.Println(srv.ListenAndServe())
}

func HandlerGetHarga(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var harga models.Harga
		if err := db.Model(harga).Order("created_at desc").First(&harga).Error; err != nil {
			code := http.StatusInternalServerError
			if err == gorm.ErrRecordNotFound {
				code = http.StatusNotFound
			}

			w.WriteHeader(code)
			json.NewEncoder(w).Encode(models.Response{
				Error:   true,
				Message: err.Error(),
			})
		}

		json.NewEncoder(w).Encode(models.Response{
			Error: false,
			Data: struct {
				HargaTopup   int64 `json:"harga_topup"`
				HargaBuyback int64 `json:"harga_buyback"`
			}{
				HargaTopup:   harga.HargaTopup,
				HargaBuyback: harga.HargaBuyback,
			},
		})
	}
}

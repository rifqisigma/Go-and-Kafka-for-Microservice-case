package main

import (
	"log"
	"service_store/cmd/database"
	"service_store/entity"
)

func main() {

	db, _, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entity.Store{}); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Migrasi selesai dan database siap")
}

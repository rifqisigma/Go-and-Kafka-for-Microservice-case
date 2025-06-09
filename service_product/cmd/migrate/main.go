package main

import (
	"log"
	"service_product/cmd/database"
	"service_product/entity"
)

func main() {

	db, _, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entity.Product{}); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Migrasi selesai dan database siap")
}

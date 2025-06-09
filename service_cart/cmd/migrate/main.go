package main

import (
	"log"
	"service_cart/cmd/database"
	"service_cart/entity"
)

func main() {

	db, _, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entity.CartItem{}); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Migrasi selesai dan database siap")
}

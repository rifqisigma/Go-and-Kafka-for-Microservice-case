package main

import (
	"log"
	"service_user/cmd/database"
	"service_user/entity"
)

func main() {

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entity.User{}); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Migrasi selesai dan database siap")
}

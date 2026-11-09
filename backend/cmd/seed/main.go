package main

import (
	"log"

	"cloudprep-backend/internal/config"
	"cloudprep-backend/internal/db"
	"cloudprep-backend/internal/models"
)

func main() {
	cfg := config.Load()

	database, dbDriver, err := db.ConnectAndMigrate(cfg)
	if err != nil {
		log.Fatalf("[cmd/seed] database migration & seeding failed: %v", err)
	}

	var totalQuestions, totalOptions int64
	database.Model(&models.Question{}).Where("status = ?", models.QuestionStatusApproved).Count(&totalQuestions)
	database.Model(&models.QuestionOption{}).Count(&totalOptions)

	log.Printf("[cmd/seed] SUCCESS (driver=%s): %d approved questions and %d options verified in database!", dbDriver, totalQuestions, totalOptions)
}


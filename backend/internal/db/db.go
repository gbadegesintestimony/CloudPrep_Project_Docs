package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloudprep-backend/internal/config"
	"cloudprep-backend/internal/curriculum"
	"cloudprep-backend/internal/models"
	"cloudprep-backend/internal/seed"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// ConnectAndMigrate opens a GORM connection (Supabase PostgreSQL if DATABASE_URL starts with
// postgres:// or postgresql://, otherwise pure-Go local SQLite at cloudprep.db), runs AutoMigrate
// on all 12 domain models, and idempotently seeds the 10 CurriculumTopics and 31 LearningObjectives.
func ConnectAndMigrate(cfg *config.Config) (*gorm.DB, string, error) {
	var dialector gorm.Dialector
	var driverName string

	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
		dialector = postgres.Open(dbURL)
		driverName = "postgres"
	} else {
		sqlitePath := dbURL
		if sqlitePath == "" {
			sqlitePath = "cloudprep.db"
		}
		dialector = sqlite.Open(sqlitePath)
		driverName = "sqlite"
	}

	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true, // Prevents PII/email leaking into logs on normal 404 lookups
			ParameterizedQueries:      true, // Never print raw SQL parameter literals to logs
			Colorful:                  false,
		},
	)

	gormCfg := &gorm.Config{
		Logger: customLogger,
	}

	database, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, driverName, fmt.Errorf("failed to connect to %s database: %w", driverName, err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, driverName, fmt.Errorf("failed to acquire underlying sql.DB: %w", err)
	}

	if driverName == "postgres" {
		sqlDB.SetMaxOpenConns(15)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	} else {
		sqlDB.SetMaxOpenConns(1)
	}

	if err := database.AutoMigrate(
		&models.User{},
		&models.CurriculumTopic{},
		&models.LearningObjective{},
		&models.Question{},
		&models.QuestionOption{},
		&models.PracticeSession{},
		&models.SessionQuestion{},
		&models.Response{},
		&models.QuestionExposure{},
		&models.QuestionBookmark{},
		&models.GenerationRecord{},
		&models.AuditEvent{},
	); err != nil {
		return nil, driverName, fmt.Errorf("failed to auto-migrate schema: %w", err)
	}

	if driverName == "postgres" {
		_ = database.Exec("ALTER TABLE questions ALTER COLUMN scenario_type TYPE varchar(255);").Error
	}

	if err := SeedCurriculum(database); err != nil {
		return nil, driverName, fmt.Errorf("failed to seed curriculum topics and objectives: %w", err)
	}

	if err := seed.SeedQuestionBank(database); err != nil {
		return nil, driverName, fmt.Errorf("failed to seed 1,000+ question bank: %w", err)
	}

	return database, driverName, nil
}

// SeedCurriculum idempotently upserts all 10 CurriculumTopics and 31 LearningObjectives
// strictly grounded in the 8 course slide PDFs in school notes/.
func SeedCurriculum(database *gorm.DB) error {
	now := time.Now().UTC()
	for _, topicDef := range curriculum.GetCurriculumTopics() {
		topic := models.CurriculumTopic{
			ID:           topicDef.ID,
			Slug:         topicDef.Slug,
			ModuleNumber: topicDef.ModuleNumber,
			ModuleTitle:  topicDef.ModuleTitle,
			WeekCode:     topicDef.WeekCode,
			Title:        topicDef.Title,
			Description:  topicDef.Description,
			SourcePdf:    topicDef.SourcePdf,
			Ordering:     topicDef.Ordering,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := database.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"slug", "module_number", "module_title", "week_code", "title", "description", "source_pdf", "ordering", "updated_at"}),
		}).Create(&topic).Error; err != nil {
			return fmt.Errorf("failed upserting topic %s: %w", topic.ID, err)
		}

		for _, objDef := range topicDef.Objectives {
			obj := models.LearningObjective{
				ID:                objDef.ID,
				TopicID:           topic.ID,
				Code:              objDef.Code,
				Title:             objDef.Title,
				Statement:         objDef.Statement,
				CurriculumExcerpt: objDef.CurriculumExcerpt,
				SourceReference:   objDef.SourceReference,
				Ordering:          objDef.Ordering,
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			if err := database.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"topic_id", "code", "title", "statement", "curriculum_excerpt", "source_reference", "ordering", "updated_at"}),
			}).Create(&obj).Error; err != nil {
				return fmt.Errorf("failed upserting objective %s: %w", obj.ID, err)
			}
		}
	}

	var topicCount, objCount int64
	database.Model(&models.CurriculumTopic{}).Count(&topicCount)
	database.Model(&models.LearningObjective{}).Count(&objCount)
	log.Printf("[db] Curriculum verified: %d topics, %d learning objectives seeded", topicCount, objCount)
	return nil
}

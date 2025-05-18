package db

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vikasavnish/trademicro/internal/config"
	"github.com/vikasavnish/trademicro/internal/models"
)

// Connect establishes a connection to the database
func Connect(config config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.URL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create a default admin user if none exists
	createDefaultAdmin(db)

	return db, nil
}

// ConnectRedis establishes a connection to Redis
func ConnectRedis(config config.RedisConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test the connection
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// createDefaultAdmin creates a default admin user if no users exist
func createDefaultAdmin(db *gorm.DB) {
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Servloci@54321"), bcrypt.DefaultCost)
		db.Create(&models.User{
			Username:       "vikasavnish",
			HashedPassword: string(hashedPassword),
			Email:          "bizpowersolution@gmail.com",
			Role:           "admin",
		})
		log.Println("Created default admin user")
	}
}

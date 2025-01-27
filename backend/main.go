package main

import (
	"QResume/controllers"
	"QResume/models"
	"QResume/repo"
	"QResume/service"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

func main() {
	// Retry database connection
	var db *gorm.DB
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	// Retry loop for 30 seconds
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to the database, retrying... %v", err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to the database after multiple attempts: %v", err)
	}

	// Automigrate the User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Inject dependencies
	userRepo := repo.NewUserRepo(db)
	userService := service.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	// Initialize Gin router
	r := gin.Default()

	// CORS middleware setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},                   // Allow your frontend origin
		AllowMethods:     []string{"GET", "POST", "PATCH"},                   // Allowed HTTP methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // Allowed headers
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
	}))

	// Define routes
	r.POST("api/users/sign-on", userController.RegisterUser)
	r.PATCH("/api/users/details/:user-email", userController.UpdateDetails)
	r.GET("/api/users/details/:user-email", userController.GetUserDetails)
	r.GET("/api/users/my-qr/:user-email", userController.GetUserQRCode)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}

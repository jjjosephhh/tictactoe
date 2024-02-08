package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

func main() {
	e := echo.New()

// Initialize database connection
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect database")
		return
	}
	// Auto migrate the User model
	db.AutoMigrate(&User{})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})


	// Define the handler for POST request to create a new user
	e.POST("/users", func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			return err
		}
		// Validate username
		if !usernameRegex.MatchString(u.Username) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid username format"})
		}
		// Check if username already exists
		var existingUser User
		if db.Where("username = ?", u.Username).First(&existingUser).Error == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username already exists"})
		}
		// Encrypt password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
		// Create user record
		if err := db.Create(&u).Error; err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, u)
	})

	e.Logger.Fatal(e.Start(":1323"))
}

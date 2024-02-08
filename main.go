package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)



// Game represents a Tic Tac Toe game
type Game struct {
	gorm.Model
	Board    string         // Representing the game board state (e.g., "XOXOXOXOX")
	Users    []*User       `gorm:"many2many:game_users;"`
	GameUsers []*GameUser `gorm:"foreignKey:GameID"`
}

// User represents a user
type User struct {
	gorm.Model
	Username string
	Games    []*Game       `gorm:"many2many:game_users;"`
	GameUsers []*GameUser `gorm:"foreignKey:UserID"`
}

// GameUser represents the join table between games and users
type GameUser struct {
	gorm.Model
	GameID  uint
	UserID  uint
	Role    string // Creator, Opponent, Spectator
}

// Define a JWT claims struct
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
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
	db.AutoMigrate(&Game{}, &User{}, &GameUser{})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})


	// Define the handler for POST request to generate JWT token
	e.POST("/login", func(c echo.Context) error {
		req := new(User)
		if err := c.Bind(req); err != nil {
			return err
		}
		// Find user by username
		var user User
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		}
		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		}
		// Generate JWT token
		claims := &Claims{
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("secret"))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{"token": tokenString})
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

package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/modules/auth"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories"
	"gitlab.com/jideobs/nebularcore/modules/health"
	"gitlab.com/jideobs/nebularcore/examples/models"
	"golang.org/x/crypto/bcrypt"
)

// ExampleSettings demonstrates project-specific settings
type ExampleSettings struct {
	App struct {
		Name        string        `yaml:"name" validate:"required"`
		Description string        `yaml:"description"`
		Timeout     time.Duration `yaml:"timeout" validate:"required"`
	} `yaml:"app"`

	Metrics struct {
		Enabled bool   `yaml:"enabled"`
		Path    string `yaml:"path"`
	} `yaml:"metrics"`
}

// Implement Settings interface
func (s ExampleSettings) Validate() error {
	return nil
}

func (s ExampleSettings) IsProduction() bool {
	return false
}

func main() {
	// Create application
	app := nebularcore.New(core.Options[ExampleSettings]{
		ConfigPath: "./config.yml",
		EnvPrefix:  "NEBULAR",
	})

	// Create and register health module
	healthModule := health.New()
	if err := app.RegisterModule(healthModule); err != nil {
		panic(err)
	}

	// Initialize user repository with custom user model
	userFactory := models.NewCustomUserFactory()
	userRepo := repositories.NewUserRepository(app.DB(), userFactory)

	// Create and register auth module with custom user repository
	authModule := auth.New().WithUserRepository(userRepo)
	if err := app.RegisterModule(authModule); err != nil {
		panic(err)
	}

	if err := app.Bootstrap(context.Background()); err != nil {
		panic(err)
	}

	// Register handler
	app.Router().Group("").POST("/register", func(c *gin.Context) {
		type RegisterRequest struct {
			Email       string     `json:"email" binding:"required,email"`
			Username    string     `json:"username" binding:"required"`
			Password    string     `json:"password" binding:"required"`
			PhoneNumber string     `json:"phone_number"`
			FirstName   string     `json:"first_name" binding:"required"`
			LastName    string     `json:"last_name" binding:"required"`
			CompanyName string     `json:"company_name"`
			Department  string     `json:"department"`
			Role       string     `json:"role"`
			Address     string     `json:"address"`
			DateOfBirth *time.Time `json:"date_of_birth"`
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Hash password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}

		// Create user data
		userData := map[string]any{
			"email":        req.Email,
			"username":     req.Username,
			"password":     string(passwordHash),
			"phone_number": req.PhoneNumber,
			"first_name":   req.FirstName,
			"last_name":    req.LastName,
			"company_name": req.CompanyName,
			"department":   req.Department,
			"role":        req.Role,
			"address":     req.Address,
			"date_of_birth": req.DateOfBirth,
			"active":       true,
		}

		// Create user
		user, err := userRepo.Create(c.Request.Context(), userData)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Type assert to get custom user methods
		if customUser, ok := user.(*models.CustomUser); ok {
			c.JSON(201, gin.H{
				"id":           customUser.GetID(),
				"email":        customUser.GetEmail(),
				"username":     customUser.GetUsername(),
				"full_name":    customUser.GetFullName(),
				"company_info": gin.H{
					"company":    customUser.CompanyName,
					"department": customUser.Department,
				},
			})
		} else {
			c.JSON(201, user)
		}
	})

	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}

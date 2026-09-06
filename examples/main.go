package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/examples/models"
	"github.com/volvlabs/nebularcore/modules/auth"
	"github.com/volvlabs/nebularcore/modules/auth/repositories"
	"github.com/volvlabs/nebularcore/modules/event"
	"github.com/volvlabs/nebularcore/modules/storage"
	"golang.org/x/crypto/bcrypt"
)

// ExampleSettings demonstrates project-specific settings
type ExampleSettings struct {
	App struct {
	} `yaml:"app"`

	Metrics struct {
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
	app := nebularcore.New(core.Options[ExampleSettings]{
		ConfigPath: "./config.yml",
		EnvPrefix:  "NEBULAR",
	})

	userFactory := models.NewCustomUserFactory()
	userRepo := repositories.NewUserRepository(app.DB(), userFactory)

	eventBus, err := event.New()
	if err != nil {
		panic(err)
	}

	if err := app.RegisterModule(eventBus); err != nil {
		panic(err)
	}

	authModule := auth.New(eventBus).WithUserRepository(userRepo)
	if err := app.RegisterModule(authModule); err != nil {
		panic(err)
	}

	storageModule := storage.New()

	if err := app.RegisterModule(storageModule); err != nil {
		panic(err)
	}

	if err := app.Bootstrap(context.Background()); err != nil {
		panic(err)
	}

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
			Role        string     `json:"role"`
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
			"email":         req.Email,
			"username":      req.Username,
			"password":      string(passwordHash),
			"phone_number":  req.PhoneNumber,
			"first_name":    req.FirstName,
			"last_name":     req.LastName,
			"company_name":  req.CompanyName,
			"department":    req.Department,
			"role":          req.Role,
			"address":       req.Address,
			"date_of_birth": req.DateOfBirth,
			"active":        true,
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
				"id":        customUser.GetID(),
				"email":     customUser.GetEmail(),
				"username":  customUser.GetUsername(),
				"full_name": customUser.GetFullName(),
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

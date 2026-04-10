package main

import (
	"log"
	"path/filepath"

	"github.com/volvlabs/nebularcore"
	"github.com/volvlabs/nebularcore/apis"
	"github.com/volvlabs/nebularcore/models/config"
	"github.com/volvlabs/nebularcore/tools/auth"
	"github.com/volvlabs/nebularcore/tools/filesystem"
)

func main() {
	cfg, err := config.New("config.yml")
	if err != nil {
		log.Fatalf("error setting up config: %v", err)
	}

	cfg.Env = "development"
	cfg.IsDev = true
	cfg.EnforceAcl = true
	cfg.Server.AllowedOrigins = "*"
	app := nebularcore.New(cfg).(*nebularcore.NebularCore)

	// routes
	r := app.Router()
	rg := r.Group("/api")
	apis.BindHealthApi(rg)

	// Register roles
	policyPath := filepath.Join(filesystem.GetRootDir(""), "test/data/policy.csv")
	confPath := filepath.Join(filesystem.GetRootDir(""), "test/data/conf.csv")
	app.Acm().RegisterAll([]auth.AclConfig{
		{Role: "superAdmin", PolicyPath: policyPath, ConfPath: confPath},
		{Role: "admin", PolicyPath: policyPath, ConfPath: confPath},
		{Role: "user", PolicyPath: policyPath, ConfPath: confPath},
		{Role: "developers", PolicyPath: policyPath, ConfPath: confPath},
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"gitlab.com/jideobs/nebularcore"
	"gitlab.com/jideobs/nebularcore/apis"
	"gitlab.com/jideobs/nebularcore/models/config"
	"gitlab.com/jideobs/nebularcore/tools/auth"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
)

func main() {
	cfg, err := config.New("config.yml")
	if err != nil {
		log.Fatalf("error setting up config: %v", err)
	}

	cfg.Env = "test"
	cfg.IsDev = true
	cfg.BaseDir = fmt.Sprintf("%s/test/data", filesystem.GetRootDir(""))
	cfg.EnforceAcl = true
	cfg.Server.AllowedOrigins = "*"
	app := nebularcore.New(cfg)

	// routes
	r := app.Router()
	rg := r.Group("/api")
	apis.BindAdminApi(app, rg)
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

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"gitlab.com/volvlabs/nebularcore"
	"gitlab.com/volvlabs/nebularcore/apis"
	"gitlab.com/volvlabs/nebularcore/models/config"
	"gitlab.com/volvlabs/nebularcore/tools/auth"
	"gitlab.com/volvlabs/nebularcore/tools/filesystem"
)

func main() {
	cfg, err := config.New("")
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

	app.Execute()
}

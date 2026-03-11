package main

import (
	"example/env"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/thoriqadillah/gema"
)

func emailConfig() *gema.EmailerOption {
	return &gema.EmailerOption{
		Env:        env.APP_ENV,
		TemplateFs: templateFs,
		Host:       env.MAILER_HOST,
		Port:       env.MAILER_PORT,
		Username:   env.MAILER_USER,
		Password:   env.MAILER_PASS,
		From:       env.MAILER_FROM,
		Name:       env.MAILER_NAME,
	}
}

func storageConfig() *gema.LocalStorageOption {
	return &gema.LocalStorageOption{
		TempDir:       "./storage",
		FullRoutePath: fmt.Sprintf("http://localhost:%d/storage", env.PORT),
	}
}

func queueConfig() map[string]river.QueueConfig {
	return map[string]river.QueueConfig{
		river.QueueDefault: {
			MaxWorkers: 100,
		},
	}
}

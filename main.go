package main

import (
	"log"

	"github.com/Milad75Rasouli/portfolio/internal/config"
	"github.com/Milad75Rasouli/portfolio/internal/handler"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.New()
	log.Printf("Config:%+v", cfg)

	var (
		logger *zap.Logger
		err    error
	)
	if cfg.Debug == true {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalln(err)
	}
	defer logger.Sync()
	app := fiber.New(fiber.Config{
		Immutable: true,
	})
	{
		logger := logger.Named("http")
		h := handler.Home{
			Logger: logger.Named("home"),
		}

		b := handler.Blog{
			Logger: logger.Named("blog"),
		}

		a := handler.Auth{
			Logger: logger.Named("auth"),
		}

		home := app.Group("/")
		blog := app.Group("/blog")
		auth := app.Group("/user")

		h.Register(home)
		b.Register(blog)
		a.Register(auth)
	}

	app.Static("/static", "./frontend/static")
	app.Listen(":5000")
}
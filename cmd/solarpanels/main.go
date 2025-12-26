package main

import (
	"context"
	"fmt"

	"lab/internal/app/config"
	"lab/internal/app/dsn"
	"lab/internal/app/handler"
	"lab/internal/app/repository"
	"lab/internal/app/service"
	"lab/internal/pkg"
	"lab/internal/redis"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "lab/docs"
)

// @title Solar Panels Power Calculator API
// @version 1.0
// @description API для управления солнечными панелями и расчета их мощности
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8001
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	router := gin.Default()
	router.Use(
		cors.New(cors.Config{AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8001",
			"*",
		},
			AllowMethods: []string{
				"GET", "POST", "PUT", "DELETE", "OPTIONS",
			},
			AllowHeaders: []string{
				"Origin", "Content-Type", "Accept", "Authorization",
			},
			ExposeHeaders: []string{
				"Content-Length",
			},
			AllowCredentials: true,
		}),
	)
	router.GET("/swagger/", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger/json", func(ctx *gin.Context) {
		ctx.File("docs/swagger.json")
	})
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}
	serv := service.NewService(rep)
	redis, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}
	hand := handler.NewHandler(serv, redis, conf)

	SolarPanelPowerCalculator := pkg.NewApp(conf, router, hand)

	SolarPanelPowerCalculator.RunApp()
}

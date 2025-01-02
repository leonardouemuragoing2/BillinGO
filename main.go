package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"billingo/config"
	"billingo/controllers"
	"billingo/models"
	"billingo/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// setupLog configures log output level with "INFO" as default
func setupLog(conf *config.Config) {
	executable, _ := os.Executable()
	executable = filepath.FromSlash(executable)
	directory := filepath.Dir(executable)
	log.SetFormatter(&log.TextFormatter{
		DisableQuote: true,
	})
	logLevel := strings.ToUpper(conf.LogLevel)
	if logLevel == "DEBUG" {
		log.SetLevel(log.DebugLevel)

	} else if logLevel == "ERROR" {
		log.SetLevel(log.ErrorLevel)

	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Infof("Executable: %v", executable)
	log.Infof("Dir: %v", directory)
	log.Infof("Log Level: %v", log.GetLevel())

}

func run(conf *config.Config) {
	setupLog(conf)

	log.Info("**** vStack Billing Service")
	log.Infof("CPU: %v", runtime.GOARCH)
	log.Infof("Platform: %v", runtime.GOOS)

	// Create router
	r := gin.Default()
	db := models.SetupModels(conf)
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	panic(fmt.Errorf("failed to obtain sql.DB instance: %w", err))
	// }
	manager := controllers.NewManager(ctx, db)

	// Register tasks
	manager.AddTask(controllers.NewMetrics("MetricsTask"))
	manager.AddTask(controllers.NewObserverBuffer("ObserverBufferTask"))
	manager.AddTask(controllers.NewBatchSaveToDatabaseTask("BatchSaveToDatabaseTask", "ObserverBufferTask", 3000))
	// manager.AddTask(controllers.NewMQTTPublisherTask("MQTTPublisherTask", "mqtt://public.mqtt.pro:1883", "12345qwert54321", "ajbkvbp/device-test", "wNmDtwPghSTaBmJJ"))
	manager.AddTask(controllers.NewMQTTPublisherTask("MQTTPublisherTask", "mqtt://localhost:1883", "12345qwert54321", "ajbkvbp/device-test", "wNmDtwPghSTaBmJJ"))
	manager.AddTask(controllers.NewMQTTSubscriber("mqtt://localhost:1883", "12345qwert54321"))

	// manager.AddTask(controllers.NewDatabaseSaverTask("ObserverBufferTask"))

	// Provide db and manager to context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("manager", manager)
		c.Next()
	})

	// Endpoints configuration
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// r.POST("/api/v1/auth/token", routers.Generate)
	api := r.Group("/api/v1") //.Use(routers.Authentication).Use(limit.MaxAllowed(30))
	{
		routers.GetEndpoints(api)
	}

	go func() {
		if err := r.Run(fmt.Sprintf(":%v", conf.GinPort)); err != nil {
			log.Fatal(err)
		}
	}()

	manager.StartAll()
	<-c
	manager.StopAll()
	cancel()
}
func main() {
	conf := config.LoadConfig()
	run(conf)
}

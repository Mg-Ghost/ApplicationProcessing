package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"meddoc/internal/handlers"
	"meddoc/internal/middleware"
	"meddoc/internal/repository"
	"meddoc/internal/scheduler"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using system environment")
	}

	db, err := repository.NewDB(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Start cron: auto-escalate tickets after 5 working days
	sched := scheduler.New(db)
	sched.Start()
	defer sched.Stop()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	h := handlers.New(db)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/admin/login", h.AdminLogin)
		}

		user := api.Group("/user")
		user.Use(middleware.AuthRequired())
		{
			user.GET("/profile", h.GetProfile)
			user.PUT("/profile", h.UpdateProfile)
		}

		tickets := api.Group("/tickets")
		tickets.Use(middleware.AuthRequired())
		{
			tickets.GET("", h.ListTickets)
			tickets.POST("", h.CreateTicket)
			tickets.GET("/:id", h.GetTicket)
			tickets.PUT("/:id", h.UpdateTicket)
			tickets.PATCH("/:id/cancel", h.CancelTicket)
			tickets.PATCH("/:id/close", h.CloseTicket)
			tickets.POST("/:id/reply", h.UserReply) 
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.GET("/tickets", h.AdminListTickets)
			admin.GET("/tickets/export", h.ExportTickets)
			admin.GET("/tickets/:id", h.AdminGetTicket)
			admin.DELETE("/tickets/:id", h.AdminDeleteTicket)
			admin.PATCH("/tickets/:id/close", h.AdminCloseTicket)
			admin.POST("/tickets/:id/comment", h.AddComment)
			admin.GET("/ip-logs", h.GetIPLogs)
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s", port)
	r.Run(":" + port)
}

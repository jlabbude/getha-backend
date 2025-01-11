package main

import (
	"fmt"
	"getha/aparelhos"
	"getha/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func main() {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=5432",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_HOST"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&models.Aparelho{}); err != nil {
		panic(err)
	}
	models.DATABASE = db

	router := gin.Default()
	router.POST("/create_aparelho", aparelhos.CreateAparelho)
	router.GET("/serve_ids", aparelhos.ServeAparelhoIDList)
	router.GET("/serve_image", aparelhos.ServeImage)
	router.GET("/serve_manual", aparelhos.ServeManual)
	router.GET("/serve_video", aparelhos.ServeVideo)
	router.DELETE("/delete_aparelho", aparelhos.DeleteAparelho)
	err = router.Run(":80")
	if err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"getha/aparelhos"
	"getha/models"
	"getha/zoonose"
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
	if err = db.AutoMigrate(&models.Aparelho{}, &models.Zoonose{}); err != nil {
		panic(err)
	}
	models.DATABASE = db

	router := gin.Default()

	router.POST("/create_aparelho", aparelhos.CreateAparelho)
	router.DELETE("/delete_aparelho", aparelhos.DeleteAparelho)
	router.GET("/serve_ids", aparelhos.ServeAparelhoIDList)
	router.GET("/serve_image", aparelhos.ServeImage)
	router.GET("/serve_manual", aparelhos.ServeManual)
	router.GET("/serve_video", aparelhos.ServeVideo)

	router.POST("/create_zoonose", zoonose.CreateZoonose)
	router.DELETE("/delete_zoonose", zoonose.DeleteZoonose)
	router.GET("/serve_zoonose_ids", zoonose.ServeZoonoseIDList)
	router.GET("/get_zoonose", zoonose.GetZoonoseInfo)

	err = router.Run(":80")
	if err != nil {
		panic(err)
	}
}

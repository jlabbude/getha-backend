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
	db.Exec("CREATE TYPE organismo AS ENUM ('Virus', 'Bacteria', 'Fungo', 'Protozoario', 'Helminto')")
	if err = db.AutoMigrate(
		&models.Aparelhos{},
		&models.Zoonose{},
		&models.Vetores{},
		&models.Agentes{},
		&models.Transmissoes{},
		&models.Profilaxias{},
		&models.Diagnosticos{},
		&models.Regioes{},
	); err != nil {
		panic(err)
	}
	models.DATABASE = db

	router := gin.Default()

	router.POST("/create_aparelho", aparelhos.CreateAparelho)
	router.DELETE("/delete_aparelho", aparelhos.DeleteAparelho)
	router.GET("/serve_aparelhos_ids", aparelhos.ServeAparelhoIDList)
	router.GET("/serve_image", aparelhos.ServeImage)
	router.GET("/serve_manual", aparelhos.ServeManual)
	router.GET("/serve_video", aparelhos.ServeVideo)
	router.PUT("/update_aparelho_video", aparelhos.UpdateAparelhoVideo)

	router.POST("/create_zoonose", zoonose.CreateZoonose)
	router.DELETE("/delete_zoonose", zoonose.DeleteZoonose)
	router.GET("/serve_zoonose_ids", zoonose.ServeZoonoseIDList)
	router.GET("/get_card_info", zoonose.GetZoonoseCardInfo)
	router.GET("/get_zoonose_full", zoonose.GetZoonoseFullInfo)

	err = router.Run(":80")
	if err != nil {
		panic(err)
	}
}

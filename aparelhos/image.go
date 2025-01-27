package aparelhos

import (
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func ServeImage(context *gin.Context) {
	var imagemPath string

	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = models.DATABASE.Model(&models.Aparelhos{}).
		Where("id = ?", id).
		Select("ImagePath").
		Limit(1).
		Scan(&imagemPath).
		Error

	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if imagemPath == "" {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Aparelho não encontrado"})
		return
	}

	//fixme handle file not existing
	context.File(imagemPath)
}

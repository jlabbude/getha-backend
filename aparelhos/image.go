package aparelhos

import (
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"path/filepath"
)

func ServeImage(context *gin.Context) {
	var imagemPath string

	id, err := uuid.Parse(context.Query("ID"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = models.DATABASE.Model(&models.Aparelhos{}).
		Where("ID = ?", id).
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

	if matches, err := filepath.Glob(imagemPath); len(matches) <= 0 {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "arquivo de imagem não encontrado, considere deletar e recadastrar esse aparelho",
		})
		return
	} else if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.File(imagemPath)
}

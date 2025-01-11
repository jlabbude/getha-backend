package aparelhos

import (
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func ServeManual(context *gin.Context) {
	var manualPath string

	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = models.DATABASE.Model(&models.Aparelho{}).
		Where("id = ?", id).
		Select("ManualPath").
		Limit(1).
		Scan(&manualPath).
		Error

	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}
	context.File(manualPath)
}

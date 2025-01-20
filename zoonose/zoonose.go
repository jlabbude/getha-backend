package zoonose

// Condensed all the code here since it was less complex compared to the file streaming needed on aparelhos

import (
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func ServeZoonoseIDList(context *gin.Context) {
	var ids []uuid.UUID

	if err := models.DATABASE.Model(&models.Zoonose{}).Pluck("ID", &ids).Error; err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, ids)
}

func CreateZoonose(context *gin.Context) {
	id := uuid.New()
	nome := context.PostForm("nome")
	transmissao := context.PostForm("transmissao")
	agente := context.PostForm("agente")
	descricao := context.PostForm("descricao")
	profilaxia := context.PostForm("profilaxia")
	vetor := context.PostForm("vetor")

	if nome == "" ||
		transmissao == "" ||
		agente == "" ||
		descricao == "" ||
		profilaxia == "" ||
		vetor == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Todos os campos devem ser preenchidos."})
	}

	zoonose := models.Zoonose{
		ID:          id,
		Nome:        nome,
		Transmissao: transmissao,
		Agente:      agente,
		Descricao:   descricao,
		Profilaxia:  profilaxia,
		Vetor:       vetor,
	}

	if result := models.DATABASE.Create(&zoonose); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"zoonose": zoonose, "id": zoonose.ID})

}

func DeleteZoonose(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID não fornecido."})
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido."})
		return
	}

	zoonose := models.Zoonose{}
	if result := models.DATABASE.Where("id = ?", id).First(&zoonose); result.Error != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Zoonose não encontrada."})
		return
	}

	if result := models.DATABASE.Delete(&zoonose); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Zoonose deletada."})
}

func GetZoonoseInfo(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID não fornecido."})
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido."})
		return
	}

	var zoonose models.Zoonose
	if result := models.DATABASE.First(&zoonose, "id = ?", id); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"id":          zoonose.ID,
		"nome":        zoonose.Nome,
		"agente":      zoonose.Agente,
		"vetor":       zoonose.Vetor,
		"transmissao": zoonose.Transmissao,
		"profilaxia":  zoonose.Profilaxia,
		"descricao":   zoonose.Descricao,
	})

}

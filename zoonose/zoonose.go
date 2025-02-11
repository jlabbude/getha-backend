package zoonose

// Condensed all the code here since it was less complex compared to the file streaming needed on aparelhos

import (
	"errors"
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type JSONZoonose struct {
	Nome           string   `json:"nome" binding:"required"`
	NomeCientifico string   `json:"nome_cientifico" binding:"required"`
	Organismo      string   `json:"organismo" binding:"required"`
	Descricao      string   `json:"descricao" binding:"required"`
	Vetores        []string `json:"vetores" binding:"required"`
	Agentes        []string `json:"agentes" binding:"required"`
	Transmissoes   []string `json:"transmissoes" binding:"required"`
	Regioes        []string `json:"regioes" binding:"required"`
	Profilaxias    []string `json:"profilaxias" binding:"required"`
	Diagnosticos   []string `json:"diagnosticos" binding:"required"`
}

func ServeZoonoseIDList(context *gin.Context) {
	var ids []uuid.UUID

	if err := models.DATABASE.Model(&models.Zoonose{}).Pluck("ID", &ids).Error; err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, ids)
}

func CreateZoonose(context *gin.Context) {
	context.Header("Content-Type", "application/json; charset=utf-8")
	var auxZoo JSONZoonose
	id := uuid.New()

	if err := context.ShouldBindBodyWithJSON(&auxZoo); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if organismo := auxZoo.Organismo; organismo != string(models.Bacteria) &&
		organismo != string(models.Virus) &&
		organismo != string(models.Fungo) &&
		organismo != string(models.Protozoario) &&
		organismo != string(models.Helminto) {

		context.JSON(http.StatusBadRequest, gin.H{"error": "Organismo inválido.", "organismo": organismo})
		return
	}

	if auxZoo.Nome == "" ||
		auxZoo.Descricao == "" ||
		//organismo == "" ||
		auxZoo.NomeCientifico == "" ||
		len(auxZoo.Agentes) == 0 ||
		len(auxZoo.Vetores) == 0 ||
		len(auxZoo.Transmissoes) == 0 ||
		len(auxZoo.Profilaxias) == 0 ||
		len(auxZoo.Regioes) == 0 ||
		len(auxZoo.Diagnosticos) == 0 {

		context.JSON(http.StatusBadRequest, gin.H{"error": "Todos os campos devem ser preenchidos."})
		return
	}

	zoonose := models.Zoonose{
		ID:             id,
		Nome:           auxZoo.Nome,
		NomeCientifico: auxZoo.NomeCientifico,
		Descricao:      auxZoo.Descricao,
		Organismo:      models.Organismo(auxZoo.Organismo),
		Agentes:        make([]models.Agentes, len(auxZoo.Agentes)),
		Transmissoes:   make([]models.Transmissoes, len(auxZoo.Transmissoes)),
		Vetores:        make([]models.Vetores, len(auxZoo.Vetores)),
		Regioes:        make([]models.Regioes, len(auxZoo.Regioes)),
		Profilaxias:    make([]models.Profilaxias, len(auxZoo.Profilaxias)),
		Diagnosticos:   make([]models.Diagnosticos, len(auxZoo.Diagnosticos)),
	}

	for i, agente := range auxZoo.Agentes {
		zoonose.Agentes[i] = models.Agentes{
			Agentes:   agente,
			ZoonoseID: id,
		}
	}

	for i, vetor := range auxZoo.Vetores {
		zoonose.Vetores[i] = models.Vetores{
			Vetores:   vetor,
			ZoonoseID: id,
		}
	}

	for i, transmissao := range auxZoo.Transmissoes {
		zoonose.Transmissoes[i] = models.Transmissoes{
			Transmissoes: transmissao,
			ZoonoseID:    id,
		}
	}

	for i, profilaxia := range auxZoo.Profilaxias {
		zoonose.Profilaxias[i] = models.Profilaxias{
			Profilaxias: profilaxia,
			ZoonoseID:   id,
		}
	}

	for i, sintoma := range auxZoo.Diagnosticos {
		zoonose.Diagnosticos[i] = models.Diagnosticos{
			Diagnosticos: sintoma,
			ZoonoseID:    id,
		}
	}

	for i, regiao := range auxZoo.Regioes {
		zoonose.Regioes[i] = models.Regioes{
			Regioes:   regiao,
			ZoonoseID: id,
		}
	}

	if result := models.DATABASE.Create(&zoonose); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error(), "zoonose": zoonose})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"zoonose": zoonose, "id": zoonose.ID})

}

func DeleteZoonose(context *gin.Context) {
	id := context.Query("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID não fornecido."})
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido."})
	}

	tx := models.DATABASE.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var zoonose models.Zoonose
	if err := tx.First(&zoonose, "id = ?", id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusNotFound, gin.H{"error": "Zoonose not found"})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := tx.Where("zoonose_id = ?", id).Delete(&models.Agentes{}).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agentes"})
		return
	}

	if err := tx.Where("zoonose_id = ?", id).Delete(&models.Transmissoes{}).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transmissoes"})
		return
	}

	if err := tx.Where("zoonose_id = ?", id).Delete(&models.Vetores{}).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vetores"})
		return
	}

	if err := tx.Where("zoonose_id = ?", id).Delete(&models.Regioes{}).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete regioes"})
		return
	}

	if err := tx.Where("zoonose_id = ?", id).Delete(&models.Profilaxias{}).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profilaxias"})
		return
	}

	if err := tx.Where("zoonose_id = ?", id).Delete(&models.Diagnosticos{}).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete diagnosticos"})
		return
	}

	if err := tx.Delete(&zoonose).Error; err != nil {
		tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete zoonose"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	context.String(http.StatusNoContent, "Zoonose deletada com sucesso.")
}

func GetZoonoseCardInfo(context *gin.Context) {
	context.Header("Content-Type", "application/json; charset=utf-8")
	id := context.Query("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID não fornecido."})
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido."})
		return
	}

	var zoonose models.Zoonose

	if result := models.
		DATABASE.
		Select("id", "nome", "nome_cientifico", "organismo").
		First(&zoonose, "id = ?", id); result.Error != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"id":              zoonose.ID,
		"nome":            zoonose.Nome,
		"nome_cientifico": zoonose.NomeCientifico,
		"organismo":       zoonose.Organismo,
	})

}

type InfoAuxZoonose struct {
	Agentes      []string
	Vetores      []string
	Transmissoes []string
	Profilaxias  []string
	Diagnosticos []string
	Regioes      []string
}

func GetZoonoseFullInfo(context *gin.Context) {
	context.Header("Content-Type", "application/json; charset=utf-8")
	id := context.Query("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID não fornecido."})
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido."})
		return
	}

	var zoonose models.Zoonose

	if result := models.
		DATABASE.
		Preload("Agentes").
		Preload("Vetores").
		Preload("Transmissoes").
		Preload("Profilaxias").
		Preload("Regioes").
		Preload("Diagnosticos").
		First(&zoonose, "id = ?", id); result.Error != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	infoauxzoo := InfoAuxZoonose{
		Diagnosticos: mapSlice(zoonose.Diagnosticos, func(sintomas models.Diagnosticos) string { return sintomas.Diagnosticos }),
		Profilaxias:  mapSlice(zoonose.Profilaxias, func(profilaxias models.Profilaxias) string { return profilaxias.Profilaxias }),
		Transmissoes: mapSlice(zoonose.Transmissoes, func(transmissoes models.Transmissoes) string { return transmissoes.Transmissoes }),
		Vetores:      mapSlice(zoonose.Vetores, func(vetores models.Vetores) string { return vetores.Vetores }),
		Regioes:      mapSlice(zoonose.Regioes, func(regioes models.Regioes) string { return regioes.Regioes }),
		Agentes:      mapSlice(zoonose.Agentes, func(agentes models.Agentes) string { return agentes.Agentes }),
	}

	context.JSON(http.StatusOK, gin.H{
		"id":              zoonose.ID,
		"nome":            zoonose.Nome,
		"nome_cientifico": zoonose.NomeCientifico,
		"descricao":       zoonose.Descricao,
		"organismo":       zoonose.Organismo,
		"regioes":         infoauxzoo.Regioes,
		"agentes":         infoauxzoo.Agentes,
		"vetores":         infoauxzoo.Vetores,
		"transmissoes":    infoauxzoo.Transmissoes,
		"profilaxia":      infoauxzoo.Profilaxias,
		"diagnosticos":    infoauxzoo.Diagnosticos,
	})
}

func mapSlice[T any](source []T, extract func(T) string) []string {
	result := make([]string, len(source))
	for i, item := range source {
		result[i] = extract(item)
	}
	return result
}

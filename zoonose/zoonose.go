package zoonose

// Condensed all the code here since it was less complex compared to the file streaming needed on aparelhos

import (
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	Profilaxias    []string `json:"profilaxias" binding:"required"`
	Sintomas       []string `json:"sintomas" binding:"required"`
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
	var auxZoo JSONZoonose
	id := uuid.New()

	if err := context.ShouldBindBodyWithJSON(&auxZoo); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if organismo := auxZoo.Organismo; organismo != string(models.Bacteria) &&
		organismo != string(models.Virus) &&
		organismo != string(models.Fungo) &&
		organismo != string(models.Protozoario) {

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
		len(auxZoo.Sintomas) == 0 {

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
		Vetores:        make([]models.Vetores, len(auxZoo.Vetores)),
		Transmissoes:   make([]models.Transmissoes, len(auxZoo.Transmissoes)),
		Profilaxias:    make([]models.Profilaxias, len(auxZoo.Profilaxias)),
		Sintomas:       make([]models.Sintomas, len(auxZoo.Sintomas)),
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

	for i, sintoma := range auxZoo.Sintomas {
		zoonose.Sintomas[i] = models.Sintomas{
			Sintomas:  sintoma,
			ZoonoseID: id,
		}
	}

	if result := models.DATABASE.Create(&zoonose); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error(), "zoonose": zoonose})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"zoonose": zoonose, "id": zoonose.ID})

}

func DeleteZoonose(context *gin.Context) { // fixme
	id := context.Query("id")
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

func GetZoonoseCardInfo(context *gin.Context) {
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
	Sintomas     []string
}

func GetZoonoseFullInfo(context *gin.Context) {
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
		Preload("Sintomas").
		First(&zoonose, "id = ?", id); result.Error != nil {

		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	infoauxzoo := InfoAuxZoonose{
		Sintomas:     mapSlice(zoonose.Sintomas, func(sintomas models.Sintomas) string { return sintomas.Sintomas }),
		Profilaxias:  mapSlice(zoonose.Profilaxias, func(profilaxias models.Profilaxias) string { return profilaxias.Profilaxias }),
		Transmissoes: mapSlice(zoonose.Transmissoes, func(transmissoes models.Transmissoes) string { return transmissoes.Transmissoes }),
		Vetores:      mapSlice(zoonose.Vetores, func(vetores models.Vetores) string { return vetores.Vetores }),
		Agentes:      mapSlice(zoonose.Agentes, func(agentes models.Agentes) string { return agentes.Agentes }),
	}

	context.JSON(http.StatusOK, gin.H{
		"id":              zoonose.ID,
		"nome":            zoonose.Nome,
		"nome_cientifico": zoonose.NomeCientifico,
		"descricao":       zoonose.Descricao,
		"organismo":       zoonose.Organismo,
		"agentes":         infoauxzoo.Agentes,
		"vetores":         infoauxzoo.Vetores,
		"transmissoes":    infoauxzoo.Transmissoes,
		"profilaxia":      infoauxzoo.Profilaxias,
		"sintomas":        infoauxzoo.Sintomas,
	})
}

func mapSlice[T any](source []T, extract func(T) string) []string {
	result := make([]string, len(source))
	for i, item := range source {
		result[i] = extract(item)
	}
	return result
}

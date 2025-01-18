package aparelhos

import (
	"fmt"
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const AparelhosPath = "/app/aparelhos"

func ServeAparelhoIDList(context *gin.Context) {
	var ids []uuid.UUID

	if err := models.DATABASE.Model(&models.Aparelho{}).Pluck("ID", &ids).Error; err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, ids)
}

func CreateAparelho(context *gin.Context) {
	id := uuid.New()
	localDir := path.Join(AparelhosPath, id.String())
	if err := os.Mkdir(localDir, 0777); err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
	}

	nome := context.PostForm("nome")
	if nome == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no preenchimento do nome."})
		return
	}

	image, err := context.FormFile("image_path")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no upload da imagem, " + err.Error()})
		return
	} else if path.Ext(image.Filename) != ".png" && path.Ext(image.Filename) != ".jpg" && path.Ext(image.Filename) != ".jpeg" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formato de imagem inválido. Apenas .png, .jpg e .jpeg são aceitos."})
		return
	}
	imageSrc, err := image.Open()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer imageSrc.Close()
	imagePath := fmt.Sprintf("%s/%s", localDir, id.String()+path.Ext(image.Filename))
	if imageDest, err := os.Create(imagePath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		// fixme resize to square
		if _, err = io.Copy(imageDest, imageSrc); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	videoDest, err := context.FormFile("video_path")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no upload do vídeo, " + err.Error()})
		return
	} else if path.Ext(videoDest.Filename) != ".mp4" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formato de vídeo inválido. Apenas .mp4 é aceito."})
		return
	}
	videoSrc, err := videoDest.Open()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer videoSrc.Close()
	videoPath := fmt.Sprintf("%s/%s", localDir, id.String()+path.Ext(videoDest.Filename))
	if videoDest, err := os.Create(videoPath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		if _, err = io.Copy(videoDest, videoSrc); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	manual, err := context.FormFile("manual_path")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no upload do manual, " + err.Error()})
		return
	} else if path.Ext(manual.Filename) != ".pdf" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formato de manual inválido. Apenas .pdf é aceito"})
		return
	}
	manualPath := fmt.Sprintf("%s/%s", localDir, id.String()+path.Ext(manual.Filename))
	manualSrc, err := manual.Open()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer manualSrc.Close()
	if manualDest, err := os.Create(manualPath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		if _, err = io.Copy(manualDest, manualSrc); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	aparelho := models.Aparelho{
		ID:         id,
		Nome:       nome,
		ImagePath:  imagePath,
		VideoPath:  videoPath,
		ManualPath: manualPath,
	}

	if result := models.DATABASE.Create(&aparelho); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	} else {
		context.String(http.StatusOK, "Aparelho criado com id: "+id.String())
		return
	}
}

func DeleteAparelho(context *gin.Context) { // fixme auth
	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formatação de ID inválida"})
		return
	}

	if _, err = models.DATABASE.Model(&models.Aparelho{}).Where("id = ?", id).Delete(&models.Aparelho{}).Rows(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	matches, err := filepath.Glob(id.String())
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, match := range matches {
		if err = os.Remove(match); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	context.JSON(http.StatusOK, gin.H{"message": "Aparelho deletado com sucesso"})
	return
}

package aparelhos

import (
	"errors"
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

var AparelhoPath = "/app/aparelhos"

type FileType []string

var (
	Image  = FileType([]string{".png", ".jpg", ".jpeg"})
	Video  = FileType([]string{".mp4"})
	Manual = FileType([]string{".pdf"})
)

type AparelhoJSON struct {
	ID   uuid.UUID `json:"id"`
	Nome string    `json:"nome"`
}

func ServeAparelhos(context *gin.Context) {
	var aparelhos []models.Aparelhos

	if err := models.
		DATABASE.
		Model(&models.Aparelhos{}).
		Select("ID", "nome").
		Find(&aparelhos).Error; err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	aparelhosJSON := make([]AparelhoJSON, len(aparelhos))
	for i, aparelho := range aparelhos {
		aparelhosJSON[i] = AparelhoJSON{
			ID:   aparelho.ID,
			Nome: aparelho.Nome,
		}
	}

	context.JSON(200, aparelhosJSON)
}

func CreateAparelho(context *gin.Context) {
	id := uuid.New()
	localDir := path.Join(AparelhoPath, id.String())
	if err := os.Mkdir(localDir, 0777); err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	nome := context.PostForm("nome")
	if nome == "" {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": "Falha no preenchimento do nome."}, localDir)
		return
	}

	image, err := context.FormFile("image_path")
	if err != nil {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": "Falha no upload da imagem, " + err.Error()}, localDir)
		return
	}
	imagePath, err := CreateFile(id, image, localDir, Image)
	if err != nil {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": err.Error()}, localDir)
		return
	}

	video, err := context.FormFile("video_path")
	if err != nil {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": "Falha no upload do vídeo, " + err.Error()}, localDir)
		return
	}
	videoPath, err := CreateFile(id, video, localDir, Video)
	if err != nil {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": err.Error()}, localDir)
		return
	}

	manual, err := context.FormFile("manual_path")
	if err != nil {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": "Falha no upload do manual, " + err.Error()}, localDir)
		return
	}
	manualPath, err := CreateFile(id, manual, localDir, Manual)
	if err != nil {
		cleanUpAfterError(&context, http.StatusBadRequest, gin.H{"error": err.Error()}, localDir)
		return
	}

	aparelho := models.Aparelhos{
		ID:         id,
		Nome:       nome,
		ImagePath:  imagePath,
		VideoPath:  videoPath,
		ManualPath: manualPath,
	}

	if result := models.DATABASE.Create(&aparelho); result.Error != nil {
		cleanUpAfterError(&context, http.StatusInternalServerError, gin.H{"error": result.Error.Error()}, localDir)
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"message": "Aparelho criado com sucesso",
		"id":      aparelho.ID.String(),
	})
}

func DeleteAparelho(context *gin.Context) { // fixme auth
	id, err := uuid.Parse(context.Query("ID"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formatação de ID inválida"})
		return
	}

	if _, err = models.DATABASE.Model(&models.Aparelhos{}).Where("ID = ?", id).Delete(&models.Aparelhos{}).Rows(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	matches, err := filepath.Glob(path.Join(AparelhoPath, id.String()))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, match := range matches {
		if err = os.RemoveAll(match); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	context.JSON(http.StatusOK, gin.H{"message": "Aparelho deletado com sucesso"})
	return
}

func CreateFile(id uuid.UUID, file *multipart.FileHeader, destPath string, ftypes FileType) (string, error) {
	ext := path.Ext(file.Filename)
	for i, ftype := range ftypes {
		if ext != ftype && i == len(ftypes)-1 {
			return "", errors.New("formato de arquivo inválido " + ext + " " + ftype)
		} else {
			break
		}
	}
	fileHandler, err := file.Open()
	if err != nil {
		return "", errors.New(err.Error() + "aqui1" + ftypes[0])
	}
	defer fileHandler.Close()
	filePath := path.Join(destPath, id.String()+path.Ext(file.Filename))
	if finalFile, err := os.Create(filePath); err != nil {
		return "", errors.New(err.Error() + "aqui2" + ftypes[0])
	} else if _, err = io.Copy(finalFile, fileHandler); err != nil {
		return "", errors.New(err.Error() + "aqui3" + ftypes[0]) // todo consider client sided cropping/resizing for images
	}
	return filePath, nil
}

func UpdateAparelhoVideo(context *gin.Context) {
	id, err := uuid.Parse(context.Query("ID"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formatação de ID inválida"})
		return
	}
	localDir := path.Join(AparelhoPath, id.String())
	video, err := context.FormFile("video_path")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no upload do vídeo, " + err.Error()})
		return
	}
	videoPath, err := CreateFile(id, video, localDir, Video)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result := models.DATABASE.Model(&models.Aparelhos{}).
		Where("ID = ?", id).
		Update("video_path", videoPath); result.Error != nil {
		if errOS := os.Remove(videoPath); errOS != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover o vídeo antigo: " + errOS.Error()})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Vídeo atualizado com sucesso", "video_path": videoPath})
}

func cleanUpAfterError(
	context **gin.Context,
	httpErrorCode int,
	message gin.H,
	localDir string,
) {
	if err := os.RemoveAll(localDir); err != nil {
		(*context).JSON(http.StatusInternalServerError, gin.H{"fatal-error": err, "original-error": message})
		return
	}
	(*context).JSON(httpErrorCode, message)
}

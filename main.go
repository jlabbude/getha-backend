package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

const ImagePath = "/app/aparelhos/image"
const VideoPath = "/app/aparelhos/video"
const ManualPath = "/app/aparelhos/manual"

func serveIDList(context *gin.Context) {
	var ids []uuid.UUID

	if err := DATABASE.Model(&Aparelho{}).Pluck("ID", &ids).Error; err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, ids)
}

func createAparelho(context *gin.Context) {
	id := uuid.New()

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
	imagePath := fmt.Sprintf("%s/%s", ImagePath, id.String()+path.Ext(image.Filename))
	// fixme resize to square
	if _, err = os.Create(imagePath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	video, err := context.FormFile("video_path")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no upload do vídeo, " + err.Error()})
		return
	} else if path.Ext(video.Filename) != ".mp4" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formato de vídeo inválido. Apenas .mp4 é aceito."})
		return
	}
	videoPath := fmt.Sprintf("%s/%s", VideoPath, id.String()+path.Ext(video.Filename))
	if _, err = os.Create(videoPath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	manual, err := context.FormFile("manual_path")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Falha no upload do manual, " + err.Error()})
		return
	} else if path.Ext(manual.Filename) != ".pdf" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Formato de manual inválido. Apenas .pdf é aceito"})
		return
	}
	manualPath := fmt.Sprintf("%s/%s", ManualPath, id.String()+path.Ext(manual.Filename))
	if _, err = os.Create(manualPath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	aparelho := Aparelho{
		ID:         id,
		Nome:       nome,
		ImagePath:  imagePath,
		VideoPath:  videoPath,
		ManualPath: manualPath,
	}

	if result := DATABASE.Create(&aparelho); result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	} else {
		context.String(http.StatusOK, "Aparelho criado com id: ", id.String())
		return
	}
}

func serveImage(context *gin.Context) {
	var imagemPath string

	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = DATABASE.Model(&Aparelho{}).
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

func serveManual(context *gin.Context) {
	var manualPath string

	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = DATABASE.Model(&Aparelho{}).
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

func serveVideo(context *gin.Context) {
	var videoPath string

	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = DATABASE.Model(&Aparelho{}).
		Where("id = ?", id).
		Select("VideoPath").
		Limit(1).
		Scan(&videoPath).
		Error

	file, err := os.Open(videoPath)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir o vídeo"})
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao fechar o file handler do vídeo"})
		}
	}(file)

	fileStat, err := file.Stat()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Falha em adquirir tamanho do arquivo"})
		return
	}

	fileSize := fileStat.Size()
	rangeHeader := context.GetHeader("Range")
	if rangeHeader == "" {
		context.DataFromReader(http.StatusOK, fileSize, "video/mp4", file, nil)
		return
	}
	var start, end int64
	_, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if err != nil || end == 0 {
		end = fileSize - 1
	}

	if start >= fileSize || end >= fileSize || start > end {
		context.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": "Range inválido"})
		return
	}

	contentLength := end - start + 1
	context.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	context.Header("Accept-Ranges", "bytes")
	context.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	context.Status(http.StatusPartialContent)

	_, err = file.Seek(start, 0)
	if err != nil {
		context.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": "Falha ao buscar nessa porção do vídeo"})
	}

	context.Stream(func(w io.Writer) bool {
		buffer := make([]byte, 1024*1024)
		bytesToRead := int64(len(buffer))
		if contentLength < bytesToRead {
			bytesToRead = contentLength
		}
		bytesRead, err := file.Read(buffer[:bytesToRead])
		if bytesRead > 0 {
			_, err := w.Write(buffer[:bytesRead])
			if err != nil {
				return false
			}
		}
		contentLength -= int64(bytesRead)
		return contentLength > 0 && err == nil
	})
}

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
	err = db.AutoMigrate(&Aparelho{})
	if err != nil {
		panic(err)
	}
	DATABASE = db

	router := gin.Default()
	router.POST("/create_aparelho", createAparelho)
	router.GET("/serve_ids", serveIDList)
	router.GET("/serve_image", serveImage)
	router.GET("/serve_manual", serveManual)
	router.GET("/serve_video", serveVideo)
	_ = router.Run(":8000")
}

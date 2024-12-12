package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
)

func aparelhosIDs(context *gin.Context) {
	var ids []uint

	if err := DATABASE.Model(&Aparelho{}).Pluck("ID", &ids).Error; err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}

	context.JSON(200, ids)
}

func aparelhosImage(context *gin.Context) {
	var imagemPath string

	id, err := strconv.Atoi(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = DATABASE.Model(&Aparelho{}).
		Where("id = ?", id).
		Select("ImagemPath").
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

func aparelhosManual(context *gin.Context) {
	var manualPath string

	id, err := strconv.Atoi(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = DATABASE.Model(&Manual{}).
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

func aparelhosVideo(context *gin.Context) {
	var videoPath string

	id, err := strconv.Atoi(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = DATABASE.Model(&Video{}).
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
	dsn := "host=localhost user=admin password=enzofernandes123 dbname=gethadb port=5432 sslmode=disable TimeZone=Brazil/East"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Aparelho{}, &Manual{}, &Video{})
	if err != nil {
		panic(err)
	}
	DATABASE = db

	router := gin.Default()
	router.GET("/aparelhos_ids", aparelhosIDs)
	router.GET("/aparelhos_image", aparelhosImage)
	router.GET("/manual", aparelhosManual)
	router.GET("/video", aparelhosVideo)
	_ = router.Run("192.168.15.12:8000")
}

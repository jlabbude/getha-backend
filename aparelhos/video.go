package aparelhos

import (
	"fmt"
	"getha/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"strconv"
)

func ServeVideo(context *gin.Context) {
	var videoPath string

	id, err := uuid.Parse(context.Query("id"))
	if err != nil {
		context.String(http.StatusBadRequest, "Formatação de ID inválida")
		return
	}

	err = models.DATABASE.Model(&models.Aparelhos{}).
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

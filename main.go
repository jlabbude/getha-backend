package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strconv"
)

func aparelhosImage(context *gin.Context) {
	context.File(`D:\GETHA\app\src\main\assets\logo.png`)
}

func aparelhosManual(context *gin.Context) {
	id, err := strconv.Atoi(context.Query("id"))
	if err != nil { context.String(http.StatusBadRequest, "Formatação de ID inválida") }
	context.Header("Content-Type", "application/pdf")
	var manualPath string
	switch id {
		case 1:
			manualPath = `D:\9O8UYL\POP_BrennaMoizinho.pdf`
		default:
			manualPath = `D:\9O8UYL\1768590.pdf`
	}
	context.File(manualPath)
}

func aparelhosVideo(context *gin.Context) {
	id, err := strconv.Atoi(context.Query("id"))
	if err != nil { context.String(http.StatusBadRequest, "Formatação de ID inválida") }

	var videoPath string
	switch id {
		case 1:
			videoPath =  `D:\Videos\2024-07-12 11-27-54.mp4 `
		default:
			videoPath =  `D:\9O8UYL\JIKOL\y2meta.com - Nikocado Avocado Cheers unedited HD(360p).mp4 `
	}

	file, err := os.Open(videoPath)
    if err != nil {
        context.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir o vídeo"})
        return
    }
    defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H {"error": "Erro ao fechar o file handler do vídeo"})
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
	router := gin.Default()
	router.GET("/aparelhos_image", aparelhosImage)
	router.GET("/manual", aparelhosManual)
	router.GET("/video", aparelhosVideo)
	_ = router.Run("192.168.15.12:8000")
}
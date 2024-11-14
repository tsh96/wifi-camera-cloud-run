package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	r := gin.Default()

	googleCredential, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_CREDENTIAL"))
	folderID := os.Getenv("FOLDER_ID")
	accessToken := os.Getenv("ACCESS_TOKEN")

	if err != nil {
		panic(fmt.Errorf("error decoding google credential: %v", err))
	}

	clientOption := option.WithCredentialsJSON(googleCredential)

	r.POST("photos", func(ctx *gin.Context) {
		reqAccessToken := ctx.GetHeader("Authorization")
		if reqAccessToken == "" {
			ctx.JSON(400, gin.H{"error": "missing access token"})
			return
		}
		if reqAccessToken != accessToken {
			ctx.JSON(401, gin.H{"error": "invalid access token"})
			return
		}

		fileHeader, err := ctx.FormFile("imageFile")
		if err != nil {
			ctx.JSON(400, gin.H{"error": fmt.Sprintf("error getting file: %v", err)})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("error opening file: %v", err)})
			return
		}

		driveService, err := drive.NewService(ctx, clientOption)
		if err != nil {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("error creating drive service: %v", err)})
			return
		}

		// filename = yyyy-MM-ddTHH-mm-ss.jpg
		uploadedFile, err := driveService.Files.Create(&drive.File{
			Name:    time.Now().Format("2006-01-02T15-04-05.jpg"),
			Parents: []string{folderID},
		}).Media(file).Do()
		if err != nil {
			ctx.JSON(500, gin.H{"error": fmt.Sprintf("error creating file: %v", err)})
			return
		}

		ctx.JSON(200, gin.H{"fileId": uploadedFile.Id})
	})

	// health check
	r.GET("health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	r.Run()
}

type PhotoRequest struct {
	PhotoBase64 string `json:"photoBase64"`
}

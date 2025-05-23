package repositories

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/0ero-1ne/martha-storage/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return FileRepository{
		db: db,
	}
}

func (repo FileRepository) GetFile(
	fileUUID string,
	filetype models.Filetype,
) string {
	var file models.File

	tx := repo.db.First(&file, "filename = ?", fileUUID)
	if tx.Error != nil || file.Filetype != filetype {
		return ""
	}

	return file.Filename + file.Extension
}

func (repo FileRepository) UploadFile(
	ctx *gin.Context,
	file *multipart.FileHeader,
	dirPath string,
	filetype models.Filetype,
) (string, error) {
	fileExt := filepath.Ext(file.Filename)
	newFilename := repo.generateRandomFilename()

	saveFile := &models.File{
		Filename:  newFilename,
		Extension: fileExt,
		Filetype:  filetype,
	}

	tx := repo.db.Create(&saveFile)
	if tx.Error != nil {
		return "", tx.Error
	}

	savePath := fmt.Sprintf("%s/%s%s", dirPath, newFilename, fileExt)
	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		repo.db.Delete(&models.File{}, "filename = ?", newFilename)
		return "", err
	}

	return newFilename, nil
}

func (repo FileRepository) DeleteFile(
	path string,
	fileUUID string,
	filetype models.Filetype,
) error {
	var file models.File

	tx := repo.db.First(&file, "filename = ?", fileUUID)
	if tx.Error != nil {
		return tx.Error
	}

	err := os.Remove(path + "/" + file.Filename + file.Extension)
	if err != nil {
		return err
	}

	tx = repo.db.Delete(&file)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (repo FileRepository) generateRandomFilename() string {
	return uuid.New().String()
}

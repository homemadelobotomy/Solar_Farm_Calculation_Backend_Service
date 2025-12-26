package service

import (
	"errors"
	"lab/internal/app/ds"
	"lab/internal/app/repository"
	"log"
	"math"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrNoRecords = errors.New("записи не найдены")
var ErrForbidden = errors.New("пользователь не имеет доступа к этой заявке")
var ErrBadRequest = errors.New("введены некорректные данные")
var ErrSolarPanelDeleted = errors.New("эта солнечная панель удалена")

type Service struct {
	repository  *repository.Repository
	minioClient *minio.Client
}

func NewService(repository *repository.Repository) *Service {
	minioClient, err := minio.New(os.Getenv("MINIO_HOST")+":"+os.Getenv("MINIO_PORT"), &minio.Options{
		Creds:  credentials.NewStaticV4("minio", "minio124", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &Service{
		repository:  repository,
		minioClient: minioClient,
	}
}

func formateDate(date time.Time, layout string) string {
	if date.IsZero() {
		return ""
	}
	return date.Format(layout)
}

func extractFilenameFromURL(imageURL string) string {
	parsedUrl, err := url.Parse(imageURL)
	if err != nil {
		return ""
	}
	return path.Base(parsedUrl.Path)
}

func CalculateTotalPower(panels []ds.RequestPanels, insolation float64) float64 {
	power := 0.0
	for _, panel := range panels {
		power += float64(panel.SolarPanel.Power) * panel.Area / (float64(panel.SolarPanel.Width*panel.SolarPanel.Height) / 1000000)
	}
	v := power * insolation / 1000
	return math.Round(v*100) / 100
}

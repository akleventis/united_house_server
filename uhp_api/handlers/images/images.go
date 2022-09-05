package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

type s3ObjInput struct {
	Key string `json:"key"`
}

type Handler struct {
	db        *uhp_db.UhpDB
	s3Client  *s3.S3
	s3Session *aws_session.Session
	s3Bucket  string
}

func NewHandler(db *uhp_db.UhpDB, s3Client *s3.S3, s3Session *aws_session.Session, s3Bucket string) *Handler {
	return &Handler{
		db:        db,
		s3Client:  s3Client,
		s3Session: s3Session,
		s3Bucket:  s3Bucket,
	}
}

func toMegaBytes(size int64) float64 {
	log.Info("size: ", size)
	return float64(size) / 1000000
}

// UploadImage accepts form file image and uploads it to an aws s3 bucket.
// JPEG only for memory optomization.
func (h *Handler) UploadImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get form file
		r.ParseMultipartForm(0)
		imageFile, _, err := r.FormFile("image")
		if err != nil {
			http.Error(w, lib.ErrImageFile.Error(), http.StatusBadRequest)
			return
		}
		defer imageFile.Close()

		key := r.FormValue("key")
		if key == "" {
			http.Error(w, lib.ErrFormValue.Error(), http.StatusBadRequest)
			return
		}

		var buff bytes.Buffer
		fileSize, err := buff.ReadFrom(imageFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imageFile.Seek(0, 0)

		if toMegaBytes(fileSize) > 3 {
			http.Error(w, lib.ErrImageTooLarge.Error(), http.StatusBadRequest)
			return
		}

		imageBytes, err := ioutil.ReadAll(imageFile)
		if err != nil {
			http.Error(w, lib.ErrImageFile.Error(), http.StatusBadRequest)
			return
		}

		contentType := http.DetectContentType(imageBytes)
		if contentType != "image/jpeg" {
			http.Error(w, lib.ErrFileType.Error(), http.StatusBadRequest)
			return
		}

		uploader := s3manager.NewUploader(h.s3Session)

		input := &s3manager.UploadInput{
			Bucket:      aws.String(h.s3Bucket),                  // bucket's name
			Key:         aws.String(fmt.Sprintf("%s.jpeg", key)), // files destination location
			Body:        bytes.NewReader(imageBytes),             // content of the file
			ContentType: aws.String("image/jpeg"),                // content type
		}

		output, err := uploader.UploadWithContext(context.Background(), input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lib.ApiResponse(w, http.StatusCreated, output)
	}
}

func (h *Handler) GetImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s3Obj *s3ObjInput
		if err := json.NewDecoder(r.Body).Decode(&s3Obj); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}

		obj, err := h.s3Client.GetObject(
			&s3.GetObjectInput{
				Bucket: aws.String(h.s3Bucket),
				Key:    aws.String(s3Obj.Key),
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buff := new(bytes.Buffer)
		buff.ReadFrom(obj.Body)

		lib.ApiResponse(w, http.StatusCreated, buff.Bytes())
	}
}

func (h *Handler) DeleteImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s3Obj *s3ObjInput
		if err := json.NewDecoder(r.Body).Decode(&s3Obj); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}

		obj, err := h.s3Client.DeleteObject(
			&s3.DeleteObjectInput{
				Bucket: aws.String(h.s3Bucket),
				Key:    aws.String(s3Obj.Key),
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if obj == nil {
			http.Error(w, lib.ErrNoImage.Error(), http.StatusBadRequest)
			return
		}

		lib.ApiResponse(w, http.StatusGone, "Gone")
	}
}

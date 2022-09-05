package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
)

type Handler struct {
	db        *uhp_db.UhpDB
	s3Session *aws_session.Session
}

func NewHandler(db *uhp_db.UhpDB, s3Session *aws_session.Session) *Handler {
	return &Handler{
		db:        db,
		s3Session: s3Session,
	}
}

func toMegaBytes(size int64) float64 {
	return float64(size) / 1000000
}
func (h *Handler) UploadImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		key := vars["key"]

		// get form file
		r.ParseMultipartForm(0)
		imageFile, _, err := r.FormFile("image")
		if err != nil {
			http.Error(w, lib.ErrImageFile.Error(), http.StatusBadRequest)
			return
		}
		defer imageFile.Close()

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

		uploader := s3manager.NewUploader(h.s3Session)

		imageBytes, err := ioutil.ReadAll(imageFile)
		if err != nil {
			http.Error(w, lib.ErrImageFile.Error(), http.StatusBadRequest)
			return
		}

		input := &s3manager.UploadInput{
			Bucket:      aws.String("uhp-image-upload"),         // bucket's name
			Key:         aws.String(fmt.Sprintf("%s.png", key)), // files destination location
			Body:        bytes.NewReader(imageBytes),            // content of the file
			ContentType: aws.String("image/png"),                // content type
		}

		output, err := uploader.UploadWithContext(context.Background(), input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lib.ApiResponse(w, http.StatusCreated, output)
	}
}

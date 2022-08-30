package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/gorilla/mux"
)

func (h *Handler) GetFeaturedArtists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		songs, err := h.db.GetFeaturedArtists()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, songs)
	}
}

// func (h *Handler) GetFeaturedArtist() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		if vars == nil {
// 			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
// 			return
// 		}
// 		id := vars["id"]

// 		artist, err := h.db.GetFeaturedArtist(id)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		lib.ApiResponse(w, http.StatusOK, artist)
// 	}
// }

func (h *Handler) CreateFeaturedArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var artist uhp_db.FeaturedArtist
		if err := json.NewDecoder(r.Body).Decode(&artist); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.db.CreateFeaturedArtist(artist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusCreated, res)
	}
}

func (h *Handler) UpdateFeaturedArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		id := vars["id"]

		// Get featuredartist to update
		updateFeatureArtist, err := h.db.GetFeaturedArtist(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updateFeatureArtist == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}

		// Decode body, use updateFeatureArtist object and unmarshal over to replace any fields found in req body
		if err := json.NewDecoder(r.Body).Decode(&updateFeatureArtist); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}

		// prevent client from modifying id
		intID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		updateFeatureArtist.ID = intID

		// update featured artist
		featuredArtist, err := h.db.UpdateFeaturedArtist(updateFeatureArtist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, featuredArtist)
	}
}

func (h *Handler) DeleteFeaturedArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		if err := h.db.DeleteFeaturedArtist(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusGone, "Gone")
	}
}

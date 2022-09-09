package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/gorilla/mux"
)

func (h *Handler) GetArtists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		artists, err := h.db.GetArtists()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, artists)
	}
}

func (h *Handler) GetArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		artist, err := h.db.GetArtist(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, artist)
	}
}

func (h *Handler) CreateArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var artist uhp_db.Artist
		if err := json.NewDecoder(r.Body).Decode(&artist); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.db.CreateArtist(artist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusCreated, res)
	}
}

func (h *Handler) UpdateArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusInternalServerError)
			return
		}

		id := vars["id"]

		// Get artist to update
		updateArtist, err := h.db.GetArtist(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updateArtist == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}

		// Decode body, use updateArtist object and unmarshal over to replace adn fields found in req body
		if err := json.NewDecoder(r.Body).Decode(&updateArtist); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}

		// prevent client from modifying id
		intID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		updateArtist.ID = intID

		// udpate artist
		artist, err := h.db.UpdateArtist(updateArtist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lib.ApiResponse(w, http.StatusOK, artist)
	}
}

func (h *Handler) DeleteArtist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		if err := h.db.DeleteArtist(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusGone, "Gone")
	}
}

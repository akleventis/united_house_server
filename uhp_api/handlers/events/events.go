package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/gorilla/mux"
)

type Handler struct {
	db *uhp_db.UhpDB
}

func NewHandler(db *uhp_db.UhpDB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) GetEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := h.db.GetEvents()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, events)
	}
}

func (h *Handler) GetEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		event, err := h.db.GetEvent(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, event)
	}
}

func (h *Handler) CreateEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var event uhp_db.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.db.CreateEvent(event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusCreated, res)
	}
}

// only provide updated fields
func (h *Handler) UpdateEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		id := vars["id"]

		// Get event to update
		updateEvent, err := h.db.GetEvent(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updateEvent == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}

		// Decode body, use updateEvent object and unmarshal over to replace any fields found in req body
		if err := json.NewDecoder(r.Body).Decode(&updateEvent); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		// prevent client from modifying id
		intID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		updateEvent.ID = intID

		// Update event
		event, err := h.db.UpdateEvent(updateEvent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, event)
	}
}

// DeleteEvent deletes an existing event using id
func (h *Handler) DeleteEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		if err := h.db.DeleteEvent(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusGone, "Gone")
	}
}

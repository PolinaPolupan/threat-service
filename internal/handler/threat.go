package handler

import (
	"encoding/json"
	"net/http"

	"threat-service/api"
	"threat-service/internal/service"
)

type ThreatHandler struct {
	service *service.ThreatService
}

func NewThreatHandler(service *service.ThreatService) *ThreatHandler {
	return &ThreatHandler{
		service: service,
	}
}

func (h *ThreatHandler) GetInfoByMd5(
	w http.ResponseWriter,
	r *http.Request,
	md5 string,
) {
	threat, err := h.service.Get(md5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(threat)
}

func (h *ThreatHandler) UpdateThreat(
	w http.ResponseWriter,
	r *http.Request,
	md5 string,
) {
	var request api.ThreatUpdateRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.service.Update(md5, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ThreatHandler) CreateThreat(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request api.ThreatCreateRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.service.Create(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

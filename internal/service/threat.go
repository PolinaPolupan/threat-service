package service

import (
	"time"

	"threat-service/api"
	"threat-service/internal/repository"
)

type ThreatService struct {
	repo repository.ThreatRepository
}

func NewThreatService(repo repository.ThreatRepository) *ThreatService {
	return &ThreatService{
		repo: repo,
	}
}

func (s *ThreatService) Get(md5 string) (api.Threat, error) {
	return s.repo.Get(md5)
}

func (s *ThreatService) Update(md5 string, request api.ThreatUpdateRequest) error {
	threat, err := s.repo.Get(md5)
	if err != nil {
		return err
	}

	threat.Name = request.Name

	if request.Description != nil {
		threat.Description = request.Description
	}

	if request.FileName != nil {
		threat.FileName = request.FileName
	}

	if request.FileSize != nil {
		threat.FileSize = request.FileSize
	}

	if request.FirstSeen != nil {
		threat.FirstSeen = request.FirstSeen
	}

	if request.LastSeen != nil {
		threat.LastSeen = request.LastSeen
	}

	if request.Sha256 != nil {
		threat.Sha256 = request.Sha256
	}

	if request.Source != nil {
		threat.Source = request.Source
	}

	if request.Tags != nil {
		threat.Tags = request.Tags
	}

	if request.Severity != nil {
		severity := api.ThreatSeverity(*request.Severity)
		threat.Severity = &severity
	}

	if request.Type != nil {
		threatType := api.ThreatType(*request.Type)
		threat.Type = &threatType
	}

	now := time.Now()
	threat.UpdatedAt = &now

	return s.repo.Update(md5, threat)
}

func (s *ThreatService) Create(request api.ThreatCreateRequest) error {
	threat := api.Threat{
		Md5:         request.Md5,
		Name:        request.Name,
		Description: request.Description,
		FileName:    request.FileName,
		FileSize:    request.FileSize,
		FirstSeen:   request.FirstSeen,
		LastSeen:    request.LastSeen,
		Sha256:      request.Sha256,
		Source:      request.Source,
		Tags:        request.Tags,
	}

	if request.Severity != nil {
		severity := api.ThreatSeverity(*request.Severity)
		threat.Severity = &severity
	}

	if request.Type != nil {
		threatType := api.ThreatType(*request.Type)
		threat.Type = &threatType
	}

	now := time.Now()
	threat.CreatedAt = &now
	threat.UpdatedAt = &now

	return s.repo.Create(threat)
}

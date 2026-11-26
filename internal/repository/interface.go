package repository

import "threat-service/api"

type ThreatRepository interface {
    Create(threat api.Threat) error
    Get(md5 string) (api.Threat, error)
    Update(md5 string, threat api.Threat) error
}

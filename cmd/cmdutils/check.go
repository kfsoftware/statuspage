package cmdutils

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type StatusKind string

const (
	HttpHealthCheck   StatusKind = "HttpHealthCheck"
	TLSHealthCheck    StatusKind = "TLSHealthCheck"
	StatusPageKind    StatusKind = "StatusPage"
	StatusPageUnknown StatusKind = "Unknown"
)

func GetFileType(fileBytes []byte) (StatusKind, error) {
	var initialUnmarshall struct {
		Kind StatusKind `yaml:"kind"`
	}
	err := yaml.Unmarshal(fileBytes, &initialUnmarshall)
	if err != nil {
		return StatusPageUnknown, err
	}
	log.Debugf("Kind: %s", initialUnmarshall.Kind)
	return initialUnmarshall.Kind, nil
}

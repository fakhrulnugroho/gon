package service

import (
	"fmt"
	"gon/internal/core/port/driving"
)

type versionService struct {
	version string
	os      string
	arch    string
}

func NewVersionService(version string, os string, arch string) driving.VersionService {
	return &versionService{version: version, os: os, arch: arch}
}

func (s *versionService) GetVersion() string {
	return fmt.Sprintf("gon %s (%s/%s)", s.version, s.os, s.arch)
}

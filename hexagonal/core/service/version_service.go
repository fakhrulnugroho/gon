package service

import "gon/hexagonal/core/port/driven"

type versionService struct {
	version string
}

func NewVersionService(version string) driven.VersionService {
	return &versionService{version: version}
}

func (s *versionService) GetVersion() string {
	return s.version
}

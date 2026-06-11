package model

import "gon/internal/core/domain"

type CollectionModel struct {
	Name        string      `yaml:"name,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Config      ConfigModel `yaml:"config,omitempty"`
}

func NewCollectionModelFromDomain(collection domain.Collection) *CollectionModel {
	return &CollectionModel{
		Name:        collection.Name,
		Description: collection.Description,
		Config:      *NewConfigModelFromDomain(collection.Config),
	}
}

func (m *CollectionModel) ToDomain() *domain.Collection {
	return &domain.Collection{
		Name:        m.Name,
		Description: m.Description,
		Config:      m.Config.ToDomain(),
	}
}

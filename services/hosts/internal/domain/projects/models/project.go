package models

import (
	"github.com/google/uuid"
)

type ProjectType string

const (
	TypeServer       ProjectType = "server"
	TypeVM           ProjectType = "vm"
	TypeMac          ProjectType = "mac"
	TypeShadowServer ProjectType = "shadow-server"
)

type Project struct {
	ID           string       `bson:"_id"`
	Name         string       `bson:"name"`
	Type         ProjectType  `bson:"type"`
	Tags         []string     `bson:"tags"`
	Description  string       `bson:"description"`
	CMS          CMS          `bson:"cms"`
	Network      Network      `bson:"network"`
	Deploying    Deploying    `bson:"deploying"`
	Profiling    Profiling    `bson:"profiling"`
	Notification Notification `bson:"notification"`
	Monitoring   Monitoring   `bson:"monitoring"`
	Task         Task         `bson:"task"`
	Tier         byte         `bson:"tier"`
}

// NewProject creates a new Project instance
func NewProject(id uuid.UUID, name string, projectType ProjectType) *Project {
	return &Project{
		ID:   id.String(),
		Name: name,
		Type: projectType,
		Tags: make([]string, 0),
	}
}

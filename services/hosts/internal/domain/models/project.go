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

// Project represents a project entity
type Project struct {
	ID           string              `bson:"_id"`
	Name         string              `bson:"name"`
	Type         ProjectType         `bson:"type"`
	Tags         []string            `bson:"tags"`
	Description  string              `bson:"description"`
	CMS          ProjectCMS          `bson:"cms"`
	Network     ProjectNetwork      `bson:"network"`
	Deploying    ProjectDeploying    `bson:"deploying"`
	Profiling    ProjectProfiling    `bson:"profiling"`
	Notification ProjectNotification `bson:"notification"`
	Monitoring   ProjectMonitoring   `bson:"monitoring"`
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

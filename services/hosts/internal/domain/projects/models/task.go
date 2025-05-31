package models

type Task struct {
	DeactivateWithoutCMS bool `bson:"deactivate_without_cms"`
}

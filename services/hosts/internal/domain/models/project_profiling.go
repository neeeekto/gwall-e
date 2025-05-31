package models

type ProjectProfiling struct {
	Name string   `bson:"name"`
	Tags []string `bson:"tags"`
}

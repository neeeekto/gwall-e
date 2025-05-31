package models

type Profiling struct {
	Name string   `bson:"name"`
	Tags []string `bson:"tags"`
}

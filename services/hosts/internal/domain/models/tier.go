package models

type Tier struct {
	Name      string   `bson:"name"`
	Level     int      `bson:"level"`
	Resources []string `bson:"resources"`
}
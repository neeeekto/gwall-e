package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gwall-e/services/inventory/internal/application/queries"
)

// MongoHostReadModel реализует queries.HostReadModel.
type MongoHostReadModel struct {
	collection *mongo.Collection
}

func NewMongoHostReadModel(db *mongo.Database) *MongoHostReadModel {
	return &MongoHostReadModel{
		collection: db.Collection("hosts"),
	}
}

var _ queries.HostReadModel = (*MongoHostReadModel)(nil)

func (r *MongoHostReadModel) ListHosts(ctx context.Context, query queries.ListHostsQuery) (queries.ListHostsResult, error) {
	filter := bson.M{}
	if query.ProjectID != "" {
		filter["project_id"] = query.ProjectID
	}
	if query.Kind != "" {
		filter["kind"] = kindStringToInt(query.Kind)
	}
	if len(query.Tags) > 0 {
		filter["tags"] = bson.M{"$all": query.Tags}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return queries.ListHostsResult{}, fmt.Errorf("count hosts: %w", err)
	}

	skip := int64((query.Page - 1) * query.Limit)
	limit := int64(query.Limit)

	projection := bson.M{
		"_id":        1,
		"fqdn":       1,
		"inv":        1,
		"kind":       1,
		"status":     1,
		"project_id": 1,
		"tags":       1,
		"location":   1,
		"hardware":   1,
		"created_at": 1,
		"updated_at": 1,
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetProjection(projection).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return queries.ListHostsResult{}, fmt.Errorf("find hosts: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []hostDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return queries.ListHostsResult{}, fmt.Errorf("decode hosts: %w", err)
	}

	views := make([]queries.HostView, 0, len(docs))
	for _, doc := range docs {
		views = append(views, toHostView(doc))
	}

	return queries.ListHostsResult{
		Hosts:      views,
		TotalCount: int(total),
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}

func (r *MongoHostReadModel) GetHostByID(ctx context.Context, id string) (*queries.HostView, error) {
	var doc hostDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("find host by id: %w", err)
	}
	view := toHostView(doc)
	return &view, nil
}

func toHostView(doc hostDocument) queries.HostView {
	return queries.HostView{
		ID:        doc.ID,
		FQDN:      doc.FQDN,
		Inv:       doc.Inv,
		Kind:      kindIntToString(doc.Kind),
		Status:    statusIntToString(doc.Status),
		ProjectID: doc.ProjectID,
		Tags:      doc.Tags,
		Location: queries.HostLocationView{
			Country:  doc.Location.Country,
			City:     doc.Location.City,
			Building: doc.Location.Building,
			Rack:     doc.Location.Rack,
			Unit:     doc.Location.Unit,
		},
		Hardware: queries.HostHardwareView{
			Name:     doc.Hardware.Name,
			Platform: doc.Hardware.Platform,
			CPUCount: len(doc.Hardware.MACs),
			RAMCount: 0,
		},
		CreatedAt: time.Unix(doc.CreatedAt, 0).UTC().Format(time.RFC3339),
		UpdatedAt: time.Unix(doc.UpdatedAt, 0).UTC().Format(time.RFC3339),
	}
}

func kindStringToInt(kind string) int {
	switch kind {
	case "server":
		return 1
	case "mac":
		return 2
	default:
		return 0
	}
}

func kindIntToString(kind int) string {
	switch kind {
	case 1:
		return "server"
	case 2:
		return "mac"
	default:
		return "unknown"
	}
}

func statusIntToString(status int) string {
	switch status {
	case 0:
		return "pending"
	case 1:
		return "active"
	case 2:
		return "decommissioned"
	default:
		return "unknown"
	}
}

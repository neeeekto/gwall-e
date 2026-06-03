package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// projectDocument — MongoDB-специфичная структура хранения проекта.
type projectDocument struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name"`
	Description string   `bson:"description"`
	Kind        int      `bson:"kind"`
	Status      int      `bson:"status"`
	Tags        []string `bson:"tags"`
	CreatedAt   int64    `bson:"created_at"`
	UpdatedAt   int64    `bson:"updated_at"`
}

// MongoProjectRepository реализует domain.ProjectRepository.
type MongoProjectRepository struct {
	collection *mongo.Collection
}

func NewMongoProjectRepository(db *mongo.Database) *MongoProjectRepository {
	return &MongoProjectRepository{
		collection: db.Collection("projects"),
	}
}

var _ domain.ProjectRepository = (*MongoProjectRepository)(nil)

func (r *MongoProjectRepository) Save(ctx context.Context, project *domain.Project) error {
	doc := r.toDocument(project)
	_, err := r.collection.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrProjectAlreadyExists
	}
	return err
}

func (r *MongoProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	doc := r.toDocument(project)
	filter := bson.M{"_id": string(project.ID())}
	opts := options.Replace().SetUpsert(false)
	result, err := r.collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *MongoProjectRepository) FindByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error) {
	var doc projectDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": string(id)}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toDomain(doc), nil
}

func (r *MongoProjectRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"name": name})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoProjectRepository) toDocument(p *domain.Project) projectDocument {
	return projectDocument{
		ID:          string(p.ID()),
		Name:        p.Name(),
		Description: p.Description(),
		Kind:        int(p.Kind()),
		Status:      int(p.Status()),
		Tags:        p.Tags(),
		CreatedAt:   p.CreatedAt().Unix(),
		UpdatedAt:   p.UpdatedAt().Unix(),
	}
}

func (r *MongoProjectRepository) toDomain(doc projectDocument) *domain.Project {
	return domain.RestoreProject(domain.RestoreProjectParams{
		ID:          domain.ProjectID(doc.ID),
		Name:        doc.Name,
		Description: doc.Description,
		Kind:        domain.ProjectKind(doc.Kind),
		Status:      domain.ProjectStatus(doc.Status),
		Tags:        doc.Tags,
		CreatedAt:   time.Unix(doc.CreatedAt, 0).UTC(),
		UpdatedAt:   time.Unix(doc.UpdatedAt, 0).UTC(),
	})
}

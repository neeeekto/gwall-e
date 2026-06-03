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

// hostDocument — MongoDB-специфичная структура хранения.
type hostDocument struct {
	ID        string          `bson:"_id"`
	ProjectID string          `bson:"project_id"`
	FQDN      string          `bson:"fqdn"`
	Inv       int             `bson:"inv"`
	Kind      int             `bson:"kind"`
	Status    int             `bson:"status"`
	Tags      []string        `bson:"tags"`
	Location  locationDoc     `bson:"location"`
	Hardware  hostHardwareDoc `bson:"hardware"`
	CreatedAt int64           `bson:"created_at"`
	UpdatedAt int64           `bson:"updated_at"`
}

type locationDoc struct {
	Country  string `bson:"country"`
	City     string `bson:"city"`
	Building string `bson:"building"`
	Module   string `bson:"module"`
	Rack     string `bson:"rack"`
	Unit     string `bson:"unit"`
	Object   string `bson:"object"`
	RoomType string `bson:"room_type"`
}

type hostHardwareDoc struct {
	Name        string   `bson:"name"`
	Platform    string   `bson:"platform"`
	IPMIMac     string   `bson:"ipmi_mac"`
	Motherboard string   `bson:"motherboard"`
	MACs        []string `bson:"macs"`
}

// MongoHostRepository реализует domain.HostRepository.
type MongoHostRepository struct {
	collection *mongo.Collection
}

func NewMongoHostRepository(db *mongo.Database) *MongoHostRepository {
	return &MongoHostRepository{
		collection: db.Collection("hosts"),
	}
}

var _ domain.HostRepository = (*MongoHostRepository)(nil)

func (r *MongoHostRepository) Save(ctx context.Context, host *domain.Host) error {
	doc := r.toDocument(host)
	_, err := r.collection.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrHostAlreadyExists
	}
	return err
}

func (r *MongoHostRepository) Update(ctx context.Context, host *domain.Host) error {
	doc := r.toDocument(host)
	filter := bson.M{"_id": string(host.ID())}
	opts := options.Replace().SetUpsert(false)
	result, err := r.collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrHostNotFound
	}
	return nil
}

func (r *MongoHostRepository) FindByID(ctx context.Context, id domain.HostID) (*domain.Host, error) {
	var doc hostDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": string(id)}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrHostNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toDomain(doc), nil
}

func (r *MongoHostRepository) ExistsByFQDN(ctx context.Context, fqdn string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"fqdn": fqdn})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoHostRepository) Delete(ctx context.Context, id domain.HostID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": string(id)})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrHostNotFound
	}
	return nil
}

func (r *MongoHostRepository) toDocument(host *domain.Host) hostDocument {
	hw := host.Hardware()
	loc := host.Location()
	return hostDocument{
		ID:        string(host.ID()),
		ProjectID: string(host.ProjectID()),
		FQDN:      host.FQDN(),
		Inv:       host.Inv(),
		Kind:      int(host.Kind()),
		Status:    int(host.Status()),
		Tags:      host.Tags(),
		CreatedAt: host.CreatedAt().Unix(),
		UpdatedAt: host.UpdatedAt().Unix(),
		Location: locationDoc{
			Country:  loc.Country,
			City:     loc.City,
			Building: loc.Building,
			Module:   loc.Module,
			Rack:     loc.Rack,
			Unit:     loc.Unit,
			Object:   loc.Object,
			RoomType: loc.RoomType,
		},
		Hardware: hostHardwareDoc{
			Name:        hw.Name,
			Platform:    hw.Platform,
			IPMIMac:     hw.IPMIMac,
			Motherboard: hw.Motherboard,
			MACs:        hw.MACs,
		},
	}
}

func (r *MongoHostRepository) toDomain(doc hostDocument) *domain.Host {
	return domain.RestoreHost(domain.RestoreHostParams{
		ID:        domain.HostID(doc.ID),
		ProjectID: domain.ProjectID(doc.ProjectID),
		FQDN:      doc.FQDN,
		Inv:       doc.Inv,
		Kind:      domain.HostKind(doc.Kind),
		Status:    domain.HostStatus(doc.Status),
		Tags:      doc.Tags,
		Location: domain.Location{
			Country:  doc.Location.Country,
			City:     doc.Location.City,
			Building: doc.Location.Building,
			Module:   doc.Location.Module,
			Rack:     doc.Location.Rack,
			Unit:     doc.Location.Unit,
			Object:   doc.Location.Object,
			RoomType: doc.Location.RoomType,
		},
		Hardware: domain.HostHardware{
			Name:        doc.Hardware.Name,
			Platform:    doc.Hardware.Platform,
			IPMIMac:     doc.Hardware.IPMIMac,
			Motherboard: doc.Hardware.Motherboard,
			MACs:        doc.Hardware.MACs,
		},
		CreatedAt: time.Unix(doc.CreatedAt, 0).UTC(),
		UpdatedAt: time.Unix(doc.UpdatedAt, 0).UTC(),
	})
}

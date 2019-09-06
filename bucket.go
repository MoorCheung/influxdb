package influxdb

import (
	"context"
	"strings"
	"time"
)

// TasksSystemBucketID and MonitoringSystemBucketID are IDs that are reserved for system buckets.
// If any system bucket IDs are added, Bucket.IsSystem must be updated to include them.
const (
	// TasksSystemBucketID is the fixed ID for our tasks system bucket
	TasksSystemBucketID = ID(10)
	// MonitoringSystemBucketID is the fixed ID for our monitoring system bucket
	MonitoringSystemBucketID = ID(11)
)

// InfiniteRetention is default infinite retention period.
const InfiniteRetention = 0

// Bucket is a bucket. 🎉
type Bucket struct {
	ID                  ID            `json:"id,omitempty"`
	OrgID               ID            `json:"orgID,omitempty"`
	Name                string        `json:"name"`
	Description         string        `json:"description"`
	RetentionPolicyName string        `json:"rp,omitempty"` // This to support v1 sources
	RetentionPeriod     time.Duration `json:"retentionPeriod"`
	CRUDLog
}

// TODO(jade): move this logic to a type set directly on Bucket.
// IsSystem returns true if a bucket is a known system bucket
func (b *Bucket) IsSystem() bool {
	return b.ID == TasksSystemBucketID || b.ID == MonitoringSystemBucketID
}

// ops for buckets error and buckets op logs.
var (
	OpFindBucketByID = "FindBucketByID"
	OpFindBucket     = "FindBucket"
	OpFindBuckets    = "FindBuckets"
	OpCreateBucket   = "CreateBucket"
	OpUpdateBucket   = "UpdateBucket"
	OpDeleteBucket   = "DeleteBucket"
)

// BucketService represents a service for managing bucket data.
type BucketService interface {
	// FindBucketByID returns a single bucket by ID.
	FindBucketByID(ctx context.Context, id ID) (*Bucket, error)

	// FindBucket returns the first bucket that matches filter.
	FindBucket(ctx context.Context, filter BucketFilter) (*Bucket, error)

	// FindBuckets returns a list of buckets that match filter and the total count of matching buckets.
	// Additional options provide pagination & sorting.
	FindBuckets(ctx context.Context, filter BucketFilter, opt ...FindOptions) ([]*Bucket, int, error)

	// CreateBucket creates a new bucket and sets b.ID with the new identifier.
	CreateBucket(ctx context.Context, b *Bucket) error

	// UpdateBucket updates a single bucket with changeset.
	// Returns the new bucket state after update.
	UpdateBucket(ctx context.Context, id ID, upd BucketUpdate) (*Bucket, error)

	// DeleteBucket removes a bucket by ID.
	DeleteBucket(ctx context.Context, id ID) error
}

// BucketUpdate represents updates to a bucket.
// Only fields which are set are updated.
type BucketUpdate struct {
	Name            *string        `json:"name,omitempty"`
	Description     *string        `json:"description,omitempty"`
	RetentionPeriod *time.Duration `json:"retentionPeriod,omitempty"`
}

// BucketFilter represents a set of filter that restrict the returned results.
type BucketFilter struct {
	ID             *ID
	Name           *string
	OrganizationID *ID
	Org            *string
}

// QueryParams Converts BucketFilter fields to url query params.
func (f BucketFilter) QueryParams() map[string][]string {
	qp := map[string][]string{}
	if f.ID != nil {
		qp["id"] = []string{f.ID.String()}
	}

	if f.Name != nil {
		qp["name"] = []string{*f.Name}
	}

	if f.OrganizationID != nil {
		qp["orgID"] = []string{f.OrganizationID.String()}
	}

	if f.Org != nil {
		qp["org"] = []string{*f.Org}
	}

	return qp
}

// String returns a human-readable string of the BucketFilter,
// particularly useful for error messages.
func (f BucketFilter) String() string {
	// There should always be exactly 2 fields set, but if it's somehow more, that's fine.
	parts := make([]string, 0, 2)
	if f.ID != nil {
		parts = append(parts, "Bucket ID: "+f.ID.String())
	}
	if f.Name != nil {
		parts = append(parts, "Bucket Name: "+*f.Name)
	}
	if f.OrganizationID != nil {
		parts = append(parts, "Org ID: "+f.OrganizationID.String())
	}
	if f.Org != nil {
		parts = append(parts, "Org Name: "+*f.Org)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
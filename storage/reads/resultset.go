package reads

import (
	"context"

	"github.com/influxdata/influxdb/v2/models"
	"github.com/influxdata/influxdb/v2/storage/reads/datatypes"
	"github.com/influxdata/influxdb/v2/tsdb/cursors"
)

type resultSet struct {
	ctx          context.Context
	seriesCursor SeriesCursor
	seriesRow    *SeriesRow
	arrayCursors *arrayCursors
	cursor       cursors.Cursor
}

func NewFilteredResultSet(ctx context.Context, req *datatypes.ReadFilterRequest, seriesCursor SeriesCursor) ResultSet {
	return &resultSet{
		ctx:          ctx,
		seriesCursor: seriesCursor,
		arrayCursors: newArrayCursors(ctx, req.Range.Start, req.Range.End, true),
	}
}

func (r *resultSet) Err() error { return nil }

// Close closes the result set. Close is idempotent.
func (r *resultSet) Close() {
	if r == nil {
		return // Nothing to do.
	}
	r.seriesRow = nil
	r.seriesCursor.Close()
}

// Next returns true if there are more results available.
func (r *resultSet) Next() bool {
	if r == nil {
		return false
	}

	r.seriesRow = r.seriesCursor.Next()
	if r.seriesRow == nil {
		return false
	}

	r.cursor = r.arrayCursors.createCursor(*r.seriesRow)
	return true
}

func (r *resultSet) Cursor() cursors.Cursor {
	return r.cursor
}

func (r *resultSet) Tags() models.Tags {
	return r.seriesRow.Tags
}

// Stats returns the stats for the underlying cursors.
// Available after resultset has been scanned.
func (r *resultSet) Stats() cursors.CursorStats {
	if r.seriesRow.Query == nil {
		return cursors.CursorStats{}
	}
	return r.seriesRow.Query.Stats()
}

package postgres

import (
	"context"
	"time"

	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	"gct/internal/kernel/consts"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableDataExports

var writeColumns = []string{
	"id", "type", "status", "file_url",
	"created_by", "created_at", "completed_at",
}

// DataExportWriteRepo implements exportentity.DataExportRepository using PostgreSQL.
type DataExportWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewDataExportWriteRepo creates a new DataExportWriteRepo.
func NewDataExportWriteRepo(pool *pgxpool.Pool) *DataExportWriteRepo {
	return &DataExportWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new DataExport aggregate into the database.
func (r *DataExportWriteRepo) Save(ctx context.Context, q shareddomain.Querier, de *exportentity.DataExport) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "DataExportWriteRepo.Save")
	defer func() { end(err) }()

	fileURL := ""
	if de.FileURL() != nil {
		fileURL = *de.FileURL()
	}

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			de.ID(),
			de.DataType(),
			de.Status(),
			fileURL,
			de.UserID(),
			de.CreatedAt(),
			nil,
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Update updates an existing DataExport aggregate in the database.
func (r *DataExportWriteRepo) Update(ctx context.Context, q shareddomain.Querier, de *exportentity.DataExport) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "DataExportWriteRepo.Update")
	defer func() { end(err) }()

	updateFileURL := ""
	if de.FileURL() != nil {
		updateFileURL = *de.FileURL()
	}

	sql, args, err := r.builder.
		Update(tableName).
		Set("status", de.Status()).
		Set("file_url", updateFileURL).
		Where(squirrel.Eq{"id": de.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves a DataExport aggregate by its ID.
func (r *DataExportWriteRepo) FindByID(ctx context.Context, id exportentity.DataExportID) (result *exportentity.DataExport, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "DataExportWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanDataExport(row)
}

// Delete removes a data export by its ID.
func (r *DataExportWriteRepo) Delete(ctx context.Context, q shareddomain.Querier, id exportentity.DataExportID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "DataExportWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

func scanDataExport(row pgx.Row) (*exportentity.DataExport, error) {
	var (
		id          uuid.UUID
		dataType    string
		status      string
		fileURL     string
		createdBy   *uuid.UUID
		createdAt   time.Time
		completedAt *time.Time
	)

	err := row.Scan(&id, &dataType, &status, &fileURL, &createdBy, &createdAt, &completedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	_ = completedAt

	userID := uuid.Nil
	if createdBy != nil {
		userID = *createdBy
	}

	var fileURLPtr *string
	if fileURL != "" {
		fileURLPtr = &fileURL
	}

	return exportentity.ReconstructDataExport(id, createdAt, createdAt, userID, dataType, "", status, fileURLPtr, nil), nil
}

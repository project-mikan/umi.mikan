package database

// Code generated by xo. DO NOT EDIT.

import (
	"context"

	"github.com/google/uuid"
)

// EntityCategory represents a row from 'public.entity_categories'.
type EntityCategory struct {
	ID        uuid.UUID `json:"id"`         // id
	Name      string    `json:"name"`       // name
	CreatedAt int64     `json:"created_at"` // created_at
	UpdatedAt int64     `json:"updated_at"` // updated_at
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [EntityCategory] exists in the database.
func (ec *EntityCategory) Exists() bool {
	return ec._exists
}

// Deleted returns true when the [EntityCategory] has been marked for deletion
// from the database.
func (ec *EntityCategory) Deleted() bool {
	return ec._deleted
}

// Insert inserts the [EntityCategory] to the database.
func (ec *EntityCategory) Insert(ctx context.Context, db DB) error {
	switch {
	case ec._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case ec._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (manual)
	const sqlstr = `INSERT INTO public.entity_categories (` +
		`id, name, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4` +
		`)`
	// run
	logf(sqlstr, ec.ID, ec.Name, ec.CreatedAt, ec.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, ec.ID, ec.Name, ec.CreatedAt, ec.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	ec._exists = true
	return nil
}

// Update updates a [EntityCategory] in the database.
func (ec *EntityCategory) Update(ctx context.Context, db DB) error {
	switch {
	case !ec._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case ec._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.entity_categories SET ` +
		`name = $1, created_at = $2, updated_at = $3 ` +
		`WHERE id = $4`
	// run
	logf(sqlstr, ec.Name, ec.CreatedAt, ec.UpdatedAt, ec.ID)
	if _, err := db.ExecContext(ctx, sqlstr, ec.Name, ec.CreatedAt, ec.UpdatedAt, ec.ID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [EntityCategory] to the database.
func (ec *EntityCategory) Save(ctx context.Context, db DB) error {
	if ec.Exists() {
		return ec.Update(ctx, db)
	}
	return ec.Insert(ctx, db)
}

// Upsert performs an upsert for [EntityCategory].
func (ec *EntityCategory) Upsert(ctx context.Context, db DB) error {
	switch {
	case ec._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO public.entity_categories (` +
		`id, name, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`name = EXCLUDED.name, created_at = EXCLUDED.created_at, updated_at = EXCLUDED.updated_at `
	// run
	logf(sqlstr, ec.ID, ec.Name, ec.CreatedAt, ec.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, ec.ID, ec.Name, ec.CreatedAt, ec.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	ec._exists = true
	return nil
}

// Delete deletes the [EntityCategory] from the database.
func (ec *EntityCategory) Delete(ctx context.Context, db DB) error {
	switch {
	case !ec._exists: // doesn't exist
		return nil
	case ec._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.entity_categories ` +
		`WHERE id = $1`
	// run
	logf(sqlstr, ec.ID)
	if _, err := db.ExecContext(ctx, sqlstr, ec.ID); err != nil {
		return logerror(err)
	}
	// set deleted
	ec._deleted = true
	return nil
}

// EntityCategoryByID retrieves a row from 'public.entity_categories' as a [EntityCategory].
//
// Generated from index 'entity_categories_pkey'.
func EntityCategoryByID(ctx context.Context, db DB, id uuid.UUID) (*EntityCategory, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, created_at, updated_at ` +
		`FROM public.entity_categories ` +
		`WHERE id = $1`
	// run
	logf(sqlstr, id)
	ec := EntityCategory{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, id).Scan(&ec.ID, &ec.Name, &ec.CreatedAt, &ec.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &ec, nil
}

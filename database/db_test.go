package database

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mopemope/quicknews/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTx_Success(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create an ent client
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	defer func() { _ = client.Close() }()

	// Set up expectations for the transaction
	mock.ExpectBegin()
	mock.ExpectCommit()

	// Call WithTx with a function that succeeds
	err = WithTx(context.Background(), client, func(tx *ent.Tx) error {
		// Simulate some work in the transaction
		return nil
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RollbackOnError(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create an ent client
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	defer func() { _ = client.Close() }()

	// Set up expectations for the transaction with rollback
	mock.ExpectBegin()
	mock.ExpectRollback()

	// Call WithTx with a function that returns an error
	expectedErr := assert.AnError
	err = WithTx(context.Background(), client, func(tx *ent.Tx) error {
		return expectedErr
	})

	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create an ent client
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	defer func() { _ = client.Close() }()

	// Set up expectations for the transaction with rollback due to panic
	mock.ExpectBegin()
	mock.ExpectRollback()

	// The WithTx function should handle the panic internally and not return an error
	// The panic is recovered and the transaction is rolled back
	err = WithTx(context.Background(), client, func(tx *ent.Tx) error {
		panic("test panic")
	})

	// The function should have handled the panic internally and not return an error
	// since the panic is recovered in the defer function
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnErrOnTargetSameAsCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, 2, 2),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 2, 2))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, 1, 1),
		fmt.Sprintf(ErrFmtMigrateAlreadyOnTargetVersion, 1, 1))
}

func TestShouldReturnErrOnUpMigrationTargetVersionLessTHanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, 0, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 0, testLatestVersion))

	assert.NoError(t,
		schemaMigrateChecks(providerPostgres, true, testLatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, 0, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 0, testLatestVersion))

	assert.NoError(t,
		schemaMigrateChecks(providerSQLite, true, testLatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, 0, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetLessThanCurrent, 0, testLatestVersion))

	assert.NoError(t,
		schemaMigrateChecks(providerMySQL, true, testLatestVersion, 0))
}

func TestMigrationUpShouldReturnErrOnAlreadyLatest(t *testing.T) {
	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest, testLatestVersion))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest, testLatestVersion))

	assert.Equal(t,
		ErrSchemaAlreadyUpToDate,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest, testLatestVersion))
}

func TestShouldReturnErrOnVersionDoesntExits(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest-1, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest-1, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest-1, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateUpTargetGreaterThanLatest, SchemaLatest-1, testLatestVersion))
}

func TestMigrationDownShouldReturnErrOnTargetLessThanPre1(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, -4, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -4))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, -2, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, -2, testLatestVersion),
		fmt.Sprintf(ErrFmtMigrateDownTargetLessThanMinimum, -2))

	assert.NoError(t,
		schemaMigrateChecks(providerPostgres, false, -1, testLatestVersion))
}

func TestMigrationDownShouldReturnErrOnTargetVersionGreaterThanCurrent(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, false, testLatestVersion, 0),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, testLatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, false, testLatestVersion, 0),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, testLatestVersion, 0))

	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, false, testLatestVersion, 0),
		fmt.Sprintf(ErrFmtMigrateDownTargetGreaterThanCurrent, testLatestVersion, 0))
}

func TestShouldReturnErrWhenCurrentIsGreaterThanLatest(t *testing.T) {
	assert.EqualError(t,
		schemaMigrateChecks(providerPostgres, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerMySQL, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, testLatestVersion))

	assert.EqualError(t,
		schemaMigrateChecks(providerSQLite, true, SchemaLatest-4, SchemaLatest-5),
		fmt.Sprintf(errFmtSchemaCurrentGreaterThanLatestKnown, testLatestVersion))
}

func TestSchemaVersionToString(t *testing.T) {
	assert.Equal(t, "unknown", SchemaVersionToString(-2))
	assert.Equal(t, "pre1", SchemaVersionToString(-1))
	assert.Equal(t, "N/A", SchemaVersionToString(0))
	assert.Equal(t, "1", SchemaVersionToString(1))
	assert.Equal(t, "2", SchemaVersionToString(2))
}

package storage

import (
	"embed"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/authelia/authelia/v4/internal/model"
)

//go:embed migrations/*
var migrationsFS embed.FS

func latestMigrationVersion(providerName string) (version int, err error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return -1, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		m, err := scanMigration(entry.Name())
		if err != nil {
			return -1, err
		}

		if m.Provider != providerName {
			continue
		}

		if !m.Up {
			continue
		}

		if m.Version > version {
			version = m.Version
		}
	}

	return version, nil
}

func loadMigration(providerName string, version int, up bool) (migration *model.SchemaMigration, err error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		m, err := scanMigration(entry.Name())
		if err != nil {
			return nil, err
		}

		migration = &m

		if up != migration.Up {
			continue
		}

		if migration.Provider != providerAll && migration.Provider != providerName {
			continue
		}

		if version != migration.Version {
			continue
		}

		return migration, nil
	}

	return nil, errors.New("migration not found")
}

// loadMigrations scans the migrations fs and loads the appropriate migrations for a given providerName, prior and
// target versions. If the target version is -1 this indicates the latest version. If the target version is 0
// this indicates the database zero state.
func loadMigrations(providerName string, prior, target int) (migrations []model.SchemaMigration, err error) {
	if prior == target && (prior != -1 || target != -1) {
		return nil, ErrMigrateCurrentVersionSameAsTarget
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	up := prior < target

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		migration, err := scanMigration(entry.Name())
		if err != nil {
			return nil, err
		}

		if skipMigration(providerName, up, target, prior, &migration) {
			continue
		}

		migrations = append(migrations, migration)
	}

	if up {
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Version < migrations[j].Version
		})
	} else {
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Version > migrations[j].Version
		})
	}

	return migrations, nil
}

func skipMigration(providerName string, up bool, target, prior int, migration *model.SchemaMigration) (skip bool) {
	if migration.Provider != providerAll && migration.Provider != providerName {
		// Skip if migration.Provider is not a match.
		return true
	}

	if up {
		if !migration.Up {
			// Skip if we wanted an Up migration but it isn't an Up migration.
			return true
		}

		if target != -1 && (migration.Version > target || migration.Version <= prior) {
			// Skip if the migration version is greater than the target or less than or equal to the previous version.
			return true
		}
	} else {
		if migration.Up {
			// Skip if we didn't want an Up migration but it is an Up migration.
			return true
		}

		if migration.Version == 1 && target == -1 {
			// Skip if we're targeting pre1 and the migration version is 1 as this migration will destroy all data
			// preventing a successful migration.
			return true
		}

		if migration.Version <= target || migration.Version > prior {
			// Skip the migration if we want to go down and the migration version is less than or equal to the target
			// or greater than the previous version.
			return true
		}
	}

	return false
}

func scanMigration(m string) (migration model.SchemaMigration, err error) {
	result := reMigration.FindStringSubmatch(m)

	if result == nil || len(result) != 5 {
		return model.SchemaMigration{}, errors.New("invalid migration: could not parse the format")
	}

	migration = model.SchemaMigration{
		Name:     strings.ReplaceAll(result[2], "_", " "),
		Provider: result[3],
	}

	data, err := migrationsFS.ReadFile(fmt.Sprintf("migrations/%s", m))
	if err != nil {
		return model.SchemaMigration{}, err
	}

	migration.Query = string(data)

	switch result[4] {
	case "up":
		migration.Up = true
	case "down":
		migration.Up = false
	default:
		return model.SchemaMigration{}, fmt.Errorf("invalid migration: value in position 4 '%s' must be up or down", result[4])
	}

	migration.Version, _ = strconv.Atoi(result[1])

	switch migration.Provider {
	case providerAll, providerSQLite, providerMySQL, providerPostgres:
		break
	default:
		return model.SchemaMigration{}, fmt.Errorf("invalid migration: value in position 3 '%s' must be all, sqlite, postgres, or mysql", result[3])
	}

	return migration, nil
}

### Migration

The migration package provide tooling to perform DB schema and data migration.
Each migration is a directory follows the naming convention of v**N** where N=version.
Each contains a migration.go and model _package_ containing the models grouped into .go files.
The migration.go contains logic for both schema and data migration not handled by the GORM
schema migrator.

**Note: Migrations may NEVER be deleted after they have been released.**

#### Building a new migration.

To create a new migration run: `$ make migration`.
### Migration

The migration package provide tooling to perform DB schema and data migration.
Each migration is a directory follows the naming convention of v**N** where N=version.
Each contains a migration.go and model _package_ containing the models grouped into .go files.
The migration.go contains logic for both schema and data migration not handled by the GORM
schema migrator. Migrations are executed in order.

**Note: Migrations may NEVER be deleted after they have been released.** This is to support
upgrades that skip releases. For example, upgrade from Tackle v0.3 directly to v0.7
requires the intermediate migrations to be executed.

#### Building a new migration.

To create a new migration run:
```
$ make migration`.
```

#### Patch file.

Within each migration/_vN_/model there is a `mod.patch` file used to document model changes.
The file is genertaed using diff.
Example:

```
$ diff -ruN --exclude=mod.patch migration/v15/model migration/v16/model > migration/v16/model/mod.patch
```

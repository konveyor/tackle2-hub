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

#### Schema-Driven Documents

The Schema is:
- name:
- domain:
- variant:
- subject:
- []versions:
    - definition
    - migration

The domain, variant and subject provide a hierarchical reference used a locator.  
API users (the UI) can query for a schema based on these fields.  Endpoint: `/schema/:domain/:variant/:subject`.  
As an example, for schema driven fields in the inventory, here is an example for platform specific fields:  `/schema/platform/cf/coordinates`.  The variant `cf` is the Platform.Kind.

The hub will perform:
- schema/data migration.  The _patch_ field contains a yq ([yaml-query](https://mikefarah.gitbook.io/yq)) expression used to migrate (patch) the document structure and content. This will tie into the existing _migration_ mechanism for the data model.
- document schema validation. The current version is stored in the Setting table. The key has format: .jsd._name_.version and a value of:
- digest - md5
- version - number (zero-based index).

Exmaple:
```
.jsd.cloudfoundry-coordinates.version|{"digest":"1F264F2DA07BBC8E","version":1}
```

Objects in the inventory may have fields containing _schema-driven_ json documents.  
For example the Application may have a platform (type) specific field. For example, the application's 
coordinates in the source platform:
Application:
- coordinates Document

Where Document is defined as:

Document:
- schema:   // _schema name_
- content:  // _document content_

---

Examples using `Person` because it's relatable.

Person (v1):
```yaml
name: Elmer Fudd
age: 44
phone: 222-333-4444
```

Person (v2):
```yaml
name: Elmer Fudd
age: 44
phone:
  npa: 222
  nxx: 333
  number: 4444
```

Schema (json-schema):
```yaml
---
kind: Schema
apiVersion: tackle.konveyor.io/v1alpha1
metadata:
  name: people
  namespace: konveyor-tackle
spec:
  domain: people
  variant: manager
  subject: basic
  versions:
  - definition:
        '$schema': https://json-schema.org/draft/2020-12/schema
        title: Person
        type: object
        required:
          - name
          - age
          - phone
        properties:
          name:
            type: string
          age:
            type: integer
            minimum: 0
          phone:
            type: string
            pattern: '^\d{3}-\d{3}-\d{4}$'
            description: Phone number in the format 555-444-8888
  - definition:
        '$schema': https://json-schema.org/draft/2020-12/schema
        title: Person
        type: object
        required:
          - name
          - age
          - phone
        properties:
          name:
            type: string
          age:
            type: integer
            minimum: 0
          phone:
            type: object
            required:
              - npa
              - nxx
              - number
            properties:
              npa:
                type: string
                pattern: '^\d{3}$'
                description: 3-digit area code
              nxx:
                type: string
                pattern: '^\d{3}$'
                description: 3-digit exchange code
              number:
                type: string
                pattern: '^\d{4}$'
                description: 4-digit line number
    migration: >
      .phone |= (split("-") 
        | {
            "npa": .[0],
            "nxx": .[1],
            "number": .[2]
          })
```
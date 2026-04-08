# Role Permissions

This document lists what each role may do.  
Verb meanings (CRUD mapping):

- **get** → **Read**
- **post** → **Create**
- **put** / **patch** → **Update**
- **delete** → **Delete**

---

## 🛡 Role: **tackle-admin**
Full administrative access to nearly all resources — can create, read, update, and delete most entities.

| Resource                  | Create | Read | Update | Delete |
|---------------------------|--------|------|--------|--------|
| addons                    | ✅ | ✅ | ✅ | ✅ |
| adoptionplans             | ✅ | ❌ | ❌ | ❌ |
| applications              | ✅ | ✅ | ✅ | ✅ |
| applications.facts        | ✅ | ✅ | ✅ | ✅ |
| applications.tags         | ✅ | ✅ | ✅ | ✅ |
| applications.bucket       | ✅ | ✅ | ✅ | ✅ |
| applications.analyses     | ✅ | ✅ | ✅ | ✅ |
| applications.manifests    | ✅ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.assessments  | ✅ | ✅ | ❌ | ❌ |
| assessments               | ✅ | ✅ | ✅ | ✅ |
| businessservices          | ✅ | ✅ | ✅ | ✅ |
| dependencies              | ✅ | ✅ | ✅ | ✅ |
| generators                | ✅ | ✅ | ✅ | ✅ |
| identities                | ✅ | ✅ | ✅ | ✅ |
| imports                   | ✅ | ✅ | ✅ | ✅ |
| jobfunctions              | ✅ | ✅ | ✅ | ✅ |
| kai                       | ✅ | ✅ | ❌ | ❌ |
| manifests                 | ✅ | ✅ | ✅ | ✅ |
| migrationwaves            | ✅ | ✅ | ✅ | ✅ |
| platforms                 | ✅ | ✅ | ✅ | ✅ |
| proxies                   | ✅ | ✅ | ✅ | ✅ |
| reviews                   | ✅ | ✅ | ✅ | ✅ |
| schemas                   | ✅ | ✅ | ✅ | ✅ |
| settings                  | ✅ | ✅ | ✅ | ✅ |
| stakeholdergroups         | ✅ | ✅ | ✅ | ✅ |
| stakeholders              | ✅ | ✅ | ✅ | ✅ |
| tags                      | ✅ | ✅ | ✅ | ✅ |
| tagcategories             | ✅ | ✅ | ✅ | ✅ |
| tasks                     | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket              | ✅ | ✅ | ✅ | ✅ |
| trackers                  | ✅ | ✅ | ✅ | ✅ |
| tickets                   | ✅ | ✅ | ✅ | ✅ |
| cache                     | ❌ | ✅ | ❌ | ✅ |
| files                     | ✅ | ✅ | ✅ | ✅ |
| buckets                   | ✅ | ✅ | ✅ | ✅ |
| rulesets                  | ✅ | ✅ | ✅ | ✅ |
| targets                   | ✅ | ✅ | ✅ | ✅ |
| analyses                  | ✅ | ✅ | ✅ | ✅ |
| archetypes                | ✅ | ✅ | ✅ | ✅ |
| archetypes.assessments    | ✅ | ✅ | ❌ | ❌ |
| questionnaires            | ✅ | ✅ | ✅ | ✅ |

---

## 🛠 Role: **architect**
Broad create/update/delete rights but restricted on identities, proxies, settings, and trackers.

| Resource                  | Create | Read | Update | Delete |
|---------------------------|--------|------|--------|--------|
| addons                    | ✅ | ✅ | ✅ | ✅ |
| adoptionplans             | ✅ | ❌ | ❌ | ❌ |
| applications              | ✅ | ✅ | ✅ | ✅ |
| applications.facts        | ✅ | ✅ | ✅ | ✅ |
| applications.tags         | ✅ | ✅ | ✅ | ✅ |
| applications.bucket       | ✅ | ✅ | ✅ | ✅ |
| applications.analyses     | ✅ | ✅ | ✅ | ✅ |
| applications.manifests    | ✅ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.assessments  | ✅ | ✅ | ❌ | ❌ |
| assessments               | ✅ | ✅ | ✅ | ✅ |
| businessservices          | ✅ | ✅ | ✅ | ✅ |
| dependencies              | ✅ | ✅ | ✅ | ✅ |
| generators                | ✅ | ✅ | ✅ | ✅ |
| identities                | ❌ | ✅ | ❌ | ❌ |
| imports                   | ✅ | ✅ | ✅ | ✅ |
| jobfunctions              | ✅ | ✅ | ✅ | ✅ |
| kai                       | ✅ | ✅ | ❌ | ❌ |
| manifests                 | ✅ | ✅ | ✅ | ✅ |
| migrationwaves            | ✅ | ✅ | ✅ | ✅ |
| platforms                 | ✅ | ✅ | ✅ | ✅ |
| proxies                   | ❌ | ✅ | ❌ | ❌ |
| reviews                   | ✅ | ✅ | ✅ | ✅ |
| schemas                   | ❌ | ✅ | ❌ | ❌ |
| settings                  | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups         | ✅ | ✅ | ✅ | ✅ |
| stakeholders              | ✅ | ✅ | ✅ | ✅ |
| tags                      | ✅ | ✅ | ✅ | ✅ |
| tagcategories             | ✅ | ✅ | ✅ | ✅ |
| tasks                     | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket              | ✅ | ✅ | ✅ | ✅ |
| trackers                  | ❌ | ✅ | ❌ | ❌ |
| tickets                   | ✅ | ✅ | ✅ | ✅ |
| cache                     | ❌ | ✅ | ❌ | ❌ |
| files                     | ✅ | ✅ | ✅ | ✅ |
| buckets                   | ✅ | ✅ | ✅ | ✅ |
| rulesets                  | ✅ | ✅ | ✅ | ✅ |
| targets                   | ✅ | ✅ | ✅ | ✅ |
| analyses                  | ✅ | ✅ | ✅ | ✅ |
| archetypes                | ✅ | ✅ | ✅ | ✅ |
| archetypes.assessments    | ✅ | ✅ | ❌ | ❌ |
| questionnaires            | ❌ | ✅ | ❌ | ❌ |

---

## 🚚 Role: **migrator**
Mostly read-only, except can fully manage dependencies and tasks.

| Resource                 | Create | Read | Update | Delete |
|--------------------------|--------|------|--------|--------|
| addons                   | ❌ | ✅ | ❌ | ❌ |
| adoptionplans            | ✅ | ❌ | ❌ | ❌ |
| applications             | ❌ | ✅ | ❌ | ❌ |
| applications.facts       | ❌ | ✅ | ❌ | ❌ |
| applications.tags        | ❌ | ✅ | ❌ | ❌ |
| applications.bucket      | ❌ | ✅ | ❌ | ❌ |
| applications.analyses    | ❌ | ✅ | ❌ | ❌ |
| applications.manifests   | ❌ | ✅ | ❌ | ❌ |
| applications.assessments | ❌ | ✅ | ❌ | ❌ |
| assessments              | ❌ | ✅ | ❌ | ❌ |
| businessservices         | ❌ | ✅ | ❌ | ❌ |
| dependencies             | ✅ | ✅ | ✅ | ✅ |
| generators               | ❌ | ✅ | ❌ | ❌ |
| identities               | ❌ | ✅ | ❌ | ❌ |
| imports                  | ❌ | ✅ | ❌ | ❌ |
| jobfunctions             | ❌ | ✅ | ❌ | ❌ |
| kai                      | ✅ | ✅ | ❌ | ❌ |
| manifests                | ❌ | ✅ | ❌ | ❌ |
| migrationwaves           | ❌ | ✅ | ❌ | ❌ |
| platforms                | ❌ | ✅ | ❌ | ❌ |
| proxies                  | ❌ | ✅ | ❌ | ❌ |
| reviews                  | ❌ | ✅ | ❌ | ❌ |
| schemas                  | ❌ | ✅ | ❌ | ❌ |
| settings                 | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups        | ❌ | ✅ | ❌ | ❌ |
| stakeholders             | ❌ | ✅ | ❌ | ❌ |
| tags                     | ❌ | ✅ | ❌ | ❌ |
| tagcategories            | ❌ | ✅ | ❌ | ❌ |
| tasks                    | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket             | ✅ | ✅ | ✅ | ✅ |
| trackers                 | ❌ | ✅ | ❌ | ❌ |
| tickets                  | ❌ | ✅ | ❌ | ❌ |
| cache                    | ❌ | ✅ | ❌ | ❌ |
| files                    | ❌ | ✅ | ❌ | ❌ |
| buckets                  | ❌ | ✅ | ❌ | ❌ |
| rulesets                 | ❌ | ✅ | ❌ | ❌ |
| targets                  | ❌ | ✅ | ❌ | ❌ |
| analyses                 | ❌ | ✅ | ❌ | ❌ |
| archetypes               | ❌ | ✅ | ❌ | ❌ |
| archetypes.assessments   | ❌ | ✅ | ❌ | ❌ |
| questionnaires           | ❌ | ✅ | ❌ | ❌ |

---

## 📋 Role: **project-manager**
Read-only for most resources, except can update `applications.stakeholders` and fully manage `migrationwaves`.

| Resource                  | Create | Read | Update | Delete |
|---------------------------|--------|------|--------|--------|
| addons                    | ❌ | ✅ | ❌ | ❌ |
| adoptionplans             | ✅ | ❌ | ❌ | ❌ |
| applications              | ❌ | ✅ | ❌ | ❌ |
| applications.facts        | ❌ | ✅ | ❌ | ❌ |
| applications.tags         | ❌ | ✅ | ❌ | ❌ |
| applications.bucket       | ❌ | ✅ | ❌ | ❌ |
| applications.analyses     | ❌ | ✅ | ❌ | ❌ |
| applications.manifests    | ❌ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.assessments  | ❌ | ✅ | ❌ | ❌ |
| assessments               | ❌ | ✅ | ❌ | ❌ |
| businessservices          | ❌ | ✅ | ❌ | ❌ |
| dependencies              | ❌ | ✅ | ❌ | ❌ |
| identities                | ❌ | ✅ | ❌ | ❌ |
| generators                | ❌ | ✅ | ❌ | ❌ |
| imports                   | ❌ | ✅ | ❌ | ❌ |
| jobfunctions              | ❌ | ✅ | ❌ | ❌ |
| kai                       | ✅ | ✅ | ❌ | ❌ |
| platforms                 | ❌ | ✅ | ❌ | ❌ |
| proxies                   | ❌ | ✅ | ❌ | ❌ |
| reviews                   | ❌ | ✅ | ❌ | ❌ |
| schemas                   | ❌ | ✅ | ❌ | ❌ |
| settings                  | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups         | ❌ | ✅ | ❌ | ❌ |
| stakeholders              | ❌ | ✅ | ❌ | ❌ |
| tags                      | ❌ | ✅ | ❌ | ❌ |
| tagcategories             | ❌ | ✅ | ❌ | ❌ |
| tasks                     | ❌ | ✅ | ❌ | ❌ |
| tasks.bucket              | ❌ | ✅ | ❌ | ❌ |
| trackers                  | ❌ | ✅ | ❌ | ❌ |
| tickets                   | ❌ | ✅ | ❌ | ❌ |
| cache                     | ❌ | ✅ | ❌ | ❌ |
| files                     | ❌ | ✅ | ❌ | ❌ |
| buckets                   | ❌ | ✅ | ❌ | ❌ |
| rulesets                  | ❌ | ✅ | ❌ | ❌ |
| migrationwaves            | ✅ | ✅ | ✅ | ✅ |
| targets                   | ❌ | ✅ | ❌ | ❌ |
| analyses                  | ❌ | ✅ | ❌ | ❌ |
| archetypes                | ❌ | ✅ | ❌ | ❌ |
| archetypes.assessments    | ❌ | ✅ | ❌ | ❌ |
| questionnaires            | ❌ | ✅ | ❌ | ❌ |


## Supported Scopes

### Addon resources
- addons:delete
- addons:get
- addons:post
- addons:put

### Adoptionplan resources
- adoptionplans:post

### Analysis resources
- analyses:delete
- analyses:get
- analyses:post
- analyses:put

### Application resources
- applications:delete
- applications:get
- applications:post
- applications:put
- applications.analyses:delete
- applications.analyses:get
- applications.analyses:post
- applications.analyses:put
- applications.assessments:get
- applications.assessments:post
- applications.bucket:delete
- applications.bucket:get
- applications.bucket:post
- applications.bucket:put
- applications.facts:delete
- applications.facts:get
- applications.facts:post
- applications.facts:put
- applications.manifests:get
- applications.manifests:post
- applications.stakeholders:put
- applications.tags:delete
- applications.tags:get
- applications.tags:post
- applications.tags:put

### Archetype resources
- archetypes:delete
- archetypes:get
- archetypes:post
- archetypes:put
- archetypes.assessments:get
- archetypes.assessments:post

### Assessment resources
- assessments:delete
- assessments:get
- assessments:post
- assessments:put

### Bucket resources
- buckets:delete
- buckets:get
- buckets:post
- buckets:put

### Businessservice resources
- businessservices:delete
- businessservices:get
- businessservices:post
- businessservices:put

### Cache resources
- cache:delete
- cache:get

### Dependency resources
- dependencies:delete
- dependencies:get
- dependencies:post
- dependencies:put

### File resources
- files:delete
- files:get
- files:post
- files:put

### Generator resources
- generators:delete
- generators:get
- generators:post
- generators:put

### Identity resources
- identities:delete
- identities:get
- identities:post
- identities:put

### Import resources
- imports:delete
- imports:get
- imports:post
- imports:put

### Jobfunction resources
- jobfunctions:delete
- jobfunctions:get
- jobfunctions:post
- jobfunctions:put

### Kai resources
- kai:get
- kai:post

### Manifest resources
- manifests:delete
- manifests:get
- manifests:post
- manifests:put

### Migrationwave resources
- migrationwaves:delete
- migrationwaves:get
- migrationwaves:post
- migrationwaves:put

### Platform resources
- platforms:delete
- platforms:get
- platforms:post
- platforms:put

### Proxy resources
- proxies:delete
- proxies:get
- proxies:post
- proxies:put

### Questionnaire resources
- questionnaires:delete
- questionnaires:get
- questionnaires:post
- questionnaires:put

### Review resources
- reviews:delete
- reviews:get
- reviews:post
- reviews:put

### Ruleset resources
- rulesets:delete
- rulesets:get
- rulesets:post
- rulesets:put

### Schema resources
- schemas:delete
- schemas:get
- schemas:post
- schemas:put

### Setting resources
- settings:delete
- settings:get
- settings:post
- settings:put

### Stakeholdergroup resources
- stakeholdergroups:delete
- stakeholdergroups:get
- stakeholdergroups:post
- stakeholdergroups:put

### Stakeholder resources
- stakeholders:delete
- stakeholders:get
- stakeholders:post
- stakeholders:put

### Tagcategory resources
- tagcategories:delete
- tagcategories:get
- tagcategories:post
- tagcategories:put

### Tag resources
- tags:delete
- tags:get
- tags:post
- tags:put

### Target resources
- targets:delete
- targets:get
- targets:post
- targets:put

### Task resources
- tasks:delete
- tasks:get
- tasks:patch
- tasks:post
- tasks:put
- tasks.bucket:delete
- tasks.bucket:get
- tasks.bucket:post
- tasks.bucket:put

### Ticket resources
- tickets:delete
- tickets:get
- tickets:post
- tickets:put

### Tracker resources
- trackers:delete
- trackers:get
- trackers:post
- trackers:put



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

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ✅ | ✅ | ✅ | ✅ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| applications | ✅ | ✅ | ✅ | ✅ |
| applications.facts | ✅ | ✅ | ✅ | ✅ |
| applications.tags | ✅ | ✅ | ✅ | ✅ |
| applications.bucket | ✅ | ✅ | ✅ | ✅ |
| applications.analyses | ✅ | ✅ | ✅ | ✅ |
| applications.manifest | ✅ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.assessments | ✅ | ✅ | ❌ | ❌ |
| assessments | ✅ | ✅ | ✅ | ✅ |
| businessservices | ✅ | ✅ | ✅ | ✅ |
| dependencies | ✅ | ✅ | ✅ | ✅ |
| generators | ✅ | ✅ | ✅ | ✅ |
| identities | ✅ | ✅ | ✅ | ✅ |
| imports | ✅ | ✅ | ✅ | ✅ |
| jobfunctions | ✅ | ✅ | ✅ | ✅ |
| kai | ✅ | ✅ | ❌ | ❌ |
| manifests | ✅ | ✅ | ✅ | ✅ |
| migrationwaves | ✅ | ✅ | ✅ | ✅ |
| platforms | ✅ | ✅ | ✅ | ✅ |
| proxies | ✅ | ✅ | ✅ | ✅ |
| reviews | ✅ | ✅ | ✅ | ✅ |
| settings | ✅ | ✅ | ✅ | ✅ |
| stakeholdergroups | ✅ | ✅ | ✅ | ✅ |
| stakeholders | ✅ | ✅ | ✅ | ✅ |
| tags | ✅ | ✅ | ✅ | ✅ |
| tagcategories | ✅ | ✅ | ✅ | ✅ |
| tasks | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket | ✅ | ✅ | ✅ | ✅ |
| trackers | ✅ | ✅ | ✅ | ✅ |
| tickets | ✅ | ✅ | ✅ | ✅ |
| cache | ❌ | ✅ | ❌ | ✅ |
| files | ✅ | ✅ | ✅ | ✅ |
| buckets | ✅ | ✅ | ✅ | ✅ |
| rulesets | ✅ | ✅ | ✅ | ✅ |
| targets | ✅ | ✅ | ✅ | ✅ |
| analyses | ✅ | ✅ | ✅ | ✅ |
| archetypes | ✅ | ✅ | ✅ | ✅ |
| archetypes.assessments | ✅ | ✅ | ❌ | ❌ |
| questionnaires | ✅ | ✅ | ✅ | ✅ |

---

## 🛠 Role: **tackle-architect**
Broad create/update/delete rights but restricted on identities, proxies, settings, and trackers.

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ✅ | ✅ | ✅ | ✅ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| applications | ✅ | ✅ | ✅ | ✅ |
| applications.facts | ✅ | ✅ | ✅ | ✅ |
| applications.tags | ✅ | ✅ | ✅ | ✅ |
| applications.bucket | ✅ | ✅ | ✅ | ✅ |
| applications.analyses | ✅ | ✅ | ✅ | ✅ |
| applications.manifest | ✅ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.assessments | ✅ | ✅ | ❌ | ❌ |
| assessments | ✅ | ✅ | ✅ | ✅ |
| businessservices | ✅ | ✅ | ✅ | ✅ |
| dependencies | ✅ | ✅ | ✅ | ✅ |
| generators | ✅ | ✅ | ✅ | ✅ |
| identities | ❌ | ✅ | ❌ | ❌ |
| imports | ✅ | ✅ | ✅ | ✅ |
| jobfunctions | ✅ | ✅ | ✅ | ✅ |
| kai | ✅ | ✅ | ❌ | ❌ |
| manifests | ✅ | ✅ | ✅ | ✅ |
| migrationwaves | ✅ | ✅ | ✅ | ✅ |
| platforms | ✅ | ✅ | ✅ | ✅ |
| proxies | ❌ | ✅ | ❌ | ❌ |
| reviews | ✅ | ✅ | ✅ | ✅ |
| settings | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups | ✅ | ✅ | ✅ | ✅ |
| stakeholders | ✅ | ✅ | ✅ | ✅ |
| tags | ✅ | ✅ | ✅ | ✅ |
| tagcategories | ✅ | ✅ | ✅ | ✅ |
| tasks | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket | ✅ | ✅ | ✅ | ✅ |
| trackers | ❌ | ✅ | ❌ | ❌ |
| tickets | ✅ | ✅ | ✅ | ✅ |
| cache | ❌ | ✅ | ❌ | ❌ |
| files | ✅ | ✅ | ✅ | ✅ |
| buckets | ✅ | ✅ | ✅ | ✅ |
| rulesets | ✅ | ✅ | ✅ | ✅ |
| targets | ✅ | ✅ | ✅ | ✅ |
| analyses | ✅ | ✅ | ✅ | ✅ |
| archetypes | ✅ | ✅ | ✅ | ✅ |
| archetypes.assessments | ✅ | ✅ | ❌ | ❌ |
| questionnaires | ❌ | ✅ | ❌ | ❌ |

---

## 🚚 Role: **tackle-migrator**
Mostly read-only, except can fully manage dependencies and tasks.

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ❌ | ✅ | ❌ | ❌ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| applications | ❌ | ✅ | ❌ | ❌ |
| applications.facts | ❌ | ✅ | ❌ | ❌ |
| applications.tags | ❌ | ✅ | ❌ | ❌ |
| applications.bucket | ❌ | ✅ | ❌ | ❌ |
| applications.analyses | ❌ | ✅ | ❌ | ❌ |
| applications.manifest | ❌ | ✅ | ❌ | ❌ |
| applications.assessments | ❌ | ✅ | ❌ | ❌ |
| assessments | ❌ | ✅ | ❌ | ❌ |
| businessservices | ❌ | ✅ | ❌ | ❌ |
| dependencies | ✅ | ✅ | ✅ | ✅ |
| generators | ❌ | ✅ | ❌ | ❌ |
| identities | ❌ | ✅ | ❌ | ❌ |
| imports | ❌ | ✅ | ❌ | ❌ |
| jobfunctions | ❌ | ✅ | ❌ | ❌ |
| kai | ✅ | ✅ | ❌ | ❌ |
| manifests | ❌ | ✅ | ❌ | ❌ |
| migrationwaves | ❌ | ✅ | ❌ | ❌ |
| platforms | ❌ | ✅ | ❌ | ❌ |
| proxies | ❌ | ✅ | ❌ | ❌ |
| reviews | ❌ | ✅ | ❌ | ❌ |
| settings | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups | ❌ | ✅ | ❌ | ❌ |
| stakeholders | ❌ | ✅ | ❌ | ❌ |
| tags | ❌ | ✅ | ❌ | ❌ |
| tagcategories | ❌ | ✅ | ❌ | ❌ |
| tasks | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket | ✅ | ✅ | ✅ | ✅ |
| trackers | ❌ | ✅ | ❌ | ❌ |
| tickets | ❌ | ✅ | ❌ | ❌ |
| cache | ❌ | ✅ | ❌ | ❌ |
| files | ❌ | ✅ | ❌ | ❌ |
| buckets | ❌ | ✅ | ❌ | ❌ |
| rulesets | ❌ | ✅ | ❌ | ❌ |
| targets | ❌ | ✅ | ❌ | ❌ |
| analyses | ❌ | ✅ | ❌ | ❌ |
| archetypes | ❌ | ✅ | ❌ | ❌ |
| archetypes.assessments | ❌ | ✅ | ❌ | ❌ |
| questionnaires | ❌ | ✅ | ❌ | ❌ |

---

## 📋 Role: **tackle-project-manager**
Read-only for most resources, except can update `applications.stakeholders` and fully manage `migrationwaves`.

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ❌ | ✅ | ❌ | ❌ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| applications | ❌ | ✅ | ❌ | ❌ |
| applications.facts | ❌ | ✅ | ❌ | ❌ |
| applications.tags | ❌ | ✅ | ❌ | ❌ |
| applications.bucket | ❌ | ✅ | ❌ | ❌ |
| applications.analyses | ❌ | ✅ | ❌ | ❌ |
| applications.manifest | ❌ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.assessments | ❌ | ✅ | ❌ | ❌ |
| assessments | ❌ | ✅ | ❌ | ❌ |
| businessservices | ❌ | ✅ | ❌ | ❌ |
| dependencies | ❌ | ✅ | ❌ | ❌ |
| identities | ❌ | ✅ | ❌ | ❌ |
| generators | ❌ | ✅ | ❌ | ❌ |
| imports | ❌ | ✅ | ❌ | ❌ |
| jobfunctions | ❌ | ✅ | ❌ | ❌ |
| kai | ✅ | ✅ | ❌ | ❌ |
| platforms | ❌ | ✅ | ❌ | ❌ |
| proxies | ❌ | ✅ | ❌ | ❌ |
| reviews | ❌ | ✅ | ❌ | ❌ |
| settings | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups | ❌ | ✅ | ❌ | ❌ |
| stakeholders | ❌ | ✅ | ❌ | ❌ |
| tags | ❌ | ✅ | ❌ | ❌ |
| tagcategories | ❌ | ✅ | ❌ | ❌ |
| tasks | ❌ | ✅ | ❌ | ❌ |
| tasks.bucket | ❌ | ✅ | ❌ | ❌ |
| trackers | ❌ | ✅ | ❌ | ❌ |
| tickets | ❌ | ✅ | ❌ | ❌ |
| cache | ❌ | ✅ | ❌ | ❌ |
| files | ❌ | ✅ | ❌ | ❌ |
| buckets | ❌ | ✅ | ❌ | ❌ |
| rulesets | ❌ | ✅ | ❌ | ❌ |
| migrationwaves | ✅ | ✅ | ✅ | ✅ |
| targets | ❌ | ✅ | ❌ | ❌ |
| analyses | ❌ | ✅ | ❌ | ❌ |
| archetypes | ❌ | ✅ | ❌ | ❌ |
| archetypes.assessments | ❌ | ✅ | ❌ | ❌ |
| questionnaires | ❌ | ✅ | ❌ | ❌ |

## Scopes

- addons:delete
- addons:get
- addons:post
- addons:put
- adoptionplans:post
- analyses:delete
- analyses:get
- analyses:post
- analyses:put
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
- applications.manifest:get
- applications.manifest:post
- applications.stakeholders:put
- applications.tags:delete
- applications.tags:get
- applications.tags:post
- applications.tags:put
- archetypes:delete
- archetypes:get
- archetypes:post
- archetypes:put
- archetypes.assessments:get
- archetypes.assessments:post
- assessments:delete
- assessments:get
- assessments:post
- assessments:put
- buckets:delete
- buckets:get
- buckets:post
- buckets:put
- businessservices:delete
- businessservices:get
- businessservices:post
- businessservices:put
- cache:delete
- cache:get
- dependencies:delete
- dependencies:get
- dependencies:post
- dependencies:put
- files:delete
- files:get
- files:post
- files:put
- generators:delete
- generators:get
- generators:post
- generators:put
- identities:delete
- identities:get
- identities:post
- identities:put
- imports:delete
- imports:get
- imports:post
- imports:put
- jobfunctions:delete
- jobfunctions:get
- jobfunctions:post
- jobfunctions:put
- kai:get
- kai:post
- manifests:delete
- manifests:get
- manifests:post
- manifests:put
- migrationwaves:delete
- migrationwaves:get
- migrationwaves:post
- migrationwaves:put
- platforms:delete
- platforms:get
- platforms:post
- platforms:put
- proxies:delete
- proxies:get
- proxies:post
- proxies:put
- questionnaires:delete
- questionnaires:get
- questionnaires:post
- questionnaires:put
- reviews:delete
- reviews:get
- reviews:post
- reviews:put
- rulesets:delete
- rulesets:get
- rulesets:post
- rulesets:put
- settings:delete
- settings:get
- settings:post
- settings:put
- stakeholdergroups:delete
- stakeholdergroups:get
- stakeholdergroups:post
- stakeholdergroups:put
- stakeholders:delete
- stakeholders:get
- stakeholders:post
- stakeholders:put
- tagcategories:delete
- tagcategories:get
- tagcategories:post
- tagcategories:put
- tags:delete
- tags:get
- tags:post
- tags:put
- targets:delete
- targets:get
- targets:post
- targets:put
- tasks:delete
- tasks:get
- tasks:patch
- tasks:post
- tasks:put
- tasks.bucket:delete
- tasks.bucket:get
- tasks.bucket:post
- tasks.bucket:put
- tickets:delete
- tickets:get
- tickets:post
- tickets:put
- trackers:delete
- trackers:get
- trackers:post
- trackers:put


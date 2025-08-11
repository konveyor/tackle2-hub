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

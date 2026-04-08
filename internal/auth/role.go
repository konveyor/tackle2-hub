package auth

import (
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var Settings = &settings.Settings

// AddonRole defines the addon scopes.
var AddonRole = []string{
	"addons:get",
	"analysis.profiles:get",
	"applications:get",
	"applications:post",
	"applications:put",
	"applications.tags:*",
	"applications.facts:*",
	"applications.bucket:*",
	"applications.analyses:*",
	"applications.manifests:*",
	"archetypes:get",
	"generators:get",
	"identities:get",
	"identities:decrypt",
	"manifests:*",
	"platforms:get",
	"proxies:get",
	"schemas:get",
	"settings:get",
	"tags:*",
	"tagcategories:*",
	"tasks:get",
	"tasks.report:*",
	"tasks.bucket:get",
	"files:*",
	"rulesets:get",
	"targets:get",
}

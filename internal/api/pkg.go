package api

// @title Konveyor Hub API
// @version 0.3.z
// @description
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @accept application/json
// @produce application/json

import (
	"github.com/gin-gonic/gin"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/api/filter"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var (
	Settings = &settings.Settings
	log      = logr.New("api", Settings.Log.Web)
)

// Params
const (
	ID        = api.ID
	ID2       = api.ID2
	Key       = api.Key
	Name      = api.Name
	Filter    = filter.QueryParam
	Wildcard  = api.Wildcard
	FileField = api.FileField
	Decrypted = api.Decrypted
)

// Scopes
const (
	MethodDecrypt = "decrypt"
)

// Headers
const (
	Accept        = api.Accept
	Authorization = api.Authorization
	ContentType   = api.ContentType
	Directory     = api.Directory
	Total         = api.Total
)

// MIME Types.
const (
	MIMEOCTETSTREAM = api.MIMEOCTETSTREAM
	TAR             = api.TAR
)

// BindMIMEs supported binding MIME types.
var BindMIMEs = []string{api.MIMEJSON, api.MIMEYAML}

// Header Values
const (
	DirectoryExpand = api.DirectoryExpand
)

// Map REST resource.
type Map = resource.Map

// Ref REST resource reference.
type Ref = resource.Ref

// All builds all handlers.
func All() []Handler {
	return []Handler{
		&AddonHandler{},
		&AdoptionPlanHandler{},
		&AnalysisProfileHandler{},
		&AnalysisHandler{},
		&ApplicationHandler{},
		&ConfigMapHandler{},
		&AuthHandler{},
		&BusinessServiceHandler{},
		&CacheHandler{},
		&DependencyHandler{},
		&GeneratorHandler{},
		&ImportHandler{},
		&JobFunctionHandler{},
		&IdentityHandler{},
		&PlatformHandler{},
		&ProxyHandler{},
		&ManifestHandler{},
		&ReviewHandler{},
		&RuleSetHandler{},
		&SchemaHandler{},
		&SettingHandler{},
		&ServiceHandler{},
		&StakeholderHandler{},
		&StakeholderGroupHandler{},
		&TagHandler{},
		&TagCategoryHandler{},
		&TaskHandler{},
		&TaskGroupHandler{},
		&TicketHandler{},
		&TrackerHandler{},
		&BucketHandler{},
		&FileHandler{},
		&MigrationWaveHandler{},
		&BatchHandler{},
		&TargetHandler{},
		&QuestionnaireHandler{},
		&AssessmentHandler{},
		&ArchetypeHandler{},
	}
}

// Handler API.
type Handler interface {
	AddRoutes(e *gin.Engine)
}

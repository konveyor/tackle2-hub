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
	"github.com/gin-gonic/gin/binding"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/settings"
)

var (
	Settings = &settings.Settings
	log      = logr.WithName("api")
)

// Params
const (
	ID        = "id"
	ID2       = "id2"
	Key       = "key"
	Name      = "name"
	Filter    = filter.QueryParam
	Wildcard  = "wildcard"
	FileField = "file"
)

// Headers
const (
	Accept        = "Accept"
	Authorization = "Authorization"
	ContentLength = "Content-Length"
	ContentType   = "Content-Type"
	Directory     = "X-Directory"
	Total         = "X-Total"
)

// MIME Types.
const (
	MIMEOCTETSTREAM = "application/octet-stream"
	TAR             = "application/x-tar"
)

// BindMIMEs supported binding MIME types.
var BindMIMEs = []string{binding.MIMEJSON, binding.MIMEYAML}

// Header Values
const (
	DirectoryArchive = "archive"
	DirectoryExpand  = "expand"
)

// All builds all handlers.
func All() []Handler {
	return []Handler{
		&AddonHandler{},
		&AdoptionPlanHandler{},
		&AnalysisHandler{},
		&ApplicationHandler{},
		&AuthHandler{},
		&BusinessServiceHandler{},
		&CacheHandler{},
		&DependencyHandler{},
		&ImportHandler{},
		&JobFunctionHandler{},
		&IdentityHandler{},
		&ProxyHandler{},
		&ReviewHandler{},
		&RuleSetHandler{},
		&SchemaHandler{},
		&SettingHandler{},
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

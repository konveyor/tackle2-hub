package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/settings"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Settings = &settings.Settings
	log      = logging.WithName("api")
)

//
// Params
const (
	ID        = "id"
	ID2       = "id2"
	Key       = "key"
	Name      = "name"
	Wildcard  = "wildcard"
	FileField = "file"
)

//
// Headers
const (
	Accept        = "Accept"
	Authorization = "Authorization"
	ContentLength = "Content-Length"
	ContentType   = "Content-Type"
	Directory     = "X-Directory"
)

//
// MIME Types.
const (
	MIMEOCTETSTREAM = "application/octet-stream"
)

//
// BindMIMEs supported binding MIME types.
var BindMIMEs = []string{binding.MIMEJSON, binding.MIMEYAML}

//
// Header Values
const (
	DirectoryArchive = "archive"
	DirectoryExpand  = "expand"
)

//
// All builds all handlers.
func All() []Handler {
	return []Handler{
		&AddonHandler{},
		&AdoptionPlanHandler{},
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
		&RuleBundleHandler{},
		&SchemaHandler{},
		&SettingHandler{},
		&StakeholderHandler{},
		&StakeholderGroupHandler{},
		&TagHandler{},
		&TagCategoryHandler{},
		&TaskHandler{},
		&TaskGroupHandler{},
		&PathfinderHandler{},
		&TicketHandler{},
		&TrackerHandler{},
		&BucketHandler{},
		&FileHandler{},
		&MigrationWaveHandler{},
	}
}

//
// Handler API.
type Handler interface {
	With(client.Client)
	AddRoutes(e *gin.Engine)
}

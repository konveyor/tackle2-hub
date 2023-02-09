package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
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
// Accepted (mime)
const (
	AppJson  = "application/json"
	AppOctet = "application/octet-stream"
)

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
		&SchemaHandler{},
		&SettingHandler{},
		&StakeholderHandler{},
		&StakeholderGroupHandler{},
		&TagHandler{},
		&TagTypeHandler{},
		&TaskHandler{},
		&TaskGroupHandler{},
		&PathfinderHandler{},
		&TicketHandler{},
		&TrackerHandler{},
		&FileHandler{},
	}
}

//
// Handler.
type Handler interface {
	With(*gorm.DB, client.Client)
	AddRoutes(e *gin.Engine)
}

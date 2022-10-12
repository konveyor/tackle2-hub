package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/auth"
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
	ID       = "id"
	Key      = "key"
	Name     = "name"
	Wildcard = "wildcard"
)

//
// Headers
const (
	Authorization = "Authorization"
	ContentLength = "Content-Length"
	ContentType   = "Content-Type"
)

//
// All builds all handlers.
func All() []Handler {
	return []Handler{
		&AddonHandler{},
		&AdoptionPlanHandler{},
		&ApplicationHandler{},
		&BusinessServiceHandler{},
		&DependencyHandler{},
		&ImportHandler{},
		&JobFunctionHandler{},
		&IdentityHandler{},
		&ProxyHandler{},
		&ReviewHandler{},
		&SettingHandler{},
		&StakeholderHandler{},
		&StakeholderGroupHandler{},
		&TagHandler{},
		&TagTypeHandler{},
		&TaskHandler{},
		&TaskGroupHandler{},
		&VolumeHandler{},
		&PathfinderHandler{},
	}
}

//
// Handler.
type Handler interface {
	With(*gorm.DB, client.Client, auth.Provider)
	AddRoutes(e *gin.Engine)
}

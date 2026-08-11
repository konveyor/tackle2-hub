package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	agent "github.com/konveyor/agentic-controller/api/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/shared/api"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// ACPPort is the ACP endpoint port on the sandbox service.
const ACPPort = 4000

// ManagedLabel is the label used to identify resources managed by the hub.
const ManagedLabel = "konveyor.io/managed"

// ACPSecretKeys are the Secret data keys tried when reading the ACP secret.
var ACPSecretKeys = []string{"secret-key", "ACP_SECRET_KEY"}

// AgenticHandler handles AI agent routes.
type AgenticHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AgenticHandler) AddRoutes(e *gin.Engine) {
	// Agent
	routeGroup := e.Group("/")
	routeGroup.Use(Required("agentic.agents"))
	routeGroup.GET(api.AgenticAgentsRoute, h.AgentList)
	routeGroup.GET(api.AgenticAgentsRoute+"/", h.AgentList)
	routeGroup.POST(api.AgenticAgentsRoute, h.AgentCreate)
	routeGroup.GET(api.AgenticAgentRoute, h.AgentGet)
	routeGroup.PUT(api.AgenticAgentRoute, h.AgentUpdate)
	routeGroup.DELETE(api.AgenticAgentRoute, h.AgentDelete)
	// SkillCard
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.skills"))
	routeGroup.GET(api.AgenticSkillsRoute, h.SkillList)
	routeGroup.GET(api.AgenticSkillsRoute+"/", h.SkillList)
	routeGroup.POST(api.AgenticSkillsRoute, h.SkillCreate)
	routeGroup.GET(api.AgenticSkillRoute, h.SkillGet)
	routeGroup.PUT(api.AgenticSkillRoute, h.SkillUpdate)
	routeGroup.DELETE(api.AgenticSkillRoute, h.SkillDelete)
	// SkillCollection
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.skillcollections"))
	routeGroup.GET(api.AgenticSkillCollectionsRoute, h.SkillCollectionList)
	routeGroup.GET(api.AgenticSkillCollectionsRoute+"/", h.SkillCollectionList)
	routeGroup.POST(api.AgenticSkillCollectionsRoute, h.SkillCollectionCreate)
	routeGroup.GET(api.AgenticSkillCollectionRoute, h.SkillCollectionGet)
	routeGroup.PUT(api.AgenticSkillCollectionRoute, h.SkillCollectionUpdate)
	routeGroup.DELETE(api.AgenticSkillCollectionRoute, h.SkillCollectionDelete)
	// Gateway
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.gateways"))
	routeGroup.GET(api.AgenticGatewaysRoute, h.GatewayList)
	routeGroup.GET(api.AgenticGatewaysRoute+"/", h.GatewayList)
	routeGroup.POST(api.AgenticGatewaysRoute, h.GatewayCreate)
	routeGroup.GET(api.AgenticGatewayRoute, h.GatewayGet)
	routeGroup.PUT(api.AgenticGatewayRoute, h.GatewayUpdate)
	routeGroup.DELETE(api.AgenticGatewayRoute, h.GatewayDelete)
	// AgentRun
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.runs"))
	routeGroup.GET(api.AgenticAgentRunsRoute, h.RunList)
	routeGroup.GET(api.AgenticAgentRunsRoute+"/", h.RunList)
	routeGroup.POST(api.AgenticAgentRunsRoute, h.RunCreate)
	routeGroup.GET(api.AgenticAgentRunRoute, h.RunGet)
	// AgentRun ACP
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.runs.acp"))
	routeGroup.GET(api.AgenticAgentRunACPRoute, h.RunACP)
	// AgentWorkflow
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.workflows"))
	routeGroup.GET(api.AgenticWorkflowsRoute, h.WorkflowList)
	routeGroup.GET(api.AgenticWorkflowsRoute+"/", h.WorkflowList)
	routeGroup.POST(api.AgenticWorkflowsRoute, h.WorkflowCreate)
	routeGroup.GET(api.AgenticWorkflowRoute, h.WorkflowGet)
	routeGroup.PUT(api.AgenticWorkflowRoute, h.WorkflowUpdate)
	routeGroup.DELETE(api.AgenticWorkflowRoute, h.WorkflowDelete)
	// AgentWorkflowRun
	routeGroup = e.Group("/")
	routeGroup.Use(Required("agentic.workflowruns"))
	routeGroup.GET(api.AgenticWorkflowRunsRoute, h.WorkflowRunList)
	routeGroup.GET(api.AgenticWorkflowRunsRoute+"/", h.WorkflowRunList)
	routeGroup.POST(api.AgenticWorkflowRunsRoute, h.WorkflowRunCreate)
	routeGroup.GET(api.AgenticWorkflowRunRoute, h.WorkflowRunGet)
}

//
// Agent
//

// AgentGet godoc
// @summary Get an agent by name.
// @description Get an agent by name.
// @tags agents
// @produce json
// @success 200 {object} Agent
// @router /agentic/agents/{name} [get]
// @param name path string true "Agent name"
func (h AgenticHandler) AgentGet(ctx *gin.Context) {
	r := &Agent{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// AgentList godoc
// @summary List all agents.
// @description List all agents.
// @tags agents
// @produce json
// @success 200 {object} []Agent
// @router /agentic/agents [get]
func (h AgenticHandler) AgentList(ctx *gin.Context) {
	list := &AgentList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Hub.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// AgentCreate godoc
// @summary Create an agent.
// @description Create an agent.
// @tags agents
// @accept json
// @produce json
// @success 201 {object} Agent
// @router /agentic/agents [post]
// @param agent body Agent true "Agent data"
func (h AgenticHandler) AgentCreate(ctx *gin.Context) {
	r := &Agent{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Namespace = Settings.Hub.Namespace
	err = h.Client(ctx).Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// AgentUpdate godoc
// @summary Update an agent.
// @description Update an agent.
// @tags agents
// @accept json
// @success 204
// @router /agentic/agents/{name} [put]
// @param name path string true "Agent name"
// @param agent body Agent true "Agent data"
func (h AgenticHandler) AgentUpdate(ctx *gin.Context) {
	current := &Agent{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &Agent{}
	err = h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	current.Spec = r.Spec
	err = h.Client(ctx).Update(context.TODO(), current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// AgentDelete godoc
// @summary Delete an agent.
// @description Delete an agent.
// @tags agents
// @success 204
// @router /agentic/agents/{name} [delete]
// @param name path string true "Agent name"
func (h AgenticHandler) AgentDelete(ctx *gin.Context) {
	r := &Agent{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Client(ctx).Delete(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

//
// SkillCard
//

// SkillGet godoc
// @summary Get a skill card by name.
// @description Get a skill card by name.
// @tags skills
// @produce json
// @success 200 {object} SkillCard
// @router /agentic/skills/{name} [get]
// @param name path string true "SkillCard name"
func (h AgenticHandler) SkillGet(ctx *gin.Context) {
	r := &SkillCard{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// SkillList godoc
// @summary List all skill cards.
// @description List all skill cards.
// @tags skills
// @produce json
// @success 200 {object} []SkillCard
// @router /agentic/skills [get]
func (h AgenticHandler) SkillList(ctx *gin.Context) {
	list := &SkillCardList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Hub.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// SkillCreate godoc
// @summary Create a skill card.
// @description Create a skill card.
// @tags skills
// @accept json
// @produce json
// @success 201 {object} SkillCard
// @router /agentic/skills [post]
// @param skill body SkillCard true "SkillCard data"
func (h AgenticHandler) SkillCreate(ctx *gin.Context) {
	r := &SkillCard{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Namespace = Settings.Hub.Namespace
	err = h.Client(ctx).Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// SkillUpdate godoc
// @summary Update a skill card.
// @description Update a skill card.
// @tags skills
// @accept json
// @success 204
// @router /agentic/skills/{name} [put]
// @param name path string true "SkillCard name"
// @param skill body SkillCard true "SkillCard data"
func (h AgenticHandler) SkillUpdate(ctx *gin.Context) {
	current := &SkillCard{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &SkillCard{}
	err = h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	current.Spec = r.Spec
	err = h.Client(ctx).Update(context.TODO(), current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// SkillDelete godoc
// @summary Delete a skill card.
// @description Delete a skill card.
// @tags skills
// @success 204
// @router /agentic/skills/{name} [delete]
// @param name path string true "SkillCard name"
func (h AgenticHandler) SkillDelete(ctx *gin.Context) {
	r := &SkillCard{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Client(ctx).Delete(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

//
// SkillCollection
//

// SkillCollectionGet godoc
// @summary Get a skill collection by name.
// @description Get a skill collection by name.
// @tags skillcollections
// @produce json
// @success 200 {object} SkillCollection
// @router /agentic/skillcollections/{name} [get]
// @param name path string true "SkillCollection name"
func (h AgenticHandler) SkillCollectionGet(ctx *gin.Context) {
	r := &SkillCollection{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// SkillCollectionList godoc
// @summary List all skill collections.
// @description List all skill collections.
// @tags skillcollections
// @produce json
// @success 200 {object} []SkillCollection
// @router /agentic/skillcollections [get]
func (h AgenticHandler) SkillCollectionList(ctx *gin.Context) {
	list := &SkillCollectionList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Hub.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// SkillCollectionCreate godoc
// @summary Create a skill collection.
// @description Create a skill collection.
// @tags skillcollections
// @accept json
// @produce json
// @success 201 {object} SkillCollection
// @router /agentic/skillcollections [post]
// @param collection body SkillCollection true "SkillCollection data"
func (h AgenticHandler) SkillCollectionCreate(ctx *gin.Context) {
	r := &SkillCollection{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Namespace = Settings.Hub.Namespace
	err = h.Client(ctx).Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// SkillCollectionUpdate godoc
// @summary Update a skill collection.
// @description Update a skill collection.
// @tags skillcollections
// @accept json
// @success 204
// @router /agentic/skillcollections/{name} [put]
// @param name path string true "SkillCollection name"
// @param collection body SkillCollection true "SkillCollection data"
func (h AgenticHandler) SkillCollectionUpdate(ctx *gin.Context) {
	current := &SkillCollection{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &SkillCollection{}
	err = h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	current.Spec = r.Spec
	err = h.Client(ctx).Update(context.TODO(), current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// SkillCollectionDelete godoc
// @summary Delete a skill collection.
// @description Delete a skill collection.
// @tags skillcollections
// @success 204
// @router /agentic/skillcollections/{name} [delete]
// @param name path string true "SkillCollection name"
func (h AgenticHandler) SkillCollectionDelete(ctx *gin.Context) {
	r := &SkillCollection{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Client(ctx).Delete(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

//
// Gateway
//

// GatewayGet godoc
// @summary Get a gateway by name.
// @description Get a gateway by name.
// @tags gateways
// @produce json
// @success 200 {object} Gateway
// @router /agentic/gateways/{name} [get]
// @param name path string true "Gateway name"
func (h AgenticHandler) GatewayGet(ctx *gin.Context) {
	r := &Gateway{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// GatewayList godoc
// @summary List all gateways.
// @description List all gateways.
// @tags gateways
// @produce json
// @success 200 {object} []Gateway
// @router /agentic/gateways [get]
func (h AgenticHandler) GatewayList(ctx *gin.Context) {
	list := &GatewayList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Hub.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// GatewayCreate godoc
// @summary Create a gateway.
// @description Create a gateway.
// @tags gateways
// @accept json
// @produce json
// @success 201 {object} Gateway
// @router /agentic/gateways [post]
// @param gateway body Gateway true "Gateway data"
func (h AgenticHandler) GatewayCreate(ctx *gin.Context) {
	r := &Gateway{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Namespace = Settings.Hub.Namespace
	err = h.Client(ctx).Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// GatewayUpdate godoc
// @summary Update a gateway.
// @description Update a gateway.
// @tags gateways
// @accept json
// @success 204
// @router /agentic/gateways/{name} [put]
// @param name path string true "Gateway name"
// @param gateway body Gateway true "Gateway data"
func (h AgenticHandler) GatewayUpdate(ctx *gin.Context) {
	current := &Gateway{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &Gateway{}
	err = h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	current.Spec = r.Spec
	err = h.Client(ctx).Update(context.TODO(), current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// GatewayDelete godoc
// @summary Delete a gateway.
// @description Delete a gateway.
// @tags gateways
// @success 204
// @router /agentic/gateways/{name} [delete]
// @param name path string true "Gateway name"
func (h AgenticHandler) GatewayDelete(ctx *gin.Context) {
	r := &Gateway{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Client(ctx).Delete(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

//
// AgentRun
//

// RunGet godoc
// @summary Get an agent run by name.
// @description Get an agent run by name.
// @tags runs
// @produce json
// @success 200 {object} AgentRun
// @router /agentic/runs/{name} [get]
// @param name path string true "AgentRun name"
func (h AgenticHandler) RunGet(ctx *gin.Context) {
	r := &AgentRun{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// RunList godoc
// @summary List all agent runs.
// @description List all agent runs.
// @tags runs
// @produce json
// @success 200 {object} []AgentRun
// @router /agentic/runs [get]
func (h AgenticHandler) RunList(ctx *gin.Context) {
	list := &AgentRunList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		k8s.InNamespace(Settings.Hub.Namespace),
		k8s.MatchingLabels{
			ManagedLabel: "true",
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// RunCreate godoc
// @summary Create an agent run.
// @description Create an agent run.
// @tags runs
// @accept json
// @produce json
// @success 201 {object} AgentRun
// @router /agentic/runs [post]
// @param run body AgentRun true "AgentRun data"
func (h AgenticHandler) RunCreate(ctx *gin.Context) {
	r := &AgentRun{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	client := h.Client(ctx)
	secret, tokenId, err := h.tokenSecret(r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = client.Create(context.TODO(), secret)
	if err != nil {
		_ = auth.Idp().Revoke(tokenId)
		_ = ctx.Error(err)
		return
	}
	defer func() {
		if err != nil {
			_ = auth.Idp().Revoke(tokenId)
			_ = client.Delete(context.TODO(), secret)
		}
	}()
	h.injectLabels(r)
	r.Namespace = Settings.Hub.Namespace
	r.Spec.EnvFrom = append(
		r.Spec.EnvFrom,
		core.EnvFromSource{
			SecretRef: &core.SecretEnvSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: secret.Name,
				},
			},
		})
	r.Spec.Env = append(
		r.Spec.Env,
		core.EnvVar{
			Name:  "HUB_BASE_URL",
			Value: Settings.Addon.Hub.URL,
		})
	err = client.Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	secret.OwnerReferences = []v1.OwnerReference{
		{
			APIVersion: "v1alpha1",
			Kind:       "AgentRun",
			Name:       r.Name,
			UID:        r.UID,
		},
	}
	err = client.Update(context.TODO(), secret)
	if err != nil {
		_ = client.Delete(context.TODO(), r)
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// RunACP godoc
// @summary WebSocket proxy to agent pod ACP endpoint.
// @description Upgrades the connection to a WebSocket and proxies frames
// @description bidirectionally to the agent pod's ACP endpoint.
// @tags runs
// @router /agentic/runs/{name}/acp [get]
// @param name path string true "AgentRun name"
func (h AgenticHandler) RunACP(ctx *gin.Context) {
	r := &AgentRun{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if r.Status.SandboxName == "" || r.Status.SecretKeyRef == nil {
		err = &NotAvailableError{
			Name:   "ACP",
			Reason: "Status not populated.",
		}
		_ = ctx.Error(err)
		return
	}
	key, err := h.acpKey(ctx, r.Status.SecretKeyRef.Name)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	agentCon := AgentConn{
		ctx:     ctx,
		sandbox: r.Status.SandboxName,
		key:     key,
	}
	agentCon.Relay()
}

//
// AgentWorkflow
//

// WorkflowGet godoc
// @summary Get an agent workflow by name.
// @description Get an agent workflow by name.
// @tags workflows
// @produce json
// @success 200 {object} AgentWorkflow
// @router /agentic/workflows/{name} [get]
// @param name path string true "AgentWorkflow name"
func (h AgenticHandler) WorkflowGet(ctx *gin.Context) {
	r := &AgentWorkflow{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// WorkflowList godoc
// @summary List all agent workflows.
// @description List all agent workflows.
// @tags workflows
// @produce json
// @success 200 {object} []AgentWorkflow
// @router /agentic/workflows [get]
func (h AgenticHandler) WorkflowList(ctx *gin.Context) {
	list := &AgentWorkflowList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Hub.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// WorkflowCreate godoc
// @summary Create an agent workflow.
// @description Create an agent workflow.
// @tags workflows
// @accept json
// @produce json
// @success 201 {object} AgentWorkflow
// @router /agentic/workflows [post]
// @param workflow body AgentWorkflow true "AgentWorkflow data"
func (h AgenticHandler) WorkflowCreate(ctx *gin.Context) {
	r := &AgentWorkflow{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Namespace = Settings.Hub.Namespace
	err = h.Client(ctx).Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// WorkflowUpdate godoc
// @summary Update an agent workflow.
// @description Update an agent workflow.
// @tags workflows
// @accept json
// @success 204
// @router /agentic/workflows/{name} [put]
// @param name path string true "AgentWorkflow name"
// @param workflow body AgentWorkflow true "AgentWorkflow data"
func (h AgenticHandler) WorkflowUpdate(ctx *gin.Context) {
	current := &AgentWorkflow{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &AgentWorkflow{}
	err = h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	current.Spec = r.Spec
	err = h.Client(ctx).Update(context.TODO(), current)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// WorkflowDelete godoc
// @summary Delete an agent workflow.
// @description Delete an agent workflow.
// @tags workflows
// @success 204
// @router /agentic/workflows/{name} [delete]
// @param name path string true "AgentWorkflow name"
func (h AgenticHandler) WorkflowDelete(ctx *gin.Context) {
	r := &AgentWorkflow{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Client(ctx).Delete(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

//
// AgentWorkflowRun
//

// WorkflowRunGet godoc
// @summary Get an agent workflow run by name.
// @description Get an agent workflow run by name.
// @tags workflowruns
// @produce json
// @success 200 {object} AgentWorkflowRun
// @router /agentic/workflowruns/{name} [get]
// @param name path string true "AgentWorkflowRun name"
func (h AgenticHandler) WorkflowRunGet(ctx *gin.Context) {
	r := &AgentWorkflowRun{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      ctx.Param(Name),
		},
		r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, r)
}

// WorkflowRunList godoc
// @summary List all agent workflow runs.
// @description List all agent workflow runs.
// @tags workflowruns
// @produce json
// @success 200 {object} []AgentWorkflowRun
// @router /agentic/workflowruns [get]
func (h AgenticHandler) WorkflowRunList(ctx *gin.Context) {
	list := &AgentWorkflowRunList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		k8s.InNamespace(Settings.Hub.Namespace),
		k8s.MatchingLabels{
			ManagedLabel: "true",
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusOK, list.Items)
}

// WorkflowRunCreate godoc
// @summary Create an agent workflow run.
// @description Create an agent workflow run.
// @tags workflowruns
// @accept json
// @produce json
// @success 201 {object} AgentWorkflowRun
// @router /agentic/workflowruns [post]
// @param run body AgentWorkflowRun true "AgentWorkflowRun data"
func (h *AgenticHandler) WorkflowRunCreate(ctx *gin.Context) {
	r := &AgentWorkflowRun{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	client := h.Client(ctx)
	secret, tokenId, err := h.tokenSecret(r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = client.Create(context.TODO(), secret)
	if err != nil {
		_ = auth.Idp().Revoke(tokenId)
		_ = ctx.Error(err)
		return
	}
	defer func() {
		if err != nil {
			_ = auth.Idp().Revoke(tokenId)
			_ = client.Delete(context.TODO(), secret)
		}
	}()
	h.injectLabels(r)
	r.Namespace = Settings.Hub.Namespace
	r.Spec.EnvFrom = append(
		r.Spec.EnvFrom,
		core.EnvFromSource{
			SecretRef: &core.SecretEnvSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: secret.Name,
				},
			},
		})
	r.Spec.Env = append(
		r.Spec.Env,
		core.EnvVar{
			Name:  "HUB_BASE_URL",
			Value: Settings.Addon.Hub.URL,
		})
	err = client.Create(context.TODO(), r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	secret.OwnerReferences = []v1.OwnerReference{
		{
			APIVersion: "v1alpha1",
			Kind:       "AgentWorkflowRun",
			Name:       r.Name,
			UID:        r.UID,
		},
	}
	err = client.Update(context.TODO(), secret)
	if err != nil {
		_ = client.Delete(context.TODO(), r)
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// acpKey reads the ACP secret key from a Secret.
func (h *AgenticHandler) acpKey(ctx *gin.Context, name string) (key string, err error) {
	secret := &core.Secret{}
	err = h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		secret)
	if err != nil {
		return
	}
	for _, k := range ACPSecretKeys {
		if v, found := secret.Data[k]; found {
			key = string(v)
			return
		}
	}
	if len(secret.Data) == 1 {
		for _, v := range secret.Data {
			key = string(v)
			return
		}
	}
	err = &BadRequestError{Reason: "ACP secret key not found."}
	return
}

// injectLabels inject labels.
func (h *AgenticHandler) injectLabels(r k8s.Object) {
	m := r.GetLabels()
	if m == nil {
		m = make(map[string]string)
		r.SetLabels(m)
	}
	m[ManagedLabel] = "true"
}

// tokenSecret returns a token secret.
func (h *AgenticHandler) tokenSecret(owner k8s.Object) (secret *core.Secret, tokenId uint, err error) {
	idp := auth.Idp()
	cache := idp.Cache()
	sa, err := cache.FindSaByName("agentic.harness")
	if err != nil {
		return
	}
	token, err := idp.NewToken(
		sa.Subject,
		time.Hour*24,
		func(m *auth.Token) {
			m.Description = fmt.Sprintf(
				"%s.%s",
				owner.GetObjectKind().GroupVersionKind().Kind,
				owner.GetName())
		})
	if err != nil {
		return
	}
	tokenId = token.ID
	secret = &core.Secret{}
	secret.Namespace = Settings.Namespace
	secret.GenerateName = "agentic-run-"
	secret.StringData = map[string]string{
		"HUB_TOKEN_ID": strconv.FormatUint(uint64(token.ID), 10),
		"HUB_TOKEN":    token.Secret,
	}
	return
}

//
// ACP (Agent Connection Protocol)
//

// AgentConn provides a WebSocket upgrade and frame relay.
type AgentConn struct {
	ctx     *gin.Context
	sandbox string
	key     string
}

// Relay upgrades the client connection and relays frames.
func (r *AgentConn) Relay() {
	u := r.wsURL()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
	client, err := upgrader.Upgrade(r.ctx.Writer, r.ctx.Request, nil)
	if err != nil {
		log.Error(
			err,
			"WebSocket upgrade failed.",
			"URL",
			u)
		return
	}
	defer func() {
		_ = client.Close()
	}()
	header := http.Header{
		"X-Secret-Key": {r.key},
	}
	remote, _, err := websocket.DefaultDialer.Dial(
		u,
		header)
	if err != nil {
		log.Error(
			err,
			"ACP connection failed.",
			"URL",
			u)
		return
	}
	defer func() {
		_ = remote.Close()
	}()
	done := make(chan struct{})
	go r.relay(done, remote, client)
	go r.relay(done, client, remote)
	<-done
}

// wsURL returns the websocket URL.
func (r *AgentConn) wsURL() (u string) {
	ns := Settings.Hub.Namespace
	scheme := strings.ToLower(r.ctx.Request.URL.Scheme)
	switch scheme {
	case "https":
		scheme = "wss"
	default:
		scheme = "ws"
	}
	u = fmt.Sprintf(
		"%s://%s.%s.svc:%d/acp",
		scheme,
		r.sandbox,
		ns,
		ACPPort)
	return
}

// relay frames until either connection closed.
func (r *AgentConn) relay(done chan struct{}, input, output *websocket.Conn) {
	defer func() {
		recover()
	}()
	defer close(done)
	defer func() {
		_ = input.Close()
	}()
	defer func() {
		_ = output.Close()
	}()
	for {
		mt, msg, err := input.ReadMessage()
		if err != nil {
			return
		}
		err = output.WriteMessage(mt, msg)
		if err != nil {
			return
		}
	}
}

//
// REST resource aliases
//

// Agent CR type.
type Agent = agent.Agent

// AgentList CR type.
type AgentList = agent.AgentList

// SkillCard CR type.
type SkillCard = agent.SkillCard

// SkillCardList CR type.
type SkillCardList = agent.SkillCardList

// SkillCollection CR type.
type SkillCollection = agent.SkillCollection

// SkillCollectionList CR type.
type SkillCollectionList = agent.SkillCollectionList

// Gateway CR type.
type Gateway = agent.Gateway

// GatewayList CR type.
type GatewayList = agent.GatewayList

// AgentRun CR type.
type AgentRun = agent.AgentRun

// AgentRunList CR type.
type AgentRunList = agent.AgentRunList

// AgentWorkflow CR type.
type AgentWorkflow = agent.AgentWorkflow

// AgentWorkflowList CR type.
type AgentWorkflowList = agent.AgentWorkflowList

// AgentWorkflowRun CR type.
type AgentWorkflowRun = agent.AgentWorkflowRun

// AgentWorkflowRunList CR type.
type AgentWorkflowRunList = agent.AgentWorkflowRunList

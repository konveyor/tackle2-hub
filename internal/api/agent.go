package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	agent "github.com/konveyor/agentic-controller/api/v1alpha1"
	"github.com/konveyor/tackle2-hub/shared/api"
	core "k8s.io/api/core/v1"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// ACPPort is the ACP endpoint port on the sandbox service.
const ACPPort = 4000

// ACPSecretKeys are the Secret data keys tried when reading the ACP secret.
var ACPSecretKeys = []string{"secret-key", "ACP_SECRET_KEY"}

// AgentHandler handles AI agent routes.
type AgentHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AgentHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Authenticate())
	// Agent
	routeGroup.GET(api.AgentAgentsRoute, h.AgentList)
	routeGroup.GET(api.AgentAgentsRoute+"/", h.AgentList)
	routeGroup.POST(api.AgentAgentsRoute, h.AgentCreate)
	routeGroup.GET(api.AgentAgentRoute, h.AgentGet)
	routeGroup.PUT(api.AgentAgentRoute, h.AgentUpdate)
	routeGroup.DELETE(api.AgentAgentRoute, h.AgentDelete)
	// SkillCard
	routeGroup.GET(api.AgentSkillsRoute, h.SkillList)
	routeGroup.GET(api.AgentSkillsRoute+"/", h.SkillList)
	routeGroup.POST(api.AgentSkillsRoute, h.SkillCreate)
	routeGroup.GET(api.AgentSkillRoute, h.SkillGet)
	routeGroup.PUT(api.AgentSkillRoute, h.SkillUpdate)
	routeGroup.DELETE(api.AgentSkillRoute, h.SkillDelete)
	// SkillCollection
	routeGroup.GET(api.AgentSkillCollectionsRoute, h.SkillCollectionList)
	routeGroup.GET(api.AgentSkillCollectionsRoute+"/", h.SkillCollectionList)
	routeGroup.POST(api.AgentSkillCollectionsRoute, h.SkillCollectionCreate)
	routeGroup.GET(api.AgentSkillCollectionRoute, h.SkillCollectionGet)
	routeGroup.PUT(api.AgentSkillCollectionRoute, h.SkillCollectionUpdate)
	routeGroup.DELETE(api.AgentSkillCollectionRoute, h.SkillCollectionDelete)
	// Gateway
	routeGroup.GET(api.AgentGatewaysRoute, h.GatewayList)
	routeGroup.GET(api.AgentGatewaysRoute+"/", h.GatewayList)
	routeGroup.POST(api.AgentGatewaysRoute, h.GatewayCreate)
	routeGroup.GET(api.AgentGatewayRoute, h.GatewayGet)
	routeGroup.PUT(api.AgentGatewayRoute, h.GatewayUpdate)
	routeGroup.DELETE(api.AgentGatewayRoute, h.GatewayDelete)
	// AgentRun
	routeGroup.GET(api.AgentRunsRoute, h.RunList)
	routeGroup.GET(api.AgentRunsRoute+"/", h.RunList)
	routeGroup.POST(api.AgentRunsRoute, h.RunCreate)
	routeGroup.GET(api.AgentRunRoute, h.RunGet)
	routeGroup.GET(api.AgentRunACPRoute, h.RunACP)
	// AgentWorkflow
	routeGroup.GET(api.AgentWorkflowsRoute, h.WorkflowList)
	routeGroup.GET(api.AgentWorkflowsRoute+"/", h.WorkflowList)
	routeGroup.POST(api.AgentWorkflowsRoute, h.WorkflowCreate)
	routeGroup.GET(api.AgentWorkflowRoute, h.WorkflowGet)
	routeGroup.PUT(api.AgentWorkflowRoute, h.WorkflowUpdate)
	routeGroup.DELETE(api.AgentWorkflowRoute, h.WorkflowDelete)
	// AgentWorkflowRun
	routeGroup.GET(api.AgentWorkflowRunsRoute, h.WorkflowRunList)
	routeGroup.GET(api.AgentWorkflowRunsRoute+"/", h.WorkflowRunList)
	routeGroup.POST(api.AgentWorkflowRunsRoute, h.WorkflowRunCreate)
	routeGroup.GET(api.AgentWorkflowRunRoute, h.WorkflowRunGet)
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
// @router /agent/agents/{name} [get]
// @param name path string true "Agent name"
func (h AgentHandler) AgentGet(ctx *gin.Context) {
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
// @router /agent/agents [get]
func (h AgentHandler) AgentList(ctx *gin.Context) {
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
// @router /agent/agents [post]
// @param agent body Agent true "Agent data"
func (h AgentHandler) AgentCreate(ctx *gin.Context) {
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
// @router /agent/agents/{name} [put]
// @param name path string true "Agent name"
// @param agent body Agent true "Agent data"
func (h AgentHandler) AgentUpdate(ctx *gin.Context) {
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
// @router /agent/agents/{name} [delete]
// @param name path string true "Agent name"
func (h AgentHandler) AgentDelete(ctx *gin.Context) {
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
// @router /agent/skills/{name} [get]
// @param name path string true "SkillCard name"
func (h AgentHandler) SkillGet(ctx *gin.Context) {
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
// @router /agent/skills [get]
func (h AgentHandler) SkillList(ctx *gin.Context) {
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
// @router /agent/skills [post]
// @param skill body SkillCard true "SkillCard data"
func (h AgentHandler) SkillCreate(ctx *gin.Context) {
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
// @router /agent/skills/{name} [put]
// @param name path string true "SkillCard name"
// @param skill body SkillCard true "SkillCard data"
func (h AgentHandler) SkillUpdate(ctx *gin.Context) {
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
// @router /agent/skills/{name} [delete]
// @param name path string true "SkillCard name"
func (h AgentHandler) SkillDelete(ctx *gin.Context) {
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
// @router /agent/skillcollections/{name} [get]
// @param name path string true "SkillCollection name"
func (h AgentHandler) SkillCollectionGet(ctx *gin.Context) {
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
// @router /agent/skillcollections [get]
func (h AgentHandler) SkillCollectionList(ctx *gin.Context) {
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
// @router /agent/skillcollections [post]
// @param collection body SkillCollection true "SkillCollection data"
func (h AgentHandler) SkillCollectionCreate(ctx *gin.Context) {
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
// @router /agent/skillcollections/{name} [put]
// @param name path string true "SkillCollection name"
// @param collection body SkillCollection true "SkillCollection data"
func (h AgentHandler) SkillCollectionUpdate(ctx *gin.Context) {
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
// @router /agent/skillcollections/{name} [delete]
// @param name path string true "SkillCollection name"
func (h AgentHandler) SkillCollectionDelete(ctx *gin.Context) {
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
// @router /agent/gateways/{name} [get]
// @param name path string true "Gateway name"
func (h AgentHandler) GatewayGet(ctx *gin.Context) {
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
// @router /agent/gateways [get]
func (h AgentHandler) GatewayList(ctx *gin.Context) {
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
// @router /agent/gateways [post]
// @param gateway body Gateway true "Gateway data"
func (h AgentHandler) GatewayCreate(ctx *gin.Context) {
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
// @router /agent/gateways/{name} [put]
// @param name path string true "Gateway name"
// @param gateway body Gateway true "Gateway data"
func (h AgentHandler) GatewayUpdate(ctx *gin.Context) {
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
// @router /agent/gateways/{name} [delete]
// @param name path string true "Gateway name"
func (h AgentHandler) GatewayDelete(ctx *gin.Context) {
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
// @router /agent/runs/{name} [get]
// @param name path string true "AgentRun name"
func (h AgentHandler) RunGet(ctx *gin.Context) {
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
// @router /agent/runs [get]
func (h AgentHandler) RunList(ctx *gin.Context) {
	list := &AgentRunList{}
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

// RunCreate godoc
// @summary Create an agent run.
// @description Create an agent run.
// @tags runs
// @accept json
// @produce json
// @success 201 {object} AgentRun
// @router /agent/runs [post]
// @param run body AgentRun true "AgentRun data"
func (h AgentHandler) RunCreate(ctx *gin.Context) {
	r := &AgentRun{}
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

// RunACP godoc
// @summary WebSocket proxy to agent pod ACP endpoint.
// @description Upgrades the connection to a WebSocket and proxies frames
// @description bidirectionally to the agent pod's ACP endpoint.
// @tags runs
// @router /agent/runs/{name}/acp [get]
// @param name path string true "AgentRun name"
func (h AgentHandler) RunACP(ctx *gin.Context) {
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
		h.Respond(
			ctx,
			http.StatusServiceUnavailable,
			gin.H{"error": "ACP endpoint not ready."})
		return
	}
	secretKey, err := h.acpSecretKey(ctx, r.Status.SecretKeyRef.Name)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	agentCon := AgentConn{
		ctx: ctx,
		sandboxURL: fmt.Sprintf(
			"ws://%s.%s.svc:%d/acp",
			r.Status.SandboxName,
			Settings.Hub.Namespace,
			ACPPort),
		secretKey: secretKey,
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
// @router /agent/workflows/{name} [get]
// @param name path string true "AgentWorkflow name"
func (h AgentHandler) WorkflowGet(ctx *gin.Context) {
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
// @router /agent/workflows [get]
func (h AgentHandler) WorkflowList(ctx *gin.Context) {
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
// @router /agent/workflows [post]
// @param workflow body AgentWorkflow true "AgentWorkflow data"
func (h AgentHandler) WorkflowCreate(ctx *gin.Context) {
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
// @router /agent/workflows/{name} [put]
// @param name path string true "AgentWorkflow name"
// @param workflow body AgentWorkflow true "AgentWorkflow data"
func (h AgentHandler) WorkflowUpdate(ctx *gin.Context) {
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
// @router /agent/workflows/{name} [delete]
// @param name path string true "AgentWorkflow name"
func (h AgentHandler) WorkflowDelete(ctx *gin.Context) {
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
// @router /agent/workflowruns/{name} [get]
// @param name path string true "AgentWorkflowRun name"
func (h AgentHandler) WorkflowRunGet(ctx *gin.Context) {
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
// @router /agent/workflowruns [get]
func (h AgentHandler) WorkflowRunList(ctx *gin.Context) {
	list := &AgentWorkflowRunList{}
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

// WorkflowRunCreate godoc
// @summary Create an agent workflow run.
// @description Create an agent workflow run.
// @tags workflowruns
// @accept json
// @produce json
// @success 201 {object} AgentWorkflowRun
// @router /agent/workflowruns [post]
// @param run body AgentWorkflowRun true "AgentWorkflowRun data"
func (h AgentHandler) WorkflowRunCreate(ctx *gin.Context) {
	r := &AgentWorkflowRun{}
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

// acpSecretKey reads the ACP secret key from a Secret.
func (h AgentHandler) acpSecretKey(ctx *gin.Context, name string) (key string, err error) {
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

//
// ACP (Agent Connection Protocol)
//

// AgentConn provides a WebSocket upgrade and frame relay.
type AgentConn struct {
	ctx        *gin.Context
	sandboxURL string
	secretKey  string
}

// Relay upgrades the client connection and relays frames.
func (r *AgentConn) Relay() {
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
			r.sandboxURL)
		return
	}
	defer func() {
		_ = client.Close()
	}()
	header := http.Header{
		"X-Secret-Key": {r.secretKey},
	}
	remote, _, err := websocket.DefaultDialer.Dial(
		r.sandboxURL,
		header)
	if err != nil {
		log.Error(
			err,
			"ACP connection failed.",
			"URL",
			r.sandboxURL)
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

package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// Effort estimates
const (
	EffortS  = "small"
	EffortM  = "medium"
	EffortL  = "large"
	EffortXL = "extra_large"
)

type AdoptionPlanHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AdoptionPlanHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("adoptionplans"))
	routeGroup.POST(api.AdoptionPlansRoute, h.Graph)
}

// Graph godoc
// @summary Generate an application dependency graph arranged in topological order.
// @description Graph generates an application dependency graph arranged in topological order.
// @tags adoptionplans
// @produce json
// @success 200 {object} []api.Vertex
// @router /adoptionplans [post]
// @param requestedApps body []uint true "List of requested App IDs"
func (h AdoptionPlanHandler) Graph(ctx *gin.Context) {
	var requestedApps []struct {
		ApplicationID uint `json:"applicationId"`
	}

	err := h.Bind(ctx, &requestedApps)
	if err != nil {
		_ = ctx.Error(err)
	}

	var ids []uint
	for _, a := range requestedApps {
		ids = append(ids, a.ApplicationID)
	}

	var reviews []model.Review
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Where("applicationId IN ?", ids).Find(&reviews)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	var deps []model.Dependency
	result = h.DB(ctx).Where("toId IN ? AND fromId IN ?", ids, ids).Find(&deps)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	graph := NewDependencyGraph()
	for i := range reviews {
		review := &reviews[i]
		vertex := Vertex{
			ID:             *review.ApplicationID,
			Name:           review.Application.Name,
			Decision:       review.ProposedAction,
			EffortEstimate: review.EffortEstimate,
			Effort:         numericEffort(review.EffortEstimate),
			PositionY:      int(review.WorkPriority),
		}
		graph.AddVertex(&vertex)
	}

	for i := range deps {
		v := deps[i].FromID
		w := deps[i].ToID
		if graph.HasVertex(v) && graph.HasVertex(w) {
			graph.AddEdge(v, w)
		}
	}

	sorted, ok := graph.TopologicalSort()
	if !ok {
		h.Respond(ctx,
			http.StatusBadRequest,
			gin.H{
				"error": "dependency cycle detected",
			})
		return
	}

	h.Respond(ctx, http.StatusOK, sorted)
}

// Vertex represents a vertex in the dependency graph.
type Vertex = resource.Vertex

// NewDependencyGraph creates an empty dependency graph.
func NewDependencyGraph() (graph DependencyGraph) {
	graph.vertices = make(map[uint]*Vertex)
	graph.edges = make(map[uint][]uint)
	graph.incoming = make(map[uint][]uint)
	graph.costs = make(map[uint]int)
	return
}

// DependencyGraph is an application dependency graph.
type DependencyGraph struct {
	// all applications
	vertices map[uint]*Vertex
	// ids of all applications a given application depends on
	edges map[uint][]uint
	// ids of all applications depending on a given application
	incoming map[uint][]uint
	// memoized total costs
	costs map[uint]int
}

// AddVertex adds a vertex to the graph.
func (r *DependencyGraph) AddVertex(v *Vertex) {
	r.vertices[v.ID] = v
}

// HasVertex checks for the presence of a vertex in the graph.
func (r *DependencyGraph) HasVertex(v uint) (ok bool) {
	_, ok = r.vertices[v]
	return
}

// AddEdge adds an edge between two vertices.
func (r *DependencyGraph) AddEdge(v, w uint) {
	r.edges[v] = append(r.edges[v], w)
	r.incoming[w] = append(r.incoming[w], v)
}

// CalculateCost calculates the total cost to reach a given vertex.
// Costs are memoized.
func (r *DependencyGraph) CalculateCost(v uint) (cumulativeCost int) {
	if len(r.edges[v]) == 0 {
		return
	}
	if cost, ok := r.costs[v]; ok {
		cumulativeCost = cost
		return
	}
	var cost int
	for _, id := range r.edges[v] {
		w := r.vertices[id]
		cost = w.Effort + r.CalculateCost(w.ID)
		if cost > cumulativeCost {
			cumulativeCost = cost
		}
	}
	r.costs[v] = cumulativeCost

	return
}

// TopologicalSort sorts the graph such that the vertices
// with fewer dependencies are first, and detects cycles.
func (r *DependencyGraph) TopologicalSort() (sorted []*Vertex, ok bool) {
	numEdges := make(map[uint]int)

	for _, v := range r.vertices {
		edges := len(r.edges[v.ID])
		numEdges[v.ID] = edges
		if edges == 0 {
			sorted = append(sorted, v)
		}
	}

	for i := 0; i < len(sorted); i++ {
		v := sorted[i]
		v.PositionY = i
		for _, w := range r.incoming[v.ID] {
			numEdges[w] -= 1
			if numEdges[w] == 0 {
				sorted = append(sorted, r.vertices[w])
			}
		}
	}

	// cycle detected
	if len(sorted) < len(r.vertices) {
		return
	}

	// calculate effort for each application
	for _, v := range r.vertices {
		v.PositionX = r.CalculateCost(v.ID)
	}

	ok = true
	return
}

func numericEffort(effort string) (cost int) {
	switch strings.ToLower(effort) {
	case EffortS:
		cost = 1
	case EffortM:
		cost = 2
	case EffortL:
		cost = 4
	case EffortXL:
		cost = 8
	}
	return
}

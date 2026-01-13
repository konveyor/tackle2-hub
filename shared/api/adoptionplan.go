package api

// Vertex represents a vertex in the dependency graph.
type Vertex struct {
	ID             uint   `json:"applicationId" yaml:"applicationId"`
	Name           string `json:"applicationName" yaml:"applicationName"`
	Decision       string `json:"decision"`
	EffortEstimate string `json:"effortEstimate" yaml:"effortEstimate"`
	Effort         int    `json:"effort"`
	PositionY      int    `json:"positionY" yaml:"positionY"`
	PositionX      int    `json:"positionX" yaml:"positionX"`
}

// DependencyGraph is an application dependency graph.
type DependencyGraph struct {
	// all applications
	Vertices map[uint]*Vertex
	// ids of all applications a given application depends on
	Edges map[uint][]uint
	// ids of all applications depending on a given application
	Incoming map[uint][]uint
	// memoized total costs
	Costs map[uint]int
}

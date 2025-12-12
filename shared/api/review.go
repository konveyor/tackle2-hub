package api

// Review REST resource.
type Review struct {
	Resource            `yaml:",inline"`
	BusinessCriticality uint   `json:"businessCriticality" yaml:"businessCriticality"`
	EffortEstimate      string `json:"effortEstimate" yaml:"effortEstimate"`
	ProposedAction      string `json:"proposedAction" yaml:"proposedAction"`
	WorkPriority        uint   `json:"workPriority" yaml:"workPriority"`
	Comments            string `json:"comments"`
	Application         *Ref   `json:"application,omitempty" binding:"required_without=Archetype,excluded_with=Archetype"`
	Archetype           *Ref   `json:"archetype,omitempty" binding:"required_without=Application,excluded_with=Application"`
}

// CopyRequest REST resource.
type CopyRequest struct {
	SourceReview       uint   `json:"sourceReview" binding:"required"`
	TargetApplications []uint `json:"targetApplications" binding:"required"`
}

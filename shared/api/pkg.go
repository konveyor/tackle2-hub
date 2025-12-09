package api

// Params
const (
	ID        = "id"
	ID2       = "id2"
	Key       = "key"
	Name      = "name"
	Wildcard  = "wildcard"
	FileField = "file"
	Decrypted = "decrypted"
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

// MIME Types
const (
	MIMEOCTETSTREAM = "application/octet-stream"
	MIMEJSON        = "application/json"
	MIMEYAML        = "application/x-yaml"
)

// Header Values
const (
	DirectoryExpand = "expand"
)

// Schema Params
const (
	Domain  = "domain"
	Variant = "variant"
	Subject = "subject"
)

// Application Params
const (
	Source = "source"
)

// Manifest Params
const (
	Injected = "injected"
)

// Routes - Addons
const (
	AddonsRoot = "/addons"
	AddonRoot  = AddonsRoot + "/:" + Name
)

// Routes - Adoption Plans
const (
	AdoptionPlansRoot = "/reports/adoptionplan"
)

// Routes - Analysis
const (
	AnalysesRoot          = "/analyses"
	AnalysisRoot          = AnalysesRoot + "/:" + ID
	AnalysisArchiveRoot   = AnalysisRoot + "/archive"
	AnalysisInsightsRoot  = AnalysisRoot + "/insights"
	AnalysisIncidentsRoot = AnalysesInsightRoot + "/incidents"
	AnalysesDepsRoot      = AnalysesRoot + "/dependencies"
	AnalysesInsightsRoot  = AnalysesRoot + "/insights"
	AnalysesInsightRoot   = AnalysesInsightsRoot + "/:" + ID
	AnalysesIncidentsRoot = AnalysesRoot + "/incidents"
	AnalysesIncidentRoot  = AnalysesIncidentsRoot + "/:" + ID

	AnalysesReportRoot             = AnalysesRoot + "/report"
	AnalysisReportDepsRoot         = AnalysesReportRoot + "/dependencies"
	AnalysisReportRuleRoot         = AnalysesReportRoot + "/rules"
	AnalysisReportInsightsRoot     = AnalysesReportRoot + "/insights"
	AnalysisReportAppsRoot         = AnalysesReportRoot + "/applications"
	AnalysisReportInsightRoot      = AnalysisReportInsightsRoot + "/:" + ID
	AnalysisReportInsightsAppsRoot = AnalysisReportInsightsRoot + "/applications"
	AnalysisReportDepsAppsRoot     = AnalysisReportDepsRoot + "/applications"
	AnalysisReportAppsInsightsRoot = AnalysisReportAppsRoot + "/:" + ID + "/insights"
	AnalysisReportFileRoot         = AnalysisReportInsightRoot + "/files"

	AppAnalysesRoot         = ApplicationRoot + "/analyses"
	AppAnalysisRoot         = ApplicationRoot + "/analysis"
	AppAnalysisReportRoot   = AppAnalysisRoot + "/report"
	AppAnalysisDepsRoot     = AppAnalysisRoot + "/dependencies"
	AppAnalysisInsightsRoot = AppAnalysisRoot + "/insights"
)

// Routes - Analysis Profiles
const (
	AnalysisProfilesRoot = "/analysis/profiles"
	AnalysisProfileRoot  = AnalysisProfilesRoot + "/:id"

	AppAnalysisProfilesRoot = ApplicationRoot + "/analysis/profiles"
)

// Routes - Applications
const (
	ApplicationsRoot     = "/applications"
	ApplicationRoot      = ApplicationsRoot + "/:" + ID
	ApplicationTagsRoot  = ApplicationRoot + "/tags"
	ApplicationTagRoot   = ApplicationTagsRoot + "/:" + ID2
	ApplicationFactsRoot = ApplicationRoot + "/facts"
	ApplicationFactRoot  = ApplicationFactsRoot + "/:" + Key
	AppBucketRoot        = ApplicationRoot + "/bucket"
	AppBucketContentRoot = AppBucketRoot + "/*" + Wildcard
	AppStakeholdersRoot  = ApplicationRoot + "/stakeholders"
	AppAssessmentsRoot   = ApplicationRoot + "/assessments"
	AppAssessmentRoot    = AppAssessmentsRoot + "/:" + ID2
)

// Routes - Archetypes
const (
	ArchetypesRoot           = "/archetypes"
	ArchetypeRoot            = ArchetypesRoot + "/:" + ID
	ArchetypeAssessmentsRoot = ArchetypeRoot + "/assessments"
)

// Routes - Assessments
const (
	AssessmentsRoot = "/assessments"
	AssessmentRoot  = AssessmentsRoot + "/:" + ID
)

// Routes - Auth
const (
	AuthRoot        = "/auth"
	AuthLoginRoot   = AuthRoot + "/login"
	AuthRefreshRoot = AuthRoot + "/refresh"
)

// Routes - Batch
const (
	BatchRoot        = "/batch"
	BatchTicketsRoot = BatchRoot + TicketsRoot
	BatchTagsRoot    = BatchRoot + TagsRoot
)

// Routes - Buckets
const (
	BucketsRoot       = "/buckets"
	BucketRoot        = BucketsRoot + "/:" + ID
	BucketContentRoot = BucketRoot + "/*" + Wildcard
)

// Routes - Business Services
const (
	BusinessServicesRoot = "/businessservices"
	BusinessServiceRoot  = BusinessServicesRoot + "/:" + ID
)

// Routes - Cache
const (
	CacheRoot    = "/cache"
	CacheDirRoot = CacheRoot + "/*" + Wildcard
)

// Routes - Config Maps
const (
	ConfigMapsRoot   = "/configmaps"
	ConfigMapRoot    = ConfigMapsRoot + "/:" + Name
	ConfigMapKeyRoot = ConfigMapRoot + "/:" + Key
)

// Routes - Dependencies
const (
	DependenciesRoot = "/dependencies"
	DependencyRoot   = DependenciesRoot + "/:" + ID
)

// Routes - Files
const (
	FilesRoot = "/files"
	FileRoot  = FilesRoot + "/:" + ID
)

// Routes - Generators
const (
	GeneratorsRoot = "/generators"
	GeneratorRoot  = GeneratorsRoot + "/:" + ID
)

// Routes - Identities
const (
	IdentitiesRoot = "/identities"
	IdentityRoot   = IdentitiesRoot + "/:" + ID

	AppIdentitiesRoot = ApplicationRoot + "/identities"
)

// Routes - Imports
const (
	SummariesRoot = "/importsummaries"
	SummaryRoot   = SummariesRoot + "/:" + ID
	UploadRoot    = SummariesRoot + "/upload"
	DownloadRoot  = SummariesRoot + "/download"
	ImportsRoot   = "/imports"
	ImportRoot    = ImportsRoot + "/:" + ID
)

// Routes - Job Functions
const (
	JobFunctionsRoot = "/jobfunctions"
	JobFunctionRoot  = JobFunctionsRoot + "/:" + ID
)

// Routes - Manifests
const (
	ManifestsRoot = "/manifests"
	ManifestRoot  = ManifestsRoot + "/:" + ID

	AppManifestRoot  = ApplicationRoot + "/manifest"
	AppManifestsRoot = ApplicationRoot + "/manifests"
)

// Routes - Migration Waves
const (
	MigrationWavesRoot = "/migrationwaves"
	MigrationWaveRoot  = MigrationWavesRoot + "/:" + ID
)

// Routes - Platforms
const (
	PlatformsRoot = "/platforms"
	PlatformRoot  = PlatformsRoot + "/:" + ID
)

// Routes - Proxies
const (
	ProxiesRoot = "/proxies"
	ProxyRoot   = ProxiesRoot + "/:" + ID
)

// Routes - Questionnaires
const (
	QuestionnairesRoot = "/questionnaires"
	QuestionnaireRoot  = QuestionnairesRoot + "/:" + ID
)

// Routes - Reviews
const (
	ReviewsRoot = "/reviews"
	ReviewRoot  = ReviewsRoot + "/:" + ID
	CopyRoot    = ReviewsRoot + "/copy"
)

// Routes - Rule Sets
const (
	RuleSetsRoot = "/rulesets"
	RuleSetRoot  = RuleSetsRoot + "/:" + ID
)

// Routes - Schemas
const (
	SchemaRoot     = "/schema"
	SchemasRoot    = "/schemas"
	SchemasGetRoot = SchemasRoot + "/:" + Name
	SchemaFindRoot = SchemaRoot + "/jsd/:" + Domain + "/:" + Variant + "/:" + Subject
)

// Routes - Services
const (
	ServicesRoot      = "/services"
	ServiceRoot       = ServicesRoot + "/:name"
	ServiceNestedRoot = ServiceRoot + "/*" + Wildcard
)

// Routes - Settings
const (
	SettingsRoot = "/settings"
	SettingRoot  = SettingsRoot + "/:" + Key
)

// Routes - Stakeholders
const (
	StakeholdersRoot = "/stakeholders"
	StakeholderRoot  = StakeholdersRoot + "/:" + ID
)

// Routes - Stakeholder Groups
const (
	StakeholderGroupsRoot = "/stakeholdergroups"
	StakeholderGroupRoot  = StakeholderGroupsRoot + "/:" + ID
)

// Routes - Tags
const (
	TagsRoot = "/tags"
	TagRoot  = TagsRoot + "/:" + ID
)

// Routes - Tag Categories
const (
	TagCategoriesRoot   = "/tagcategories"
	TagCategoryRoot     = TagCategoriesRoot + "/:" + ID
	TagCategoryTagsRoot = TagCategoryRoot + "/tags"
)

// Routes - Targets
const (
	TargetsRoot = "/targets"
	TargetRoot  = TargetsRoot + "/:" + ID
)

// Routes - Tasks
const (
	TasksRoot                = "/tasks"
	TasksReportRoot          = TasksRoot + "/report"
	TasksReportQueueRoot     = TasksReportRoot + "/queue"
	TasksReportDashboardRoot = TasksReportRoot + "/dashboard"
	TasksCancelRoot          = TasksRoot + "/cancel"
	TaskRoot                 = TasksRoot + "/:" + ID
	TaskReportRoot           = TaskRoot + "/report"
	TaskAttachedRoot         = TaskRoot + "/attached"
	TaskBucketRoot           = TaskRoot + "/bucket"
	TaskBucketContentRoot    = TaskBucketRoot + "/*" + Wildcard
	TaskSubmitRoot           = TaskRoot + "/submit"
	TaskCancelRoot           = TaskRoot + "/cancel"
)

// Routes - Task Groups
const (
	TaskGroupsRoot             = "/taskgroups"
	TaskGroupRoot              = TaskGroupsRoot + "/:" + ID
	TaskGroupBucketRoot        = TaskGroupRoot + "/bucket"
	TaskGroupBucketContentRoot = TaskGroupBucketRoot + "/*" + Wildcard
	TaskGroupSubmitRoot        = TaskGroupRoot + "/submit"
)

// Routes - Tickets
const (
	TicketsRoot = "/tickets"
	TicketRoot  = TicketsRoot + "/:" + ID
)

// Routes - Trackers
const (
	TrackersRoot             = "/trackers"
	TrackerRoot              = "/trackers" + "/:" + ID
	TrackerProjects          = TrackerRoot + "/projects"
	TrackerProject           = TrackerRoot + "/projects" + "/:" + ID2
	TrackerProjectIssueTypes = TrackerProject + "/issuetypes"
)

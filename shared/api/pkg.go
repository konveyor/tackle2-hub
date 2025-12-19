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
	Accept          = "Accept"
	Authorization   = "Authorization"
	ContentType     = "Content-Type"
	Directory       = "X-Directory"
	DirectoryExpand = "expand"
	Total           = "X-Total"
)

// MIME Types
const (
	MIMEOCTETSTREAM = "application/octet-stream"
	MIMEJSON        = "application/json"
	MIMEYAML        = "application/x-yaml"
	TAR             = "application/x-tar"
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
	AddonsRoute = "/addons"
	AddonRoute  = AddonsRoute + "/:" + Name
)

// Routes - Adoption Plans
const (
	AdoptionPlansRoute = "/reports/adoptionplan"
)

// Routes - Analysis
const (
	AnalysesRoute          = "/analyses"
	AnalysisRoute          = AnalysesRoute + "/:" + ID
	AnalysisArchiveRoute   = AnalysisRoute + "/archive"
	AnalysisInsightsRoute  = AnalysisRoute + "/insights"
	AnalysisIncidentsRoute = AnalysesInsightRoute + "/incidents"
	AnalysesDepsRoute      = AnalysesRoute + "/dependencies"
	AnalysesInsightsRoute  = AnalysesRoute + "/insights"
	AnalysesInsightRoute   = AnalysesInsightsRoute + "/:" + ID
	AnalysesIncidentsRoute = AnalysesRoute + "/incidents"
	AnalysesIncidentRoute  = AnalysesIncidentsRoute + "/:" + ID

	AnalysesReportRoute             = AnalysesRoute + "/report"
	AnalysisReportDepsRoute         = AnalysesReportRoute + "/dependencies"
	AnalysisReportRuleRoute         = AnalysesReportRoute + "/rules"
	AnalysisReportInsightsRoute     = AnalysesReportRoute + "/insights"
	AnalysisReportAppsRoute         = AnalysesReportRoute + "/applications"
	AnalysisReportInsightRoute      = AnalysisReportInsightsRoute + "/:" + ID
	AnalysisReportInsightsAppsRoute = AnalysisReportInsightsRoute + "/applications"
	AnalysisReportDepsAppsRoute     = AnalysisReportDepsRoute + "/applications"
	AnalysisReportAppsInsightsRoute = AnalysisReportAppsRoute + "/:" + ID + "/insights"
	AnalysisReportFileRoute         = AnalysisReportInsightRoute + "/files"

	AppAnalysesRoute         = ApplicationRoute + "/analyses"
	AppAnalysisRoute         = ApplicationRoute + "/analysis"
	AppAnalysisReportRoute   = AppAnalysisRoute + "/report"
	AppAnalysisDepsRoute     = AppAnalysisRoute + "/dependencies"
	AppAnalysisInsightsRoute = AppAnalysisRoute + "/insights"
)

// Routes - Analysis Profiles
const (
	AnalysisProfilesRoute = "/analysis/profiles"
	AnalysisProfileRoute  = AnalysisProfilesRoute + "/:id"
	AnalysisProfileBundle = AnalysisProfileRoute + "/bundle"

	AppAnalysisProfilesRoute = ApplicationRoute + "/analysis/profiles"
)

// Routes - Applications
const (
	ApplicationsRoute     = "/applications"
	ApplicationRoute      = ApplicationsRoute + "/:" + ID
	ApplicationTagsRoute  = ApplicationRoute + "/tags"
	ApplicationTagRoute   = ApplicationTagsRoute + "/:" + ID2
	ApplicationFactsRoute = ApplicationRoute + "/facts"
	ApplicationFactRoute  = ApplicationFactsRoute + "/:" + Key
	AppBucketRoute        = ApplicationRoute + "/bucket"
	AppBucketContentRoute = AppBucketRoute + "/*" + Wildcard
	AppStakeholdersRoute  = ApplicationRoute + "/stakeholders"
	AppAssessmentsRoute   = ApplicationRoute + "/assessments"
	AppAssessmentRoute    = AppAssessmentsRoute + "/:" + ID2
)

// Routes - Archetypes
const (
	ArchetypesRoute           = "/archetypes"
	ArchetypeRoute            = ArchetypesRoute + "/:" + ID
	ArchetypeAssessmentsRoute = ArchetypeRoute + "/assessments"
)

// Routes - Assessments
const (
	AssessmentsRoute = "/assessments"
	AssessmentRoute  = AssessmentsRoute + "/:" + ID
)

// Routes - Auth
const (
	AuthRoute        = "/auth"
	AuthLoginRoute   = AuthRoute + "/login"
	AuthRefreshRoute = AuthRoute + "/refresh"
)

// Routes - Batch
const (
	BatchRoute        = "/batch"
	BatchTicketsRoute = BatchRoute + TicketsRoute
	BatchTagsRoute    = BatchRoute + TagsRoute
)

// Routes - Buckets
const (
	BucketsRoute       = "/buckets"
	BucketRoute        = BucketsRoute + "/:" + ID
	BucketContentRoute = BucketRoute + "/*" + Wildcard
)

// Routes - Business Services
const (
	BusinessServicesRoute = "/businessservices"
	BusinessServiceRoute  = BusinessServicesRoute + "/:" + ID
)

// Routes - Cache
const (
	CacheRoute    = "/cache"
	CacheDirRoute = CacheRoute + "/*" + Wildcard
)

// Routes - Config Maps
const (
	ConfigMapsRoute   = "/configmaps"
	ConfigMapRoute    = ConfigMapsRoute + "/:" + Name
	ConfigMapKeyRoute = ConfigMapRoute + "/:" + Key
)

// Routes - Dependencies
const (
	DependenciesRoute = "/dependencies"
	DependencyRoute   = DependenciesRoute + "/:" + ID
)

// Routes - Files
const (
	FilesRoute = "/files"
	FileRoute  = FilesRoute + "/:" + ID
)

// Routes - Generators
const (
	GeneratorsRoute = "/generators"
	GeneratorRoute  = GeneratorsRoute + "/:" + ID
)

// Routes - Identities
const (
	IdentitiesRoute = "/identities"
	IdentityRoute   = IdentitiesRoute + "/:" + ID

	AppIdentitiesRoute = ApplicationRoute + "/identities"
)

// Routes - Imports
const (
	SummariesRoute = "/importsummaries"
	SummaryRoute   = SummariesRoute + "/:" + ID
	UploadRoute    = SummariesRoute + "/upload"
	DownloadRoute  = SummariesRoute + "/download"
	ImportsRoute   = "/imports"
	ImportRoute    = ImportsRoute + "/:" + ID
)

// Routes - Job Functions
const (
	JobFunctionsRoute = "/jobfunctions"
	JobFunctionRoute  = JobFunctionsRoute + "/:" + ID
)

// Routes - Manifests
const (
	ManifestsRoute = "/manifests"
	ManifestRoute  = ManifestsRoute + "/:" + ID

	AppManifestRoute  = ApplicationRoute + "/manifest"
	AppManifestsRoute = ApplicationRoute + "/manifests"
)

// Routes - Migration Waves
const (
	MigrationWavesRoute = "/migrationwaves"
	MigrationWaveRoute  = MigrationWavesRoute + "/:" + ID
)

// Routes - Platforms
const (
	PlatformsRoute = "/platforms"
	PlatformRoute  = PlatformsRoute + "/:" + ID
)

// Routes - Proxies
const (
	ProxiesRoute = "/proxies"
	ProxyRoute   = ProxiesRoute + "/:" + ID
)

// Routes - Questionnaires
const (
	QuestionnairesRoute = "/questionnaires"
	QuestionnaireRoute  = QuestionnairesRoute + "/:" + ID
)

// Routes - Reviews
const (
	ReviewsRoute = "/reviews"
	ReviewRoute  = ReviewsRoute + "/:" + ID
	CopyRoute    = ReviewsRoute + "/copy"
)

// Routes - Rule Sets
const (
	RuleSetsRoute = "/rulesets"
	RuleSetRoute  = RuleSetsRoute + "/:" + ID
)

// Routes - Schemas
const (
	SchemaRoute     = "/schema"
	SchemasRoute    = "/schemas"
	SchemasGetRoute = SchemasRoute + "/:" + Name
	SchemaFindRoute = SchemaRoute + "/jsd/:" + Domain + "/:" + Variant + "/:" + Subject
)

// Routes - Services
const (
	ServicesRoute      = "/services"
	ServiceRoute       = ServicesRoute + "/:name"
	ServiceNestedRoute = ServiceRoute + "/*" + Wildcard
)

// Routes - Settings
const (
	SettingsRoute = "/settings"
	SettingRoute  = SettingsRoute + "/:" + Key
)

// Routes - Stakeholders
const (
	StakeholdersRoute = "/stakeholders"
	StakeholderRoute  = StakeholdersRoute + "/:" + ID
)

// Routes - Stakeholder Groups
const (
	StakeholderGroupsRoute = "/stakeholdergroups"
	StakeholderGroupRoute  = StakeholderGroupsRoute + "/:" + ID
)

// Routes - Tags
const (
	TagsRoute = "/tags"
	TagRoute  = TagsRoute + "/:" + ID
)

// Routes - Tag Categories
const (
	TagCategoriesRoute   = "/tagcategories"
	TagCategoryRoute     = TagCategoriesRoute + "/:" + ID
	TagCategoryTagsRoute = TagCategoryRoute + "/tags"
)

// Routes - Targets
const (
	TargetsRoute = "/targets"
	TargetRoute  = TargetsRoute + "/:" + ID
)

// Routes - Tasks
const (
	TasksRoute                = "/tasks"
	TasksReportRoute          = TasksRoute + "/report"
	TasksReportQueueRoute     = TasksReportRoute + "/queue"
	TasksReportDashboardRoute = TasksReportRoute + "/dashboard"
	TasksCancelRoute          = TasksRoute + "/cancel"
	TaskRoute                 = TasksRoute + "/:" + ID
	TaskReportRoute           = TaskRoute + "/report"
	TaskAttachedRoute         = TaskRoute + "/attached"
	TaskBucketRoute           = TaskRoute + "/bucket"
	TaskBucketContentRoute    = TaskBucketRoute + "/*" + Wildcard
	TaskSubmitRoute           = TaskRoute + "/submit"
	TaskCancelRoute           = TaskRoute + "/cancel"
)

// Routes - Task Groups
const (
	TaskGroupsRoute             = "/taskgroups"
	TaskGroupRoute              = TaskGroupsRoute + "/:" + ID
	TaskGroupBucketRoute        = TaskGroupRoute + "/bucket"
	TaskGroupBucketContentRoute = TaskGroupBucketRoute + "/*" + Wildcard
	TaskGroupSubmitRoute        = TaskGroupRoute + "/submit"
)

// Routes - Tickets
const (
	TicketsRoute = "/tickets"
	TicketRoute  = TicketsRoute + "/:" + ID
)

// Routes - Trackers
const (
	TrackersRoute                 = "/trackers"
	TrackerRoute                  = "/trackers" + "/:" + ID
	TrackerProjectsRoute          = TrackerRoute + "/projects"
	TrackerProjectRoute           = TrackerRoute + "/projects" + "/:" + ID2
	TrackerProjectIssueTypesRoute = TrackerProjectRoute + "/issuetypes"
)

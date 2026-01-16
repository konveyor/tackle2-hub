package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/internal/api/filter"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/assessment"
	"github.com/konveyor/tackle2-hub/internal/metrics"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/trigger"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Params
const (
	Source = api.Source
)

// Tag Sources
const (
	SourceAssessment = "assessment"
	SourceArchetype  = "archetype"
)

// ApplicationHandler handles application resource routes.
type ApplicationHandler struct {
	BucketOwner
}

// AddRoutes adds routes.
func (h ApplicationHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("applications"), Transaction)
	routeGroup.GET(api.ApplicationsRoute, h.List)
	routeGroup.GET(api.ApplicationsRoute+"/", h.List)
	routeGroup.POST(api.ApplicationsRoute, h.Create)
	routeGroup.GET(api.ApplicationRoute, h.Get)
	routeGroup.PUT(api.ApplicationRoute, h.Update)
	routeGroup.DELETE(api.ApplicationsRoute, h.DeleteList)
	routeGroup.DELETE(api.ApplicationRoute, h.Delete)
	// Tags
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications"), Transaction)
	routeGroup.GET(api.ApplicationTagsRoute, h.TagList)
	routeGroup.GET(api.ApplicationTagsRoute+"/", h.TagList)
	routeGroup.POST(api.ApplicationTagsRoute, h.TagAdd)
	routeGroup.DELETE(api.ApplicationTagRoute, h.TagDelete)
	routeGroup.PUT(api.ApplicationTagsRoute, h.TagReplace)
	// Facts
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.facts"), Transaction)
	routeGroup.GET(api.ApplicationFactsRoute, h.FactGet)
	routeGroup.GET(api.ApplicationFactsRoute+"/", h.FactGet)
	routeGroup.POST(api.ApplicationFactsRoute, h.FactCreate)
	routeGroup.GET(api.ApplicationFactRoute, h.FactGet)
	routeGroup.PUT(api.ApplicationFactRoute, h.FactPut)
	routeGroup.DELETE(api.ApplicationFactRoute, h.FactDelete)
	routeGroup.PUT(api.ApplicationFactsRoute, h.FactPut)
	// Bucket
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.bucket"))
	routeGroup.GET(api.AppBucketRoute, h.BucketGet)
	routeGroup.GET(api.AppBucketContentRoute, h.BucketGet)
	routeGroup.POST(api.AppBucketContentRoute, h.BucketPut)
	routeGroup.PUT(api.AppBucketContentRoute, h.BucketPut)
	routeGroup.DELETE(api.AppBucketContentRoute, h.BucketDelete)
	// Stakeholders
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.stakeholders"), Transaction)
	routeGroup.PUT(api.AppStakeholdersRoute, h.StakeholdersUpdate)
	// Assessments
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.assessments"), Transaction)
	routeGroup.GET(api.AppAssessmentsRoute, h.AssessmentList)
	routeGroup.POST(api.AppAssessmentsRoute, h.AssessmentCreate)
}

// Get godoc
// @summary Get an application by ID.
// @description Get an application by ID.
// @tags applications
// @produce json
// @success 200 {object} api.Application
// @router /applications/{id} [get]
// @param id path int true "Application ID"
func (h ApplicationHandler) Get(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	tagMap, err := h.tagMap(ctx, []uint{id})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	idMap, err := h.idMap(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tagResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	appResolver := assessment.NewApplicationResolver(tagResolver, memberResolver, questResolver)

	r := Application{}
	tags := tagMap[m.ID]
	r.With(m, tags, idMap.List())
	err = r.WithResolver(m, appResolver)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all applications.
// @description List all applications.
// @description filters:
// @description - name
// @description - platform.id
// @description - repository.url
// @description - repository.path
// @tags applications
// @produce json
// @success 200 {object} []api.Application
// @router /applications [get]
func (h ApplicationHandler) List(ctx *gin.Context) {
	questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tagResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	appResolver := assessment.NewApplicationResolver(tagResolver, memberResolver, questResolver)

	tagMap, err := h.tagMap(ctx, nil)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "name", Kind: qf.STRING},
			{Field: "platform.id", Kind: qf.LITERAL},
			{Field: "repository.url", Kind: qf.STRING},
			{Field: "repository.path", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("platform.id", "PlatformId")

	type M struct {
		*model.Application
		IdentityId           uint
		IdentityRole         string
		IdentityName         string
		ServiceId            uint
		ServiceName          string
		OwnerId              uint
		OwnerName            string
		ContributorId        uint
		ContributorName      string
		WaveId               uint
		WaveName             string
		PlatformId           uint
		PlatformName         string
		ReviewId             uint
		AssessmentId         uint
		AssessmentSections   []byte
		AssessmentThresholds []byte
		ManifestId           uint
		QuestionnaireId      uint
		AnalysisId           uint
		Effort               int
	}
	db := h.DB(ctx)
	db = db.Select(
		"a.*",
		"id.ID              IdentityId",
		"ai.Role            IdentityRole",
		"id.Name            IdentityName",
		"bs.ID              ServiceId",
		"bs.Name            ServiceName",
		"st.ID              OwnerId",
		"st.Name            OwnerName",
		"cn.ID              ContributorId",
		"cn.Name            ContributorName",
		"mw.ID              WaveId",
		"mw.Name            WaveName",
		"pf.ID              PlatformId",
		"pf.Name            PlatformName",
		"rv.ID              ReviewId",
		"at.ID              AssessmentId",
		"at.Sections        AssessmentSections",
		"at.Thresholds     AssessmentThresholds",
		"mf.ID              ManifestId",
		"at.QuestionnaireID QuestionnaireId",
		"an.ID              AnalysisId",
		"an.Effort          Effort",
	)
	db = db.Table("Application a")
	db = db.Joins("LEFT JOIN ApplicationIdentity ai ON ai.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Identity id ON id.ID = ai.IdentityID")
	db = db.Joins("LEFT JOIN BusinessService bs ON bs.ID = a.BusinessServiceID")
	db = db.Joins("LEFT JOIN Stakeholder st ON st.ID = a.OwnerID")
	db = db.Joins("LEFT JOIN ApplicationContributors ac ON ac.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Stakeholder cn ON cn.ID = ac.StakeholderID")
	db = db.Joins("LEFT JOIN MigrationWave mw ON mw.ID = a.MigrationWaveID")
	db = db.Joins("LEFT JOIN Platform pf ON pf.ID = a.PlatformID")
	db = db.Joins("LEFT JOIN Review rv ON rv.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Assessment at ON at.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Manifest mf ON mf.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Analysis an ON an.ApplicationID = a.ID")
	db = db.Where("a.ID IN (?)", h.appIds(ctx, filter))
	db = db.Order("a.ID")
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	builder := func(batch []any) (out any, err error) {
		app := &model.Application{}
		idMap := make(IdentityMap)
		contributors := make(map[uint]model.Stakeholder)
		assessments := make(map[uint]model.Assessment)
		manifests := make(map[uint]model.Manifest)
		analyses := make(map[uint]model.Analysis)
		for i := range batch {
			m := batch[i].(*M)
			app = m.Application
			if m.PlatformId > 0 {
				app.Platform = &model.Platform{}
				app.Platform.ID = m.PlatformId
				app.Platform.Name = m.PlatformName
			}
			if m.ServiceId > 0 {
				app.BusinessService = &model.BusinessService{}
				app.BusinessService.ID = m.ServiceId
				app.BusinessService.Name = m.ServiceName
			}
			if m.OwnerId > 0 {
				app.Owner = &model.Stakeholder{}
				app.Owner.ID = m.OwnerId
				app.Owner.Name = m.OwnerName
			}
			if m.WaveId > 0 {
				app.MigrationWave = &model.MigrationWave{}
				app.MigrationWave.ID = m.WaveId
				app.MigrationWave.Name = m.WaveName
			}
			if m.ReviewId > 0 {
				app.Review = &model.Review{}
				app.Review.ID = m.ReviewId
			}
			if m.IdentityId > 0 {
				ref := IdentityRef{}
				ref.ID = m.IdentityId
				ref.Name = m.IdentityName
				ref.Role = m.IdentityRole
				idMap[ref] = 0
			}
			if m.ContributorId > 0 {
				ref := model.Stakeholder{}
				ref.ID = m.ContributorId
				ref.Name = m.ContributorName
				contributors[m.ContributorId] = ref
			}
			if m.AssessmentId > 0 {
				ref := model.Assessment{}
				ref.ID = m.AssessmentId
				ref.QuestionnaireID = m.QuestionnaireId
				_ = json.Unmarshal(m.AssessmentSections, &ref.Sections)
				_ = json.Unmarshal(m.AssessmentThresholds, &ref.Thresholds)
				assessments[m.AssessmentId] = ref
			}
			if m.ManifestId > 0 {
				ref := model.Manifest{}
				ref.ID = m.ManifestId
				manifests[m.ManifestId] = ref
			}
			if m.AnalysisId > 0 {
				ref := model.Analysis{}
				ref.ID = m.AnalysisId
				ref.Effort = m.Effort
				analyses[m.AnalysisId] = ref
			}
		}
		for _, m := range contributors {
			app.Contributors = append(app.Contributors, m)
		}
		for _, m := range assessments {
			app.Assessments = append(app.Assessments, m)
		}
		for _, m := range manifests {
			app.Manifest = append(app.Manifest, m)
		}
		for _, m := range analyses {
			app.Analyses = append(app.Analyses, m)
		}
		tagMap.Set(app)
		r := Application{}
		tags := tagMap[app.ID]
		r.With(app, tags, idMap.List())
		err = r.WithResolver(app, appResolver)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		out = r
		return
	}
	iter := NewIterator(&M{}, &cursor, builder)
	h.Respond(ctx, http.StatusOK, iter)
}

// Create godoc
// @summary Create an application.
// @description Create an application.
// @tags applications
// @accept json
// @produce json
// @success 201 {object} api.Application
// @router /applications [post]
// @param application body api.Application true "Application data"
func (h ApplicationHandler) Create(ctx *gin.Context) {
	r := &Application{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Omit(clause.Associations).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	idMap := IdentityMap{}
	idMap.With(r)
	db := h.DB(ctx).Model(m)
	err = db.Association("Contributors").Replace(m.Contributors)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.replaceIdentities(h.DB(ctx), m.ID, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	appTags, err := h.replaceTags(h.DB(ctx), m.ID, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tagResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	appResolver := assessment.NewApplicationResolver(tagResolver, memberResolver, questResolver)

	r.With(m, appTags, idMap.List())
	err = r.WithResolver(m, appResolver)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	rtx := RichContext(ctx)
	tr := trigger.Application{
		Trigger: trigger.Trigger{
			User:        rtx.User,
			TaskManager: rtx.TaskManager,
			Client:      rtx.Client,
			DB:          h.DB(ctx),
		},
	}
	err = tr.Created(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an application.
// @description Delete an application.
// @tags applications
// @success 204
// @router /applications/{id} [delete]
// @param id path int true "Application id"
func (h ApplicationHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Application{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// DeleteList godoc
// @summary Delete a applications.
// @description Delete applications.
// @tags applications
// @success 204
// @router /applications [delete]
// @param application body []uint true "List of id"
func (h ApplicationHandler) DeleteList(ctx *gin.Context) {
	ids := []uint{}
	err := h.Bind(ctx, &ids)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(
		&model.Application{},
		"id IN ?",
		ids).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update an application.
// @description Update an application.
// @tags applications
// @accept json
// @success 204
// @router /applications/{id} [put]
// @param id path int true "Application id"
// @param application body api.Application true "Application data"
func (h ApplicationHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Application{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := &model.Application{}
	db := h.DB(ctx)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	//
	// Update the application.
	m = r.Model()
	m.Tags = nil
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db = h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations, "BucketID")
	result = db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Contributors").Replace(m.Contributors)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.replaceIdentities(h.DB(ctx), m.ID, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	_, err = h.replaceTags(h.DB(ctx), m.ID, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	rtx := RichContext(ctx)
	tr := trigger.Application{
		Trigger: trigger.Trigger{
			User:        rtx.User,
			TaskManager: rtx.TaskManager,
			Client:      rtx.Client,
			DB:          h.DB(ctx),
		},
	}
	err = tr.Updated(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// BucketGet godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @description Returns index.html for directories when Accept=text/html else a tarball.
// @description ?filter=glob supports directory content filtering.
// @tags applications
// @produce octet-stream
// @success 200
// @router /applications/{id}/bucket/{wildcard} [get]
// @param id path int true "Application ID"
// @param wildcard path string true "Content path"
// @param filter query string false "Filter"
func (h ApplicationHandler) BucketGet(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		h.Status(ctx, http.StatusNotFound)
		return
	}

	h.bucketGet(ctx, *m.BucketID)
}

// BucketPut godoc
// @summary Upload bucket content by ID and path.
// @description Upload bucket content by ID and path (handles both [post] and [put] requests).
// @tags applications
// @produce json
// @success 204
// @router /applications/{id}/bucket/{wildcard} [post]
// @param id path int true "Application ID"
// @param wildcard path string true "Content path"
func (h ApplicationHandler) BucketPut(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		h.Status(ctx, http.StatusNotFound)
		return
	}

	h.bucketPut(ctx, *m.BucketID)
}

// BucketDelete godoc
// @summary Delete bucket content by ID and path.
// @description Delete bucket content by ID and path.
// @tags applications
// @produce json
// @success 204
// @router /applications/{id}/bucket/{wildcard} [delete]
// @param id path int true "Application ID"
// @param wildcard path string true "Content path"
func (h ApplicationHandler) BucketDelete(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		h.Status(ctx, http.StatusNotFound)
		return
	}

	h.bucketDelete(ctx, *m.BucketID)
}

// TagList godoc
// @summary List tag references.
// @description List tag references.
// @tags applications
// @produce json
// @success 200 {object} []api.Ref
// @router /applications/{id}/tags [get]
// @param id path int true "Application ID"
func (h ApplicationHandler) TagList(ctx *gin.Context) {
	id := h.pk(ctx)
	app := &model.Application{}
	result := h.DB(ctx).Preload("Tags").First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	source, found := ctx.GetQuery(Source)
	if found {
		condition := h.DB(ctx).Where("source = ?", source)
		db = db.Where(condition)
	}

	list := []model.ApplicationTag{}
	result = db.Find(&list, "ApplicationID = ?", id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []TagRef{}
	for i := range list {
		r := TagRef{
			ID:      list[i].Tag.ID,
			Name:    list[i].Tag.Name,
			Source:  list[i].Source,
			Virtual: false,
		}
		resources = append(resources, r)
	}

	includeAssessment := !found || source == SourceAssessment
	includeArchetype := !found || source == SourceArchetype
	if includeAssessment || includeArchetype {
		questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		tagResolver, err := assessment.NewTagResolver(h.DB(ctx))
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		appResolver := assessment.NewApplicationResolver(tagResolver, memberResolver, questResolver)
		if includeArchetype {
			archetypeTags, err := appResolver.ArchetypeTags(app)
			if err != nil {
				_ = ctx.Error(err)
				return
			}
			for i := range archetypeTags {
				r := TagRef{
					ID:      archetypeTags[i].ID,
					Name:    archetypeTags[i].Name,
					Source:  SourceArchetype,
					Virtual: true,
				}
				resources = append(resources, r)
			}
		}
		if includeAssessment {
			assessmentTags := appResolver.AssessmentTags(app)
			for i := range assessmentTags {
				r := TagRef{
					ID:      assessmentTags[i].ID,
					Name:    assessmentTags[i].Name,
					Source:  SourceAssessment,
					Virtual: true,
				}
				resources = append(resources, r)
			}
		}
	}
	h.Respond(ctx, http.StatusOK, resources)
}

// TagAdd godoc
// @summary Add tag association.
// @description Ensure tag is associated with the application.
// @tags applications
// @accept json
// @produce json
// @success 201 {object} api.Ref
// @router /applications/{id}/tags [post]
// @param tag body Ref true "Tag data"
// @param id path int true "Application ID"
func (h ApplicationHandler) TagAdd(ctx *gin.Context) {
	id := h.pk(ctx)
	ref := &TagRef{}
	err := h.Bind(ctx, ref)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if ref.Virtual {
		err = &BadRequestError{Reason: "cannot add virtual tags"}
		_ = ctx.Error(err)
		return
	}
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	tag := &model.ApplicationTag{
		ApplicationID: id,
		TagID:         ref.ID,
		Source:        ref.Source,
	}
	err = h.DB(ctx).Create(tag).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, ref)
}

// TagReplace godoc
// @summary Replace tag associations.
// @description Replace tag associations.
// @tags applications
// @accept json
// @success 204
// @router /applications/{id}/tags [patch]
// @param id path int true "Application ID"
// @param source query string false "Source"
// @param tags body []TagRef true "Tag references"
func (h ApplicationHandler) TagReplace(ctx *gin.Context) {
	id := h.pk(ctx)
	refs := []TagRef{}
	err := h.Bind(ctx, &refs)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// remove all the existing tag associations for that source and app id.
	// if source is not provided, all tag associations will be removed.
	db := h.DB(ctx).Where("ApplicationID = ?", id)
	source, found := ctx.GetQuery(Source)
	if found {
		condition := h.DB(ctx).Where("source = ?", source)
		db = db.Where(condition)
	}
	err = db.Delete(&model.ApplicationTag{}).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// create new associations
	if len(refs) > 0 {
		appTags := []model.ApplicationTag{}
		for _, ref := range refs {
			if !ref.Virtual {
				appTags = append(appTags, model.ApplicationTag{
					ApplicationID: id,
					TagID:         ref.ID,
					Source:        source,
				})
			}
		}
		err = db.Create(&appTags).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}

	h.Status(ctx, http.StatusNoContent)
}

// TagDelete godoc
// @summary Delete tag association.
// @description Ensure tag is not associated with the application.
// @tags applications
// @success 204
// @router /applications/{id}/tags/{sid} [delete]
// @param id path int true "Application ID"
// @param sid path string true "Tag ID"
func (h ApplicationHandler) TagDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	id2 := ctx.Param(ID2)
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	db := h.DB(ctx).Where("ApplicationID = ?", id).Where("TagID = ?", id2)
	source, found := ctx.GetQuery(Source)
	if found {
		condition := h.DB(ctx).Where("source = ?", source)
		db = db.Where(condition)
	}
	err := db.Delete(&model.ApplicationTag{}).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// FactList godoc
// @summary List facts.
// @description List facts by source.
// @description see api.FactKey for details on key parameter format.
// @tags applications
// @produce json
// @success 200 {object} api.Map
// @router /applications/{id}/facts/{source}: [get]
// @param id path int true "Application ID"
// @param source path string true "Source key"
func (h ApplicationHandler) FactList(ctx *gin.Context, key FactKey) {
	id := h.pk(ctx)
	list := []model.Fact{}
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	db = db.Where("Source", key.Source())
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	facts := Map{}
	for i := range list {
		fact := &list[i]
		facts[fact.Key] = fact.Value
	}
	h.Respond(ctx, http.StatusOK, facts)
}

// FactGet godoc
// @summary Get fact by name.
// @description Get fact by name.
// @description see api.FactKey for details on key parameter format.
// @tags applications
// @produce json
// @success 200 {object} object
// @router /applications/{id}/facts/{key} [get]
// @param id path int true "Application ID"
// @param key path string true "Fact key"
func (h ApplicationHandler) FactGet(ctx *gin.Context) {
	id := h.pk(ctx)
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	key := FactKey(ctx.Param(Key))
	if key.Name() == "" {
		h.FactList(ctx, key)
		return
	}

	list := []model.Fact{}
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	db = db.Where("Source", key.Source())
	db = db.Where("Key", key.Name())
	result = db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if len(list) < 1 {
		h.Status(ctx, http.StatusNotFound)
		return
	}

	h.Respond(ctx, http.StatusOK, list[0].Value)
}

// FactCreate godoc
// @summary Create a fact.
// @description Create a fact.
// @tags applications
// @accept json
// @produce json
// @success 201
// @router /applications/{id}/facts [post]
// @param id path int true "Application ID"
// @param fact body api.Fact true "Fact data"
func (h ApplicationHandler) FactCreate(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Fact{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	m := r.Model()
	m.ApplicationID = id
	result = h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Respond(ctx, http.StatusCreated, r)
}

// FactPut godoc
// @summary Update (or create) a fact.
// @description Update (or create) a fact.
// @description see api.FactKey for details on key parameter format.
// @tags applications
// @accept json
// @produce json
// @success 204
// @router /applications/{id}/facts/{key} [put]
// @param id path int true "Application ID"
// @param key path string true "Fact key"
// @param fact body object true "Fact value"
func (h ApplicationHandler) FactPut(ctx *gin.Context) {
	id := h.pk(ctx)
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	key := FactKey(ctx.Param(Key))
	if key.Name() == "" {
		h.FactReplace(ctx, key)
		return
	}
	f := Fact{}
	err := h.Bind(ctx, &f.Value)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := &model.Fact{
		Key:           key.Name(),
		Source:        key.Source(),
		ApplicationID: id,
		Value:         f.Value,
	}
	db := h.DB(ctx)
	db = db.Clauses(clause.OnConflict{UpdateAll: true})
	result = db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// FactDelete godoc
// @summary Delete a fact.
// @description Delete a fact.
// @description see api.FactKey for details on key parameter format.
// @tags applications
// @success 204
// @router /applications/{id}/facts/{key} [delete]
// @param id path int true "Application ID"
// @param key path string true "Fact key"
func (h ApplicationHandler) FactDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	fact := &model.Fact{}
	key := FactKey(ctx.Param(Key))
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	db = db.Where("Source", key.Source())
	db = db.Where("Key", key.Name())
	result = db.Delete(fact)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// FactReplace godoc
// @summary Replace all facts from a source.
// @description Replace all facts from a source.
// @description see api.FactKey for details on key parameter format.
// @tags applications
// @success 204
// @router /applications/{id}/facts/{source}: [put]
// @param id path int true "Application ID"
// @param source path string true "Fact key"
// @param factmap body api.Map true "Fact map"
func (h ApplicationHandler) FactReplace(ctx *gin.Context, key FactKey) {
	id := h.pk(ctx)
	facts := Map{}
	err := h.Bind(ctx, &facts)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// remove all the existing Facts for that source and app id.
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	db = db.Where("Source", key.Source())
	err = db.Delete(&model.Fact{}).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// create new Facts
	if len(facts) > 0 {
		newFacts := []model.Fact{}
		for k, v := range facts {
			value, _ := json.Marshal(v)
			newFacts = append(newFacts, model.Fact{
				ApplicationID: id,
				Key:           FactKey(k).Name(),
				Value:         value,
				Source:        key.Source(),
			})
		}
		err = db.Create(&newFacts).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}

	h.Status(ctx, http.StatusNoContent)
}

// StakeholdersUpdate godoc
// @summary Update the owner and contributors of an Application.
// @description Update the owner and contributors of an Application.
// @tags applications
// @success 204
// @router /applications/{id}/stakeholders [patch]
// @param id path int true "Application ID"
// @param application body api.Stakeholders true "Application stakeholders"
func (h ApplicationHandler) StakeholdersUpdate(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx))
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := &Stakeholders{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	db = h.DB(ctx).Model(m).Omit(clause.Associations, "BucketID")
	result = db.Updates(map[string]any{"OwnerID": r.ownerID()})
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	err = h.DB(ctx).Model(m).Association("Contributors").Replace(r.contributors())
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// AssessmentList godoc
// @summary List the assessments of an Application and any it inherits from its archetypes.
// @description List the assessments of an Application and any it inherits from its archetypes.
// @tags applications
// @success 200 {object} []api.Assessment
// @router /applications/{id}/assessments [get]
// @param id path int true "Application ID"
func (h ApplicationHandler) AssessmentList(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Assessments.Stakeholders",
		"Assessments.StakeholderGroups",
		"Assessments.Questionnaire")
	db = db.Omit("Analyses")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tagResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	appResolver := assessment.NewApplicationResolver(tagResolver, memberResolver, questResolver)
	archetypes, err := appResolver.Archetypes(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	assessments := m.Assessments
	for _, arch := range archetypes {
		for _, a := range arch.Assessments {
			assessments = append(assessments, *a.Assessment)
		}
	}
	resources := []Assessment{}
	for i := range assessments {
		r := Assessment{}
		r.With(&assessments[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AssessmentCreate godoc
// @summary Create an application assessment.
// @description Create an application assessment.
// @tags applications
// @accept json
// @produce json
// @success 201 {object} api.Assessment
// @router /applications/{id}/assessments [post]
// @param id path int true "Application ID"
// @param assessment body api.Assessment true "Assessment data"
func (h ApplicationHandler) AssessmentCreate(ctx *gin.Context) {
	application := &model.Application{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	db = db.Omit("Analyses")
	result := db.First(application, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := &Assessment{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Application = &resource.Ref{ID: id}
	r.Archetype = nil
	q := &model.Questionnaire{}
	db = h.preLoad(h.DB(ctx))
	result = db.First(q, r.Questionnaire.ID)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	m := r.Model()
	m.Thresholds = q.Thresholds
	m.RiskMessages = q.RiskMessages
	m.CreateUser = h.CurrentUser(ctx)
	// if sections aren't empty that indicates that this assessment is being
	// created "as-is" and should not have its sections populated or autofilled.
	newAssessment := false
	if len(m.Sections) == 0 {
		m.Sections = q.Sections
		resolver, rErr := assessment.NewTagResolver(h.DB(ctx))
		if rErr != nil {
			_ = ctx.Error(rErr)
			return
		}
		assessment.PrepareForApplication(resolver, application, m)
		newAssessment = true
	}
	result = h.DB(ctx).Omit(clause.Associations).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.DB(ctx).Model(m).Association("Stakeholders").Replace("Stakeholders", m.Stakeholders)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("StakeholderGroups").Replace("StakeholderGroups", m.StakeholderGroups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if newAssessment {
		metrics.AssessmentsInitiated.Inc()
	}

	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// tagMap returns a map of AppTag indexed by application id.
// This is a performance and memory optimization.
func (h *ApplicationHandler) tagMap(
	ctx *gin.Context,
	appIds []uint) (mp TagMap, err error) {
	tagCache := make(map[uint]*model.Tag)
	var tags []*model.Tag
	db := h.DB(ctx)
	err = db.Find(&tags).Error
	if err != nil {
		return
	}
	for _, tag := range tags {
		tagCache[tag.ID] = tag
	}
	mp = make(TagMap)
	var appTags []AppTag
	db = h.DB(ctx)
	db = db.Omit(clause.Associations)
	db = db.Table("ApplicationTags")
	if len(appIds) > 0 {
		db = db.Where("ApplicationID", appIds)
	}
	err = db.Find(&appTags).Error
	if err != nil {
		return
	}
	for _, m := range appTags {
		m.Tag = tagCache[m.TagID]
		mp[m.ApplicationID] = append(
			mp[m.ApplicationID],
			m)
	}
	return
}

// idMap returns a loaded IdentityMap.
func (h *ApplicationHandler) idMap(ctx *gin.Context, appId uint) (mp IdentityMap, err error) {
	mp = make(IdentityMap)
	var list []IdentityRef
	db := h.DB(ctx)
	db = db.Table("Identity id")
	db = db.Select(
		"id.ID",
		"id.Name",
		"ai.Role")
	db = db.Joins("LEFT JOIN ApplicationIdentity ai ON ai.IdentityID = id.ID")
	db = db.Where("ai.ApplicationID", appId)
	err = db.Find(&list).Error
	if err != nil {
		return
	}
	for _, m := range list {
		ref := IdentityRef{}
		ref.ID = m.ID
		ref.Name = m.Name
		ref.Role = m.Role
		mp[ref] = 0
	}
	return
}

// replaceTags replaces tag associations.
func (h *ApplicationHandler) replaceTags(db *gorm.DB, id uint, r *Application) (appTags []AppTag, err error) {
	appTags = []AppTag{}
	var list []model.ApplicationTag
	m := &model.ApplicationTag{}
	err = db.Delete(m, "ApplicationID", id).Error
	if err != nil {
		return
	}
	for _, ref := range r.Tags {
		if !ref.Virtual {
			appTag := AppTag{}
			appTag.WithRef(&ref)
			appTags = append(appTags, appTag)
			m := model.ApplicationTag{}
			m.ApplicationID = id
			m.TagID = ref.ID
			m.Source = ref.Source
			list = append(list, m)
		}
	}
	err = db.Create(&list).Error
	if err != nil {
		return
	}
	return
}

// replaceTags replaces identity associations.
func (h *ApplicationHandler) replaceIdentities(db *gorm.DB, id uint, r *Application) (err error) {
	m := &model.ApplicationIdentity{}
	err = db.Delete(m, "ApplicationID", id).Error
	if err != nil {
		return
	}
	var list []model.ApplicationIdentity
	for _, ref := range r.Identities {
		list = append(
			list,
			model.ApplicationIdentity{
				ApplicationID: id,
				IdentityID:    ref.ID,
				Role:          ref.Role,
			})
	}
	err = db.Create(&list).Error
	if err != nil {
		return
	}
	return
}

// appIds returns application ids based on filter.
func (h *ApplicationHandler) appIds(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Application{})
	q = q.Select("ID")
	q = f.Where(q)
	filter := f
	repository := filter.Resource("repository")
	if repository.Empty() {
		return
	}
	iq := h.DB(ctx)
	iq = iq.Select("m.ID")
	iq = iq.Table("Application m")
	iq = iq.Joins("LEFT JOIN json_tree(repository) jt")
	iq = iq.Group("m.ID")
	for _, fn := range []string{"url", "path"} {
		if f, found := repository.Field(fn); found {
			fv := make([]string, 0)
			f.Value.Into(&fv)
			iq = iq.Having(
				"SUM(jt.key = ? AND jt.value IN (?)) > 0",
				fn,
				fv)
		}
	}
	iq = f.Where(iq)
	q = q.Where("ID IN (?)", iq)
	return
}

// Application REST resource.
type Application = resource.Application

// Repository REST nested resource.
type Repository = resource.Repository

// Fact REST nested resource.
type Fact = resource.Fact

// IdentityRef REST resource.
type IdentityRef = resource.IdentityRef

// FactKey is a fact source and fact name separated by a colon.
//
//	Example: 'analysis:languages'
//
// A FactKey can be used to identify an anonymous fact.
//
//	Example: 'languages' or ':languages'
//
// A FactKey can also be used to identify just a source. This use must include the trailing
// colon to distinguish it from an anonymous fact. This is used when listing or replacing
// all facts that belong to a source.
//
//	Example: 'analysis:"
type FactKey = resource.FactKey

// Stakeholders REST subresource.
type Stakeholders struct {
	Owner        *Ref  `json:"owner"`
	Contributors []Ref `json:"contributors"`
}

func (r *Stakeholders) ownerID() (ownerID *uint) {
	if r.Owner != nil {
		ownerID = &r.Owner.ID
	}
	return
}

func (r *Stakeholders) contributors() (contributors []model.Stakeholder) {
	for _, ref := range r.Contributors {
		contributors = append(
			contributors,
			model.Stakeholder{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	return
}

// TagRef tag reference.
type TagRef = resource.TagRef

// TagMap contains a map of tags by id.
type TagMap = resource.TagMap

// AppTag represents application tag mapping.
type AppTag = resource.AppTag

// IdentityMap represents application/identity associations.
type IdentityMap = resource.IdentityMap

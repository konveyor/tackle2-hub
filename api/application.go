package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/metrics"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/trigger"
	"gorm.io/gorm/clause"
)

// Routes
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

// Params
const (
	Source = "source"
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
	routeGroup.GET(ApplicationsRoot, h.List)
	routeGroup.GET(ApplicationsRoot+"/", h.List)
	routeGroup.POST(ApplicationsRoot, h.Create)
	routeGroup.GET(ApplicationRoot, h.Get)
	routeGroup.PUT(ApplicationRoot, h.Update)
	routeGroup.DELETE(ApplicationsRoot, h.DeleteList)
	routeGroup.DELETE(ApplicationRoot, h.Delete)
	// Tags
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications"), Transaction)
	routeGroup.GET(ApplicationTagsRoot, h.TagList)
	routeGroup.GET(ApplicationTagsRoot+"/", h.TagList)
	routeGroup.POST(ApplicationTagsRoot, h.TagAdd)
	routeGroup.DELETE(ApplicationTagRoot, h.TagDelete)
	routeGroup.PUT(ApplicationTagsRoot, h.TagReplace)
	// Facts
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.facts"), Transaction)
	routeGroup.GET(ApplicationFactsRoot, h.FactGet)
	routeGroup.GET(ApplicationFactsRoot+"/", h.FactGet)
	routeGroup.POST(ApplicationFactsRoot, h.FactCreate)
	routeGroup.GET(ApplicationFactRoot, h.FactGet)
	routeGroup.PUT(ApplicationFactRoot, h.FactPut)
	routeGroup.DELETE(ApplicationFactRoot, h.FactDelete)
	routeGroup.PUT(ApplicationFactsRoot, h.FactPut)
	// Bucket
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.bucket"))
	routeGroup.GET(AppBucketRoot, h.BucketGet)
	routeGroup.GET(AppBucketContentRoot, h.BucketGet)
	routeGroup.POST(AppBucketContentRoot, h.BucketPut)
	routeGroup.PUT(AppBucketContentRoot, h.BucketPut)
	routeGroup.DELETE(AppBucketContentRoot, h.BucketDelete)
	// Stakeholders
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.stakeholders"), Transaction)
	routeGroup.PUT(AppStakeholdersRoot, h.StakeholdersUpdate)
	// Assessments
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.assessments"), Transaction)
	routeGroup.GET(AppAssessmentsRoot, h.AssessmentList)
	routeGroup.POST(AppAssessmentsRoot, h.AssessmentCreate)
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
	r.With(m, tagMap[m.ID])
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

	type M struct {
		*model.Application
		IdentityId   uint
		IdentityName string
		ServiceId    uint
		ServiceName  string
		OwnerId      uint
		OwnerName    string
		ContId       uint
		ContName     string
		WaveId       uint
		WaveName     string
		AnId         uint
		AnEffort     int
	}
	db := h.DB(ctx)
	db = db.Select(
		"a.*",
		"id.ID     IdentityId",
		"id.Name   IdentityName",
		"bs.ID     ServiceId",
		"bs.Name   ServiceName",
		"st.ID     OwnerId",
		"st.Name   OwnerName",
		"cn.ID     ContId",
		"cn.Name   ContName",
		"mw.ID     WaveId",
		"mw.Name   WaveName",
		"an.ID     AnId",
		"an.Effort AnEffort",
	)
	db = db.Table("Application a")
	db = db.Joins("LEFT JOIN Bucket b ON b.ID = a.BucketID")
	db = db.Joins("LEFT JOIN ApplicationIdentity ai ON ai.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Identity id ON id.ID = ai.IdentityID")
	db = db.Joins("LEFT JOIN BusinessService bs ON bs.ID = a.BusinessServiceID")
	db = db.Joins("LEFT JOIN Stakeholder st ON st.ID = a.OwnerID")
	db = db.Joins("LEFT JOIN ApplicationContributors ac ON ac.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Stakeholder cn ON cn.ID = ac.StakeholderID")
	db = db.Joins("LEFT JOIN MigrationWave mw ON mw.ID = a.MigrationWaveID")
	db = db.Joins("LEFT JOIN Analysis an ON an.ApplicationID = a.ID")
	db = db.Order("a.ID")
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	builder := func(batch []any) (out any, err error) {
		app := &model.Application{}
		identities := make(map[uint]model.Identity)
		contributors := make(map[uint]model.Stakeholder)
		analyses := make(map[uint]model.Analysis)
		for i := range batch {
			m := batch[i].(*M)
			app = m.Application
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
			if m.IdentityId > 0 {
				ref := model.Identity{}
				ref.ID = m.IdentityId
				ref.Name = m.IdentityName
				identities[m.IdentityId] = ref
			}
			if m.ContId > 0 {
				ref := model.Stakeholder{}
				ref.ID = m.ContId
				ref.Name = m.ContName
				contributors[m.ContId] = ref
			}
			if m.AnId > 0 {
				ref := model.Analysis{}
				ref.ApplicationID = app.ID
				ref.Effort = m.AnEffort
				analyses[m.AnId] = ref
			}
		}
		for _, m := range identities {
			app.Identities = append(app.Identities, m)
		}
		for _, m := range contributors {
			app.Contributors = append(app.Contributors, m)
		}
		for _, m := range analyses {
			app.Analyses = append(app.Analyses, m)
		}
		r := Application{}
		r.With(app, tagMap[app.ID])
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
	db := h.DB(ctx).Model(m)
	err = db.Association("Identities").Replace(m.Identities)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Contributors").Replace(m.Contributors)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	appTags := []AppTag{}
	tags := []model.ApplicationTag{}
	if len(r.Tags) > 0 {
		for _, t := range r.Tags {
			if !t.Virtual {
				appTag := AppTag{}
				appTag.withRef(&t)
				appTags = append(appTags, appTag)
				tag := model.ApplicationTag{}
				tag.ApplicationID = m.ID
				tag.TagID = t.ID
				tags = append(tags, tag)
			}
		}
		result = h.DB(ctx).Create(&tags)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
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
	r.With(m, appTags)
	err = r.WithResolver(m, appResolver)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	rtx := RichContext(ctx)
	tr := trigger.Application{
		Trigger: trigger.Trigger{
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
	//
	// Delete unwanted facts.
	m := &model.Application{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
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
	err = db.Association("Identities").Replace(m.Identities)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Contributors").Replace(m.Contributors)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// delete existing tag associations and create new ones
	err = h.DB(ctx).Delete(&model.ApplicationTag{}, "ApplicationID = ?", id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if len(r.Tags) > 0 {
		tags := []model.ApplicationTag{}
		for _, t := range r.Tags {
			if !t.Virtual {
				tags = append(tags, model.ApplicationTag{TagID: t.ID, ApplicationID: m.ID, Source: t.Source})
			}
		}
		result = h.DB(ctx).Create(&tags)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
	}

	rtx := RichContext(ctx)
	tr := trigger.Application{
		Trigger: trigger.Trigger{
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
		r := TagRef{}
		r.With(list[i].Tag.ID, list[i].Tag.Name, list[i].Source, false)
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
				r := TagRef{}
				r.With(archetypeTags[i].ID, archetypeTags[i].Name, SourceArchetype, true)
				resources = append(resources, r)
			}
		}
		if includeAssessment {
			assessmentTags := appResolver.AssessmentTags(app)
			for i := range assessmentTags {
				r := TagRef{}
				r.With(assessmentTags[i].ID, assessmentTags[i].Name, SourceAssessment, true)
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
		err = &BadRequestError{"cannot add virtual tags"}
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
	r.Application = &Ref{ID: id}
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

// Application REST resource.
type Application struct {
	Resource        `yaml:",inline"`
	Name            string      `json:"name" binding:"required"`
	Description     string      `json:"description"`
	Bucket          *Ref        `json:"bucket"`
	Repository      *Repository `json:"repository"`
	Binary          string      `json:"binary"`
	Review          *Ref        `json:"review"`
	Comments        string      `json:"comments"`
	Identities      []Ref       `json:"identities"`
	Tags            []TagRef    `json:"tags"`
	BusinessService *Ref        `json:"businessService" yaml:"businessService"`
	Owner           *Ref        `json:"owner"`
	Contributors    []Ref       `json:"contributors"`
	MigrationWave   *Ref        `json:"migrationWave" yaml:"migrationWave"`
	Platform        *Ref        `json:"platform"`
	Archetypes      []Ref       `json:"archetypes"`
	Assessments     []Ref       `json:"assessments"`
	Assessed        bool        `json:"assessed"`
	Risk            string      `json:"risk"`
	Confidence      int         `json:"confidence"`
	Effort          int         `json:"effort"`
}

// With updates the resource using the model.
func (r *Application) With(m *model.Application, tags []AppTag) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.Comments = m.Comments
	r.Binary = m.Binary
	if m.Repository != (model.Repository{}) {
		repo := Repository(m.Repository)
		r.Repository = &repo
	}
	if m.Review != nil {
		ref := &Ref{}
		ref.With(m.Review.ID, "")
		r.Review = ref
	}
	r.BusinessService = r.refPtr(m.BusinessServiceID, m.BusinessService)
	r.Identities = []Ref{}
	for _, id := range m.Identities {
		ref := Ref{}
		ref.With(id.ID, id.Name)
		r.Identities = append(
			r.Identities,
			ref)
	}
	r.Tags = []TagRef{}
	for i := range tags {
		ref := TagRef{}
		ref.With(tags[i].TagID, tags[i].Tag.Name, tags[i].Source, false)
		r.Tags = append(r.Tags, ref)
	}
	r.Owner = r.refPtr(m.OwnerID, m.Owner)
	r.Contributors = []Ref{}
	for _, c := range m.Contributors {
		ref := Ref{}
		ref.With(c.ID, c.Name)
		r.Contributors = append(
			r.Contributors,
			ref)
	}
	r.MigrationWave = r.refPtr(m.MigrationWaveID, m.MigrationWave)
	r.Platform = r.refPtr(m.PlatformID, m.Platform)
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		ref := Ref{}
		ref.With(a.ID, "")
		r.Assessments = append(r.Assessments, ref)
	}
	if len(m.Analyses) > 0 {
		sort.Slice(m.Analyses, func(i, j int) bool {
			return m.Analyses[i].ID < m.Analyses[j].ID
		})
		r.Effort = m.Analyses[len(m.Analyses)-1].Effort
	}
	r.Risk = assessment.RiskUnassessed
}

// WithVirtualTags updates the resource with tags derived from assessments.
func (r *Application) WithVirtualTags(tags []model.Tag, source string) {
	for _, t := range tags {
		ref := TagRef{}
		ref.With(t.ID, t.Name, source, true)
		r.Tags = append(r.Tags, ref)
	}
}

// WithResolver uses an ApplicationResolver to update the resource with
// values derived from the application's assessments and archetypes.
func (r *Application) WithResolver(m *model.Application, resolver *assessment.ApplicationResolver) (err error) {
	archetypes, err := resolver.Archetypes(m)
	if err != nil {
		return
	}
	for _, a := range archetypes {
		ref := Ref{}
		ref.With(a.ID, a.Name)
		r.Archetypes = append(r.Archetypes, ref)
	}
	archetypeTags, err := resolver.ArchetypeTags(m)
	if err != nil {
		return
	}
	r.WithVirtualTags(archetypeTags, SourceArchetype)
	r.WithVirtualTags(resolver.AssessmentTags(m), SourceAssessment)
	r.Assessed, err = resolver.Assessed(m)
	if err != nil {
		return
	}
	if r.Assessed {
		r.Confidence, err = resolver.Confidence(m)
		if err != nil {
			return
		}
		r.Risk, err = resolver.Risk(m)
		if err != nil {
			return
		}
	}
	return
}

// Model builds a model.
func (r *Application) Model() (m *model.Application) {
	m = &model.Application{
		Name:        r.Name,
		Description: r.Description,
		Comments:    r.Comments,
		Binary:      r.Binary,
	}
	m.ID = r.ID
	if r.Repository != nil {
		m.Repository = model.Repository(*r.Repository)
	}
	if r.BusinessService != nil {
		m.BusinessServiceID = &r.BusinessService.ID
	}
	for _, ref := range r.Identities {
		m.Identities = append(
			m.Identities,
			model.Identity{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.Tags {
		m.Tags = append(
			m.Tags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	if r.Owner != nil {
		m.OwnerID = &r.Owner.ID
	}
	for _, ref := range r.Contributors {
		m.Contributors = append(
			m.Contributors,
			model.Stakeholder{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	if r.MigrationWave != nil {
		m.MigrationWaveID = &r.MigrationWave.ID
	}
	if r.Platform != nil {
		m.PlatformID = &r.Platform.ID
	}

	return
}

// Repository REST nested resource.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

// Fact REST nested resource.
type Fact struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

func (r *Fact) With(m *model.Fact) {
	r.Key = m.Key
	r.Source = m.Source
	r.Value = m.Value
}

func (r *Fact) Model() (m *model.Fact) {
	m = &model.Fact{}
	m.Key = r.Key
	m.Source = r.Source
	m.Value = r.Value
	return
}

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
type FactKey string

// Qualify qualifies the name with the source.
func (r *FactKey) Qualify(source string) {
	*r = FactKey(
		strings.Join(
			[]string{source, r.Name()},
			":"))
}

// Source returns the source portion of a fact key.
func (r FactKey) Source() (source string) {
	s, _, found := strings.Cut(string(r), ":")
	if found {
		source = s
	}
	return
}

// Name returns the name portion of a fact key.
func (r FactKey) Name() (name string) {
	_, n, found := strings.Cut(string(r), ":")
	if found {
		name = n
	} else {
		name = string(r)
	}
	return
}

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

type TagMap map[uint][]AppTag

// AppTag is a lightweight representation of ApplicationTag model.
type AppTag struct {
	ApplicationID uint
	TagID         uint
	Source        string
	Tag           *model.Tag
}

func (r *AppTag) with(m *model.ApplicationTag) {
	r.ApplicationID = m.ApplicationID
	r.Source = m.Source
	r.Tag = &m.Tag
}

func (r *AppTag) withRef(m *TagRef) {
	r.Source = m.Source
	r.Tag = &model.Tag{}
	r.Tag.ID = m.ID
}

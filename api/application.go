package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
	"strings"
)

//
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

//
// Params
const (
	Source = "source"
)

//
// ApplicationHandler handles application resource routes.
type ApplicationHandler struct {
	BucketOwner
}

//
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
	routeGroup.Use(Required("applications"))
	routeGroup.GET(ApplicationTagsRoot, h.TagList)
	routeGroup.GET(ApplicationTagsRoot+"/", h.TagList)
	routeGroup.POST(ApplicationTagsRoot, h.TagAdd)
	routeGroup.DELETE(ApplicationTagRoot, h.TagDelete)
	routeGroup.PUT(ApplicationTagsRoot, h.TagReplace, Transaction)
	// Facts
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.facts"))
	routeGroup.GET(ApplicationFactsRoot, h.FactGet)
	routeGroup.GET(ApplicationFactsRoot+"/", h.FactGet)
	routeGroup.POST(ApplicationFactsRoot, h.FactCreate)
	routeGroup.GET(ApplicationFactRoot, h.FactGet)
	routeGroup.PUT(ApplicationFactRoot, h.FactPut)
	routeGroup.DELETE(ApplicationFactRoot, h.FactDelete)
	routeGroup.PUT(ApplicationFactsRoot, h.FactPut, Transaction)
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
	routeGroup.Use(Required("applications.stakeholders"))
	routeGroup.PUT(AppStakeholdersRoot, h.StakeholdersUpdate)
	// Assessments
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.assessments"))
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
	db = db.Omit("Analyses")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	tags := []model.ApplicationTag{}
	db = h.preLoad(h.DB(ctx), clause.Associations)
	result = db.Find(&tags, "ApplicationID = ?", id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	questionnaire, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	membership := assessment.NewMembershipResolver(h.DB(ctx))
	tagsResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resolver := assessment.NewApplicationResolver(m, tagsResolver, membership, questionnaire)
	archetypes, err := resolver.Archetypes()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	archetypeTags, err := resolver.ArchetypeTags()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Application{}
	r.With(m, tags)
	r.WithArchetypes(archetypes)
	r.WithSourcedTags(archetypeTags, "archetype")
	r.WithSourcedTags(resolver.AssessmentTags(), "assessment")
	r.Assessed, err = resolver.Assessed()
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
	var list []model.Application
	db := h.preLoad(h.DB(ctx), clause.Associations)
	db = db.Omit("Analyses")
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	questionnaire, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	membership := assessment.NewMembershipResolver(h.DB(ctx))
	tagsResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	resources := []Application{}
	for i := range list {
		tags := []model.ApplicationTag{}
		db = h.preLoad(h.DB(ctx), clause.Associations)
		result = db.Find(&tags, "ApplicationID = ?", list[i].ID)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
		resolver := assessment.NewApplicationResolver(&list[i], tagsResolver, membership, questionnaire)
		archetypes, aErr := resolver.Archetypes()
		if aErr != nil {
			_ = ctx.Error(aErr)
			return
		}
		archetypeTags, aErr := resolver.ArchetypeTags()
		if aErr != nil {
			_ = ctx.Error(aErr)
			return
		}
		r := Application{}
		r.With(&list[i], tags)
		r.WithArchetypes(archetypes)
		r.WithSourcedTags(archetypeTags, "archetype")
		r.WithSourcedTags(resolver.AssessmentTags(), "assessment")
		r.Assessed, err = resolver.Assessed()
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
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
	result := h.DB(ctx).Omit("Tags").Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	tags := []model.ApplicationTag{}
	if len(r.Tags) > 0 {
		for _, t := range r.Tags {
			tags = append(tags, model.ApplicationTag{TagID: t.ID, ApplicationID: m.ID, Source: t.Source})
		}
		result = h.DB(ctx).Create(&tags)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
	}

	questionnaire, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	membership := assessment.NewMembershipResolver(h.DB(ctx))
	tagsResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resolver := assessment.NewApplicationResolver(m, tagsResolver, membership, questionnaire)
	archetypes, err := resolver.Archetypes()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	archetypeTags, err := resolver.ArchetypeTags()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r.With(m, tags)
	r.WithArchetypes(archetypes)
	r.WithSourcedTags(archetypeTags, "archetype")
	r.WithSourcedTags(resolver.AssessmentTags(), "assessment")
	r.Assessed, err = resolver.Assessed()
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
	result = db.Updates(h.fields(m))
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
			tags = append(tags, model.ApplicationTag{TagID: t.ID, ApplicationID: m.ID, Source: t.Source})
		}
		result = h.DB(ctx).Create(&tags)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
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
// @param id path string true "Application ID"
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
// @param id path string true "Application ID"
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
// @param id path string true "Application ID"
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
// @router /applications/{id}/tags/id [get]
// @param id path string true "Application ID"
func (h ApplicationHandler) TagList(ctx *gin.Context) {
	id := h.pk(ctx)
	app := &model.Application{}
	result := h.DB(ctx).First(app, id)
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
		r.With(list[i].Tag.ID, list[i].Tag.Name, list[i].Source)
		resources = append(resources, r)
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
func (h ApplicationHandler) TagAdd(ctx *gin.Context) {
	id := h.pk(ctx)
	ref := &TagRef{}
	err := h.Bind(ctx, ref)
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
// @param id path string true "Application ID"
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
			appTags = append(appTags, model.ApplicationTag{
				ApplicationID: id,
				TagID:         ref.ID,
				Source:        source,
			})
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
// @param id path string true "Application ID"
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
// @success 200 {object} api.FactMap
// @router /applications/{id}/facts/{source}: [get]
// @param id path string true "Application ID"
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

	facts := FactMap{}
	for i := range list {
		fact := &list[i]
		var v interface{}
		_ = json.Unmarshal(fact.Value, &v)
		facts[fact.Key] = v
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
// @param id path string true "Application ID"
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

	var v interface{}
	_ = json.Unmarshal(list[0].Value, &v)
	h.Respond(ctx, http.StatusOK, v)
}

// FactCreate godoc
// @summary Create a fact.
// @description Create a fact.
// @tags applications
// @accept json
// @produce json
// @success 201
// @router /applications/{id}/facts [post]
// @param id path string true "Application ID"
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
// @param id path string true "Application ID"
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

	value, _ := json.Marshal(f.Value)
	m := &model.Fact{
		Key:           key.Name(),
		Source:        key.Source(),
		ApplicationID: id,
		Value:         value,
	}
	db := h.DB(ctx)
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
// @param id path string true "Application ID"
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
// @param id path string true "Application ID"
// @param source path string true "Fact key"
// @param factmap body api.FactMap true "Fact map"
func (h ApplicationHandler) FactReplace(ctx *gin.Context, key FactKey) {
	id := h.pk(ctx)
	facts := FactMap{}
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
	result = db.Updates(map[string]interface{}{"OwnerID": r.ownerID()})
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
	db := h.preLoad(h.DB(ctx), clause.Associations)
	db = db.Omit("Analyses")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	questionnaire, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	membership := assessment.NewMembershipResolver(h.DB(ctx))
	tagsResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resolver := assessment.NewApplicationResolver(m, tagsResolver, membership, questionnaire)
	archetypes, err := resolver.Archetypes()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	assessments := m.Assessments
	for _, a := range archetypes {
		assessments = append(assessments, a.Assessments...)
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
	m.Sections = q.Sections
	m.Thresholds = q.Thresholds
	m.RiskMessages = q.RiskMessages
	m.CreateUser = h.CurrentUser(ctx)

	resolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	assessment.PrepareForApplication(resolver, application, m)

	result = h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

//
// Application REST resource.
type Application struct {
	Resource
	Name            string      `json:"name" binding:"required"`
	Description     string      `json:"description"`
	Bucket          *Ref        `json:"bucket"`
	Repository      *Repository `json:"repository"`
	Binary          string      `json:"binary"`
	Review          *Ref        `json:"review"`
	Comments        string      `json:"comments"`
	Identities      []Ref       `json:"identities"`
	Tags            []TagRef    `json:"tags"`
	BusinessService *Ref        `json:"businessService"`
	Owner           *Ref        `json:"owner"`
	Contributors    []Ref       `json:"contributors"`
	MigrationWave   *Ref        `json:"migrationWave"`
	Archetypes      []Ref       `json:"archetypes"`
	Assessments     []Ref       `json:"assessments"`
	Assessed        bool        `json:"assessed"`
}

//
// With updates the resource using the model.
func (r *Application) With(m *model.Application, tags []model.ApplicationTag) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.Comments = m.Comments
	r.Binary = m.Binary
	_ = json.Unmarshal(m.Repository, &r.Repository)
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
	for i := range tags {
		ref := TagRef{}
		ref.With(tags[i].TagID, tags[i].Tag.Name, tags[i].Source)
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
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		ref := Ref{}
		ref.With(a.ID, "")
		r.Assessments = append(r.Assessments, ref)
	}
}

//
// WithArchetypes updates the resource with archetypes.
func (r *Application) WithArchetypes(archetypes []model.Archetype) {
	for _, a := range archetypes {
		ref := Ref{}
		ref.With(a.ID, a.Name)
		r.Archetypes = append(r.Archetypes, ref)
	}
}

//
// WithSourcedTags updates the resource with tags derived from assessments.
func (r *Application) WithSourcedTags(tags []model.Tag, source string) {
	for _, t := range tags {
		ref := TagRef{}
		ref.With(t.ID, t.Name, source)
		r.Tags = append(r.Tags, ref)
	}
}

//
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
		m.Repository, _ = json.Marshal(r.Repository)
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

	return
}

//
// Repository REST nested resource.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

//
// Fact REST nested resource.
type Fact struct {
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Source string      `json:"source"`
}

func (r *Fact) With(m *model.Fact) {
	r.Key = m.Key
	r.Source = m.Source
	_ = json.Unmarshal(m.Value, &r.Value)
}

func (r *Fact) Model() (m *model.Fact) {
	m = &model.Fact{}
	m.Key = r.Key
	m.Source = r.Source
	m.Value, _ = json.Marshal(r.Value)
	return
}

//
// FactKey is a fact source and fact name separated by a colon.
//   Example: 'analysis:languages'
//
// A FactKey can be used to identify an anonymous fact.
//   Example: 'languages' or ':languages'
//
// A FactKey can also be used to identify just a source. This use must include the trailing
// colon to distinguish it from an anonymous fact. This is used when listing or replacing
// all facts that belong to a source.
//   Example: 'analysis:"
type FactKey string

//
// Qualify qualifies the name with the source.
func (r *FactKey) Qualify(source string) {
	*r = FactKey(
		strings.Join(
			[]string{source, r.Name()},
			":"))
}

//
// Source returns the source portion of a fact key.
func (r FactKey) Source() (source string) {
	s, _, found := strings.Cut(string(r), ":")
	if found {
		source = s
	}
	return
}

//
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

//
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

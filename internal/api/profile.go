package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/assessment"
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/scm"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"github.com/konveyor/tackle2-hub/shared/tar"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AnalysisProfileHandler handles application Profile resource routes.
type AnalysisProfileHandler struct {
	BaseHandler
}

func (h AnalysisProfileHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("analysis.profiles"))
	routeGroup.GET(api.AnalysisProfileRoute, h.Get)
	routeGroup.GET(api.AnalysisProfilesRoute, h.List)
	routeGroup.GET(api.AnalysisProfilesRoute+"/", h.List)
	routeGroup.GET(api.AnalysisProfileBundle, h.GetBundle)
	routeGroup.POST(api.AnalysisProfilesRoute, h.Create, Transaction)
	routeGroup.PUT(api.AnalysisProfileRoute, h.Update, Transaction)
	routeGroup.DELETE(api.AnalysisProfileRoute, h.Delete, Transaction)
	//
	routeGroup.GET(api.AppAnalysisProfilesRoute, h.AppProfileList)
}

// Get godoc
// @summary Get a Profile by ID.
// @description Get a Profile by ID.
// @tags Profiles
// @produce json
// @success 200 {object} AnalysisProfile
// @router /analysis/profiles/{id} [get]
// @param id path int true "Profile ID"
func (h AnalysisProfileHandler) Get(ctx *gin.Context) {
	r := AnalysisProfile{}
	id := h.pk(ctx)
	m := &model.AnalysisProfile{}
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// GetBundle godoc
// @summary Get a Profile bundle by ID.
// @description Get a Profile bundle by ID.
// @tags ProfileBundles
// @produce application/gzip
// @success 200 {file} bundle
// @router /analysis/profiles/{id}/bundle [get]
// @param id path int true "Profile ID"
func (h AnalysisProfileHandler) GetBundle(ctx *gin.Context) {
	id := h.pk(ctx)
	bundle := ApBundle{}
	tmpDir, err := bundle.Build(h.DB(ctx), id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = nas.RmDir(tmpDir)
	}()
	h.Attachment(ctx, "bundle.tar.gz")
	ctx.Status(http.StatusOK)
	tarWriter := tar.NewWriter(ctx.Writer)
	defer func() {
		tarWriter.Close()
	}()
	err = tarWriter.AddDir(tmpDir)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
}

// List godoc
// @summary List all Profiles.
// @description List all Profiles.
// @tags Profiles
// @produce json
// @success 200 {object} []AnalysisProfile
// @router /analysis/profiles [get]
func (h AnalysisProfileHandler) List(ctx *gin.Context) {
	resources := []AnalysisProfile{}
	var list []model.AnalysisProfile
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		r := AnalysisProfile{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a Profile.
// @description Create a Profile.
// @tags Profiles
// @accept json
// @produce json
// @success 201 {object} AnalysisProfile
// @router /analysis/profiles [post]
// @param Profile body AnalysisProfile true "Profile data"
func (h AnalysisProfileHandler) Create(ctx *gin.Context) {
	r := &AnalysisProfile{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	db = db.Omit(clause.Associations)
	err = db.Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Targets").Replace(m.Targets)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a Profile.
// @description Delete a Profile.
// @tags Profiles
// @success 204
// @router /analysis/profiles/{id} [delete]
// @param id path int true "Profile ID"
func (h AnalysisProfileHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.AnalysisProfile{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = db.Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a Profile.
// @description Update a Profile.
// @tags AnalysisProfiles
// @accept json
// @success 204
// @router /analysis/profiles/{id} [put]
// @param id path int true "Profile ID"
// @param Profile body AnalysisProfile true "Profile data"
func (h AnalysisProfileHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &AnalysisProfile{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	db = db.Omit(clause.Associations)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Targets").Replace(m.Targets)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// AppProfileList godoc
// @summary List analysis profiles.
// @description List analysis profiles mapped to an application through archetypes.
// @tags AnalysisProfiles
// @produce json
// @success 200 {object} []AnalysisProfile
// @router /applications/{id}/analysis/profiles [get]
// @param id path int true "Application ID"
func (h AnalysisProfileHandler) AppProfileList(ctx *gin.Context) {
	resources := []AnalysisProfile{}
	// Fetch application.
	application := &model.Application{}
	id := h.pk(ctx)
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	result := db.First(application, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	// Resolve archetypes and profiles.
	memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var ids []uint
	app := assessment.Application{}
	app.With(application)
	archetypes, err := memberResolver.Archetypes(app)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for _, archetype := range archetypes {
		for _, p := range archetype.Profiles {
			if p.AnalysisProfileID != nil {
				ids = append(ids, *p.AnalysisProfileID)
			}
		}
	}
	// Fetch profiles.
	if len(ids) > 0 {
		db = h.DB(ctx)
		db = db.Preload(clause.Associations)
		var list []model.AnalysisProfile
		err = db.Find(&list, ids).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		for i := range list {
			m := &list[i]
			r := AnalysisProfile{}
			r.With(m)
			resources = append(resources, r)
		}
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AnalysisProfile REST resource.
type AnalysisProfile = resource.AnalysisProfile

// ApBundle defines and builds the application bundle.
type ApBundle struct {
	db      *gorm.DB
	tmpDir  string
	ruleDir string
}

// Build constructs the bundle.
func (b *ApBundle) Build(db *gorm.DB, id uint) (path string, err error) {
	b.db = db
	m, err := b.fetch(id)
	if err != nil {
		return
	}
	b.tmpDir, err = os.MkdirTemp("", "")
	if err != nil {
		return
	}
	b.ruleDir = filepath.Join(
		b.tmpDir,
		"rules")
	err = nas.MkDir(b.ruleDir, 0755)
	if err != nil {
		return
	}
	err = b.addProfile(m)
	if err != nil {
		return
	}
	err = b.addTargets(m)
	if err != nil {
		return
	}
	err = b.addRepository(
		filepath.Join(b.ruleDir, "repository"),
		&m.Repository)
	if err != nil {
		return
	}
	err = b.addFiles(m)
	if err != nil {
		return
	}
	path = b.tmpDir
	return
}

// fetch returns the profile.
func (b *ApBundle) fetch(id uint) (m *model.AnalysisProfile, err error) {
	m = &model.AnalysisProfile{}
	db := b.db.Preload(clause.Associations)
	db = db.Preload("Targets.RuleSet")
	db = db.Preload("Targets.RuleSet.Rules")
	db = db.Preload("Targets.RuleSet.Rules.File")
	err = db.First(m, id).Error
	return
}

// addProfile adds the profile.
func (b *ApBundle) addProfile(m *model.AnalysisProfile) (err error) {
	profile := &AnalysisProfile{}
	profile.With(m)
	profile.Rules.Labels.Included = b.expandIncluded(m)
	path := filepath.Join(b.tmpDir, "profile.yaml")
	f, err := os.Create(path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	content, err := yaml.Marshal(profile)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	_, err = f.Write(content)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// expandInclude ensures the included labels are unique
// superset of the labels listed in targets and in the profile.
func (b *ApBundle) expandIncluded(m *model.AnalysisProfile) (expanded []string) {
	included := make(map[string]byte)
	excluded := make(map[string]byte)
	for _, n := range m.Labels.Excluded {
		excluded[n] = 0
	}
	for _, ruleset := range m.Targets {
		for _, ref := range ruleset.Labels {
			v := ref.Label
			if _, found := excluded[v]; !found {
				included[v] = 0
			}
		}
	}
	for _, v := range m.Labels.Included {
		if _, found := excluded[v]; !found {
			included[v] = 0
		}
	}
	expanded = make([]string, 0, len(included))
	for v := range included {
		expanded = append(expanded, v)
	}
	return
}

// addTargets adds targets referenced by the profile.
func (b *ApBundle) addTargets(m *model.AnalysisProfile) (err error) {
	for _, target := range m.Targets {
		if target.Builtin() {
			continue
		}
		err = b.addRuleSet(target.RuleSet)
		if err != nil {
			return
		}
	}
	return
}

// AddRuleSet adds a ruleSet.
func (b *ApBundle) addRuleSet(ruleSet *model.RuleSet) (err error) {
	ruleSetId := strconv.Itoa(int(ruleSet.ID))
	ruleDir := filepath.Join(b.ruleDir, ruleSetId)
	fileDir := filepath.Join(ruleDir, "files")
	err = nas.MkDir(fileDir, 0755)
	if err != nil {
		return
	}
	for _, rule := range ruleSet.Rules {
		err = b.addRule(fileDir, &rule)
		if err != nil {
			return
		}
	}
	err = b.addRepository(
		filepath.Join(ruleDir, "repository"),
		&ruleSet.Repository)
	if err != nil {
		return
	}
	return
}

// addRule adds a rule.
func (b *ApBundle) addRule(ruleDir string, rule *model.Rule) (err error) {
	if rule.File == nil {
		return
	}
	name := fmt.Sprintf(
		"%d-%s",
		rule.File.ID,
		rule.File.Name)
	pathIn := rule.File.Path
	pathOut := filepath.Join(ruleDir, name)
	reader, err := os.Open(pathIn)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	writer, err := os.Create(pathOut)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	_, err = io.Copy(writer, reader)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// addRepository adds a repository.
func (b *ApBundle) addRepository(rootDir string, repository *model.Repository) (err error) {
	if repository.URL == "" {
		return
	}
	remote := scm.Remote{
		Kind:   repository.Kind,
		URL:    repository.URL,
		Branch: repository.Branch,
	}
	mirror := scm.GetMirror(b.db, remote)
	err = mirror.CopyTo(repository.Path, rootDir)
	if err != nil {
		return
	}
	return
}

// addFiles adds uploaded files referenced in the profile.
func (b *ApBundle) addFiles(m *model.AnalysisProfile) (err error) {
	fileDir := filepath.Join(b.ruleDir, "files")
	err = nas.MkDir(fileDir, 0755)
	if err != nil {
		return
	}
	for _, ref := range m.Files {
		err = b.addFile(fileDir, ref)
		if err != nil {
			return
		}
	}
	return
}

// addFile adds a file.
func (b *ApBundle) addFile(fileDir string, ref json.Ref) (err error) {
	file := &model.File{}
	err = b.db.First(file, ref.ID).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	name := fmt.Sprintf(
		"%d-%s",
		file.ID,
		file.Name)
	pathIn := file.Path
	pathOut := filepath.Join(fileDir, name)
	reader, err := os.Open(pathIn)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	writer, err := os.Create(pathOut)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	_, err = io.Copy(writer, reader)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

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
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"github.com/konveyor/tackle2-hub/scm"
	"github.com/konveyor/tackle2-hub/tar"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Routes
const (
	AnalysisProfilesRoot  = "/analysis/profiles"
	AnalysisProfileRoot   = AnalysisProfilesRoot + "/:id"
	AnalysisProfileBundle = AnalysisProfileRoot + "/bundle"
	//
	AppAnalysisProfilesRoot = ApplicationRoot + "/analysis/profiles"
)

// AnalysisProfileHandler handles application Profile resource routes.
type AnalysisProfileHandler struct {
	BaseHandler
}

func (h AnalysisProfileHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("Profiles"))
	routeGroup.GET(AnalysisProfileRoot, h.Get)
	routeGroup.GET(AnalysisProfilesRoot, h.List)
	routeGroup.GET(AnalysisProfilesRoot+"/", h.List)
	routeGroup.GET(AnalysisProfileBundle, h.GetBundle)
	routeGroup.POST(AnalysisProfilesRoot, h.Create, Transaction)
	routeGroup.PUT(AnalysisProfileRoot, h.Update, Transaction)
	routeGroup.DELETE(AnalysisProfileRoot, h.Delete, Transaction)
	//
	routeGroup.GET(AppAnalysisProfilesRoot, h.AppProfileList)
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
// @produce octet-stream
// @success 200 {object} AnalysisProfile
// @router /analysis/profiles/{id} [get]
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
// @success 201 {object} Profile
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

// InExList include/exclude list.
type InExList = model.InExList

// ApMode analysis mode.
type ApMode struct {
	WithDeps bool `json:"withDeps" yaml:"withDeps"`
}

// ApScope analysis scope.
type ApScope struct {
	WithKnownLibs bool     `json:"withKnownLibs" yaml:"withKnownLibs"`
	Packages      InExList `json:"packages,omitempty" yaml:",omitempty"`
}

// ApRules analysis rules.
type ApRules struct {
	Targets    []Ref       `json:"targets"`
	Labels     InExList    `json:"labels,omitempty" yaml:",omitempty"`
	Files      []Ref       `json:"files,omitempty" yaml:",omitempty"`
	Repository *Repository `json:"repository,omitempty" yaml:",omitempty"`
}

// AnalysisProfile REST resource.
type AnalysisProfile struct {
	Resource    `yaml:",inline"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty" yaml:",omitempty"`
	Mode        ApMode  `json:"mode"`
	Scope       ApScope `json:"scope"`
	Rules       ApRules `json:"rules"`
}

// With updates the resource with the model.
func (r *AnalysisProfile) With(m *model.AnalysisProfile) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Mode.WithDeps = m.WithDeps
	r.Scope.WithKnownLibs = m.WithKnownLibs
	r.Scope.Packages = m.Packages
	r.Rules.Labels = m.Labels
	if m.Repository != (model.Repository{}) {
		repository := Repository(m.Repository)
		r.Rules.Repository = &repository
	}
	r.Rules.Targets = make([]Ref, len(m.Targets))
	for i, t := range m.Targets {
		r.Rules.Targets[i] =
			Ref{
				ID:   t.ID,
				Name: t.Name,
			}
	}
	r.Rules.Files = make([]Ref, len(m.Files))
	for i, f := range m.Files {
		r.Rules.Files[i] = Ref(f)
	}
}

// Model builds a model.
func (r *AnalysisProfile) Model() (m *model.AnalysisProfile) {
	m = &model.AnalysisProfile{}
	m.Name = r.Name
	m.Description = r.Description
	m.WithDeps = r.Mode.WithDeps
	m.WithKnownLibs = r.Scope.WithKnownLibs
	m.Packages = r.Scope.Packages
	m.Labels = r.Rules.Labels
	if r.Rules.Repository != nil {
		m.Repository = model.Repository(*r.Rules.Repository)
	}
	m.Targets = make([]model.Target, len(r.Rules.Targets))
	for i, t := range r.Rules.Targets {
		m.Targets[i] =
			model.Target{
				Model: model.Model{
					ID: t.ID,
				},
				Name: t.Name,
			}
	}
	m.Files = make([]model.Ref, len(r.Rules.Files))
	for i, f := range r.Rules.Files {
		m.Files[i] = model.Ref(f)
	}
	return
}

type ApBundle struct {
	db         *gorm.DB
	tmpDir     string
	ruleSetDir string
}

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
	b.ruleSetDir = filepath.Join(
		b.tmpDir,
		"rulesets")
	err = nas.MkDir(b.ruleSetDir, 0755)
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
		filepath.Join(b.ruleSetDir, "repository"),
		&m.Repository)
	if err != nil {
		return
	}
	path = b.tmpDir
	return
}

func (b *ApBundle) fetch(id uint) (m *model.AnalysisProfile, err error) {
	m = &model.AnalysisProfile{}
	db := b.db.Preload(clause.Associations)
	db = db.Preload("Targets.RuleSet")
	db = db.Preload("Targets.RuleSet.Rules")
	db = db.Preload("Targets.RuleSet.Rules.File")
	err = db.First(m, id).Error
	return
}

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

func (b *ApBundle) addRuleSet(ruleSet *model.RuleSet) (err error) {
	ruleSetId := strconv.Itoa(int(ruleSet.ID))
	ruleSetDir := filepath.Join(b.ruleSetDir, ruleSetId)
	fileDir := filepath.Join(ruleSetDir, "files")
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
		filepath.Join(ruleSetDir, "repository"),
		&ruleSet.Repository)
	if err != nil {
		return
	}
	return
}

func (b *ApBundle) addRule(ruleSetDir string, rule *model.Rule) (err error) {
	if rule.File == nil {
		return
	}
	name := fmt.Sprintf(
		"%d-%s",
		rule.File.ID,
		rule.File.Name)
	pathIn := rule.File.Path
	pathOut := filepath.Join(ruleSetDir, name)
	reader, err := os.Open(pathIn)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	writer, err := os.Create(pathOut)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	_, err = io.Copy(writer, reader)
	_ = reader.Close()
	_ = writer.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

func (b *ApBundle) addRepository(rootDir string, repository *model.Repository) (err error) {
	if repository.URL == "" {
		return
	}
	remote := scm.Remote{
		Kind:   repository.Kind,
		URL:    repository.URL,
		Path:   repository.Path,
		Branch: repository.Branch,
	}
	mirror := scm.GetMirror(b.db, remote)
	err = mirror.CopyTo(rootDir)
	if err != nil {
		return
	}
	return
}

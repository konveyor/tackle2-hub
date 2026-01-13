package archetype

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/api/profile"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestArchetypeCRUD(t *testing.T) {
	for i := range Samples {
		var r api.Archetype
		b, _ := json.Marshal(Samples[i])
		_ = json.Unmarshal(b, &r)
		t.Run(r.Name, func(t *testing.T) {
			// generator
			genA := &api.Generator{Name: "genA", Kind: "helm"}
			genB := &api.Generator{Name: "genB", Kind: "helm"}
			genC := &api.Generator{Name: "genC", Kind: "helm"}
			genD := &api.Generator{Name: "genD", Kind: "helm"}
			for _, g := range []*api.Generator{genA, genB, genC, genD} {
				err := RichClient.Generator.Create(g)
				assert.Must(t, err)
			}
			defer func() {
				_ = RichClient.Generator.Delete(genA.ID)
				_ = RichClient.Generator.Delete(genB.ID)
				_ = RichClient.Generator.Delete(genC.ID)
				_ = RichClient.Generator.Delete(genD.ID)
			}()
			// analysis profile.
			ap1 := &profile.Base
			err := RichClient.AnalysisProfile.Create(ap1)
			assert.Must(t, err)
			defer func() {
				_ = RichClient.AnalysisProfile.Delete(ap1.ID)
			}()
			// Create.
			for i := range r.Profiles {
				p := &r.Profiles[i]
				p.AnalysisProfile = &api.Ref{
					ID:   ap1.ID,
					Name: ap1.Name,
				}
				p.Generators = append(
					p.Generators,
					api.Ref{ID: genA.ID, Name: genA.Name})
			}
			err = Archetype.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Archetype.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !assert.Eq(r, got) {
				t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
			}

			nProfiles := len(r.Profiles)

			// Update.
			// Change the name and add a profile.
			r.Name += "-Updated"
			r.Profiles = append(
				r.Profiles,
				api.TargetProfile{
					Name: "Added",
					Generators: []api.Ref{
						{
							ID:   genD.ID,
							Name: genD.Name,
						}}})
			err = Archetype.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Archetype.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !assert.Eq(r.Name, got.Name) {
				t.Errorf("Different error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
			}
			if !assert.Eq(nProfiles+1, len(got.Profiles)) {
				t.Errorf("Different error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
			}

			// add a profile and add generators and ensure ordering preserved.
			for i := range r.Profiles {
				p := &r.Profiles[i]
				p.Generators = append(
					p.Generators,
					api.Ref{ID: genC.ID, Name: genC.Name},
					api.Ref{ID: genB.ID, Name: genB.Name},
				)
			}
			err = Archetype.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}
			got, err = Archetype.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			for i := range got.Profiles {
				p := &got.Profiles[i]
				p.CreateTime = r.Profiles[i].CreateTime
			}
			// transfer updateUser and profiles Ids.
			// needs to be ignored in Eq().
			r.UpdateUser = got.UpdateUser
			for i := range got.Profiles {
				if i < len(r.Profiles) {
					(&r.Profiles[i]).ID = got.Profiles[i].ID
				}
			}
			if !assert.Eq(r, got) {
				t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
			}

			// Delete.
			err = Archetype.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Archetype.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestArchetypeList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Archetype.Create(&sample))
		samples[name] = sample
	}

	got, err := Archetype.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Archetype.Delete(r.ID))
	}
}

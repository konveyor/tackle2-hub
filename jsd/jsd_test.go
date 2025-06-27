package jsd

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
)

func TestSafeMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	in := map[any]any{
		0:   "A",
		"1": "B",
	}
	out := JsonSafe(in)
	mp, cast := out.(map[string]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(mp)).To(gomega.Equal(len(in)))
	g.Expect(mp["0"]).To(gomega.Equal(in[0]))
}

func TestSafeArrayStruct(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	type Thing struct {
		ID       uint `json:"id"`
		Activity int  `json:"activity"`
	}
	in := []Thing{
		{ID: 1, Activity: 18},
		{ID: 2, Activity: 48},
	}
	out := JsonSafe(in)
	list, cast := out.([]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(list)).To(gomega.Equal(len(in)))
	g.Expect(list[0].(Thing).ID).To(gomega.Equal(in[0].ID))
	g.Expect(list[1].(Thing).ID).To(gomega.Equal(in[1].ID))
	g.Expect(list[0].(Thing).Activity).To(gomega.Equal(in[0].Activity))
	g.Expect(list[1].(Thing).Activity).To(gomega.Equal(in[1].Activity))

	m, err := json.Marshal(list[0])
	g.Expect(err).To(gomega.BeNil())
	m2 := make(map[string]any)
	err = json.Unmarshal(m, &m2)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m2["id"]).To(gomega.Equal(float64(1)))
	g.Expect(m2["activity"]).To(gomega.Equal(float64(18)))
}

func TestSafeData(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	type Data struct {
		Any any
	}
	in := Data{
		Any: map[any]any{
			0:   "A",
			"1": "B",
		},
	}
	out := JsonSafe(in)
	d, cast := out.(Data)
	g.Expect(cast).To(gomega.BeTrue())
	dAny, cast := d.Any.(map[string]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(dAny)).To(gomega.Equal(2))
	g.Expect(dAny["0"]).To(gomega.Equal(in.Any.(map[any]any)[0]))
	g.Expect(dAny["1"]).To(gomega.Equal(in.Any.(map[any]any)["1"]))
}

func TestSafeDataPtr(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	type Data struct {
		Any any
	}
	in := &Data{
		Any: map[any]any{
			0:   "A",
			"1": "B",
			2: Data{
				Any: map[any]any{
					2:   "2A",
					"3": "3B",
				},
			},
		},
	}
	out := JsonSafe(in)
	d, cast := out.(Data)
	g.Expect(cast).To(gomega.BeTrue())
	dAny, cast := d.Any.(map[string]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(dAny)).To(gomega.Equal(3))
	g.Expect(dAny["0"]).To(gomega.Equal(in.Any.(map[any]any)[0]))
	g.Expect(dAny["1"]).To(gomega.Equal(in.Any.(map[any]any)["1"]))
	d2Any := dAny["2"]
	d2d, cast := d2Any.(Data)
	g.Expect(cast).To(gomega.BeTrue())
	d2dAny, cast := d2d.Any.(map[string]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(d2dAny["2"]).To(gomega.Equal("2A"))
	g.Expect(d2dAny["3"]).To(gomega.Equal("3B"))
}

package model

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestMapMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	js := jsonSerializer{}
	in := map[any]any{
		0:   "A",
		"1": "B",
	}
	out := js.jMap(in)
	mp, cast := out.(map[string]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(mp)).To(gomega.Equal(len(in)))
	g.Expect(mp["0"]).To(gomega.Equal(in[0]))
}

func TestMapArrayStruct(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	js := jsonSerializer{}
	in := []Attachment{
		{ID: 1, Activity: 18},
		{ID: 2, Activity: 48},
	}
	out := js.jMap(in)
	list, cast := out.([]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(list)).To(gomega.Equal(len(in)))
	g.Expect(list[0].(Attachment).ID).To(gomega.Equal(in[0].ID))
	g.Expect(list[1].(Attachment).ID).To(gomega.Equal(in[1].ID))
	g.Expect(list[0].(Attachment).Activity).To(gomega.Equal(in[0].Activity))
	g.Expect(list[1].(Attachment).Activity).To(gomega.Equal(in[1].Activity))
}

func TestMapData(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	js := jsonSerializer{}
	in := Data{
		Any: map[any]any{
			0:   "A",
			"1": "B",
		},
	}
	out := js.jMap(in)
	d, cast := out.(Data)
	g.Expect(cast).To(gomega.BeTrue())
	dAny, cast := d.Any.(map[string]any)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(dAny)).To(gomega.Equal(2))
	g.Expect(dAny["0"]).To(gomega.Equal(in.Any.(map[any]any)[0]))
	g.Expect(dAny["1"]).To(gomega.Equal(in.Any.(map[any]any)["1"]))
}

func TestMapDataPtr(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	js := jsonSerializer{}
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
	out := js.jMap(in)
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

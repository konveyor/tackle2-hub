package jsd

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

func TestSafeMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	in := map[any]any{
		0:   "A",
		"1": "B",
	}
	out := JsonSafe(in)
	mp, cast := out.(Map)
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
	m2 := make(Map)
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
	dAny, cast := d.Any.(Map)
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
	dAny, cast := d.Any.(Map)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(len(dAny)).To(gomega.Equal(3))
	g.Expect(dAny["0"]).To(gomega.Equal(in.Any.(map[any]any)[0]))
	g.Expect(dAny["1"]).To(gomega.Equal(in.Any.(map[any]any)["1"]))
	d2Any := dAny["2"]
	d2d, cast := d2Any.(Data)
	g.Expect(cast).To(gomega.BeTrue())
	d2dAny, cast := d2d.Any.(Map)
	g.Expect(cast).To(gomega.BeTrue())
	g.Expect(d2dAny["2"]).To(gomega.Equal("2A"))
	g.Expect(d2dAny["3"]).To(gomega.Equal("3B"))
}

func TestSchema(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	schemaText := `
domain: people
variant: manager
subject: n/a
versions:
  - definition:
      '$schema': https://json-schema.org/draft/2020-12/schema
      title: Person
      type: object
      required:
        - name
        - age
        - phone
      properties:
        name:
          type: string
        age:
          type: integer
          minimum: 0
        phone:
          type: string
          pattern: '^\d{3}-\d{3}-\d{4}$'
          description: Phone number in the format 555-444-8888
  - definition:
      '$schema': https://json-schema.org/draft/2020-12/schema
      title: Person
      type: object
      required:
        - name
        - age
        - phone
      properties:
        name:
          type: string
        age:
          type: integer
          minimum: 0
        phone:
          type: object
          required:
            - npa
            - nxx
            - number
          properties:
            npa:
              type: string
              pattern: '^\d{3}$'
              description: 3-digit area code
            nxx:
              type: string
              pattern: '^\d{3}$'
              description: 3-digit exchange code
            number:
              type: string
              pattern: '^\d{4}$'
              description: 4-digit line number
    migration: >
      .phone |= (split("-") 
        | {
            "npa": .[0],
            "nxx": .[1],
            "number": .[2]
          })
`

	d1Text := `
name: Elmer
age: 44
phone: "555-214-1438"
`
	d2Text := `
name: Elmer
age: 44
phone: 
  npa: "555"
  nxx: "214"
  number: "1438"
`
	// v2
	s2 := &Schema{}
	b := []byte(schemaText)
	err := yaml.Unmarshal(b, &s2)
	g.Expect(err).To(gomega.BeNil())
	err = s2.IsValid()
	g.Expect(err).To(gomega.BeNil())

	// v1
	s1 := &Schema{}
	b = []byte(schemaText)
	err = yaml.Unmarshal(b, &s1)
	g.Expect(err).To(gomega.BeNil())
	s1.Versions = s1.Versions[:1]
	err = s1.IsValid()
	g.Expect(err).To(gomega.BeNil())

	// validate v1 document
	var d1 Map
	b = []byte(d1Text)
	err = yaml.Unmarshal(b, &d1)
	g.Expect(err).To(gomega.BeNil())
	err = s1.Validate(d1)
	g.Expect(err).To(gomega.BeNil())
	err = s2.Validate(d1)
	g.Expect(err).ToNot(gomega.BeNil())

	// validate v2 document
	var d2 Map
	b = []byte(d2Text)
	err = yaml.Unmarshal(b, &d2)
	g.Expect(err).To(gomega.BeNil())
	err = s2.Validate(d2)
	g.Expect(err).To(gomega.BeNil())
	err = s1.Validate(d2)
	g.Expect(err).ToNot(gomega.BeNil())

	// migrate
	// v1 => v1.
	m1, current, err := s1.Migrate(d1, 0)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m1).To(gomega.Equal(d1))
	g.Expect(0).To(gomega.Equal(current))
	// v1 => v2.
	m2, current, err := s2.Migrate(d1, current)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m2).To(gomega.Equal(d2))
	g.Expect(1).To(gomega.Equal(current))

	// digest.
	digest1 := s1.Digest()
	digest2 := s2.Digest()
	g.Expect("4F3BC99FE3BC3D35").To(gomega.Equal(digest1))
	g.Expect("13D15FE7A1B5269D").To(gomega.Equal(digest2))
	g.Expect(digest1).ToNot(gomega.Equal(digest2))
}

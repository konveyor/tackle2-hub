package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationTagCRUD(t *testing.T) {
	// Setup tags.
	tag1 := &api.Tag{
		Name: "Tag1",
		Category: api.Ref{
			ID: 1, // Category from seeds.
		},
	}
	err := RichClient.Tag.Create(tag1)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Tag.Delete(tag1.ID)
	}()

	tag2 := &api.Tag{
		Name: "Tag2",
		Category: api.Ref{
			ID: 2, // Category from seeds.
		},
	}
	err = RichClient.Tag.Create(tag2)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Tag.Delete(tag2.ID)
	}()

	tag3 := &api.Tag{
		Name: "Tag3",
		Category: api.Ref{
			ID: 1, // Category from seeds.
		},
	}
	err = RichClient.Tag.Create(tag3)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Tag.Delete(tag3.ID)
	}()

	// Setup application.
	application := &api.Application{
		Name: t.Name(),
	}
	err = Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = Application.Delete(application.ID)
	}()

	// Test deprecated Application.Tags() API.
	tags := Application.Tags(application.ID)

	// Add first tag.
	err = tags.Add(tag1.ID)
	assert.Should(t, err)

	// List tags - should have 1.
	list, err := tags.List()
	assert.Should(t, err)
	if len(list) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(list))
	}
	if list[0].ID != tag1.ID {
		t.Errorf("Expected tag ID %d, got %d", tag1.ID, list[0].ID)
	}

	// Add second tag.
	err = tags.Add(tag2.ID)
	assert.Should(t, err)

	// List tags - should have 2.
	list, err = tags.List()
	assert.Should(t, err)
	if len(list) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(list))
	}

	// Test Ensure (should not error on duplicate).
	err = tags.Ensure(tag2.ID)
	assert.Should(t, err)

	// List should still have 2.
	list, err = tags.List()
	assert.Should(t, err)
	if len(list) != 2 {
		t.Errorf("Expected 2 tags after Ensure, got %d", len(list))
	}

	// Test Replace with source.
	source := "test-source"
	tags.Source(source)
	err = tags.Replace([]uint{tag3.ID})
	assert.Should(t, err)

	// List with source - should have 1.
	list, err = tags.List()
	assert.Should(t, err)
	if len(list) != 1 {
		t.Errorf("Expected 1 tag with source, got %d", len(list))
	}
	if list[0].ID != tag3.ID {
		t.Errorf("Expected tag ID %d, got %d", tag3.ID, list[0].ID)
	}
	if list[0].Source != source {
		t.Errorf("Expected source %s, got %s", source, list[0].Source)
	}

	// List without source - should have all tags (2 + 1).
	tagsNoSource := Application.Tags(application.ID)
	list, err = tagsNoSource.List()
	assert.Should(t, err)
	if len(list) != 3 {
		t.Errorf("Expected 3 tags total, got %d", len(list))
	}

	// Delete tag with source.
	err = tags.Delete(tag3.ID)
	assert.Should(t, err)

	// List with source - should have 0.
	list, err = tags.List()
	assert.Should(t, err)
	if len(list) != 0 {
		t.Errorf("Expected 0 tags with source after delete, got %d", len(list))
	}

	// List without source - should still have 2 (tag1, tag2).
	list, err = tagsNoSource.List()
	assert.Should(t, err)
	if len(list) != 2 {
		t.Errorf("Expected 2 tags without source, got %d", len(list))
	}

	// Delete remaining tags.
	err = tagsNoSource.Delete(tag1.ID)
	assert.Should(t, err)
	err = tagsNoSource.Delete(tag2.ID)
	assert.Should(t, err)

	// List - should have 0.
	list, err = tagsNoSource.List()
	assert.Should(t, err)
	if len(list) != 0 {
		t.Errorf("Expected 0 tags after cleanup, got %d", len(list))
	}
}

func TestApplicationTagCRUD_Select(t *testing.T) {
	// Setup tags.
	tag1 := &api.Tag{
		Name: "Tag1",
		Category: api.Ref{
			ID: 1, // Category from seeds.
		},
	}
	err := RichClient.Tag.Create(tag1)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Tag.Delete(tag1.ID)
	}()

	tag2 := &api.Tag{
		Name: "Tag2",
		Category: api.Ref{
			ID: 2, // Category from seeds.
		},
	}
	err = RichClient.Tag.Create(tag2)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Tag.Delete(tag2.ID)
	}()

	tag3 := &api.Tag{
		Name: "Tag3",
		Category: api.Ref{
			ID: 1, // Category from seeds.
		},
	}
	err = RichClient.Tag.Create(tag3)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Tag.Delete(tag3.ID)
	}()

	// Setup application.
	application := &api.Application{
		Name: t.Name(),
	}
	err = Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = Application.Delete(application.ID)
	}()

	// Test new Application.Select().Tag API.
	selected := Application.Select(application.ID)

	// Add first tag.
	err = selected.Tag.Add(tag1.ID)
	assert.Should(t, err)

	// List tags - should have 1.
	list, err := selected.Tag.List()
	assert.Should(t, err)
	if len(list) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(list))
	}
	if list[0].ID != tag1.ID {
		t.Errorf("Expected tag ID %d, got %d", tag1.ID, list[0].ID)
	}

	// Add second tag.
	err = selected.Tag.Add(tag2.ID)
	assert.Should(t, err)

	// List tags - should have 2.
	list, err = selected.Tag.List()
	assert.Should(t, err)
	if len(list) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(list))
	}

	// Test Ensure (should not error on duplicate).
	err = selected.Tag.Ensure(tag2.ID)
	assert.Should(t, err)

	// List should still have 2.
	list, err = selected.Tag.List()
	assert.Should(t, err)
	if len(list) != 2 {
		t.Errorf("Expected 2 tags after Ensure, got %d", len(list))
	}

	// Test Replace with source.
	source := "test-source"
	tagWithSource := selected.Tag.Source(source)
	err = tagWithSource.Replace([]uint{tag3.ID})
	assert.Should(t, err)

	// List with source - should have 1.
	list, err = tagWithSource.List()
	assert.Should(t, err)
	if len(list) != 1 {
		t.Errorf("Expected 1 tag with source, got %d", len(list))
	}
	if list[0].ID != tag3.ID {
		t.Errorf("Expected tag ID %d, got %d", tag3.ID, list[0].ID)
	}
	if list[0].Source != source {
		t.Errorf("Expected source %s, got %s", source, list[0].Source)
	}

	// List without source - should have all tags (2 + 1).
	list, err = selected.Tag.List()
	assert.Should(t, err)
	if len(list) != 3 {
		t.Errorf("Expected 3 tags total, got %d", len(list))
	}

	// Delete tag with source.
	err = tagWithSource.Delete(tag3.ID)
	assert.Should(t, err)

	// List with source - should have 0.
	list, err = tagWithSource.List()
	assert.Should(t, err)
	if len(list) != 0 {
		t.Errorf("Expected 0 tags with source after delete, got %d", len(list))
	}

	// List without source - should still have 2 (tag1, tag2).
	list, err = selected.Tag.List()
	assert.Should(t, err)
	if len(list) != 2 {
		t.Errorf("Expected 2 tags without source, got %d", len(list))
	}

	// Delete remaining tags.
	err = selected.Tag.Delete(tag1.ID)
	assert.Should(t, err)
	err = selected.Tag.Delete(tag2.ID)
	assert.Should(t, err)

	// List - should have 0.
	list, err = selected.Tag.List()
	assert.Should(t, err)
	if len(list) != 0 {
		t.Errorf("Expected 0 tags after cleanup, got %d", len(list))
	}
}

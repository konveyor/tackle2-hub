package setting

import (
	"testing"
)

func TestSettingCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Key, func(t *testing.T) {
			// Create.
			err := Setting.Create(&r)
			if err != nil {
				t.Error(err)
			}

			// Get.
			gotValue := ""
			err = Setting.Get(r.Key, &gotValue)
			if err != nil {
				t.Error(err)
			}
			if gotValue != r.Value {
				t.Errorf("Different response error. Got %v, expected %v", gotValue, r)
			}

			// Update.
			updateValue := "data-updated"
			r.Value = updateValue
			err = Setting.Update(&r)
			if err != nil {
				t.Error(err)
			}

			err = Setting.Get(r.Key, &r.Value)
			if err != nil {
				t.Error(err)
			}
			if r.Value != updateValue {
				t.Errorf("Different Setting Value error. Got %s, expected %s", gotValue, updateValue)
			}

			// Delete.
			err = Setting.Delete(r.Key)
			if err != nil {
				t.Error(err)
			}

			err = Setting.Get(r.Key, gotValue)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestSettingList(t *testing.T) {
	got, err := Setting.List()
	if err != nil {
		t.Error(err)
	}
	if len(got) < 1 {
		t.Errorf("Got empty Settings list.")
	}
}

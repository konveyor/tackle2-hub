package analysis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/application"
	"github.com/konveyor/tackle2-hub/test/assert"
)

//
// Test application analysis
// "Basic" means that there no other dependencies than the application itself (no need prepare credentials, proxy, etc)
func TestBasicAnalysis(t *testing.T) {
	tests := []TC{
		{
			Name:        "Pathfinder cloud-readiness",
			Application: application.PathfinderGit,
			Task: api.Task{
				Addon: "windup",
				State: "Ready",
			},
			TaskData: `{
				"mode": {
					"artifact": "",
					"binary": false,
					"withDeps": false,
					"diva": true
				},
				"output": "/windup/report",
				"rules": {
					"path": "",
					"tags": {
						"excluded": [ ]
					}
				},
				"scope": {
					"packages": {
						"excluded": [ ],
						"included": [ ]
					},
					"withKnown": false
				},
				"sources": [ ],
				"targets": [
					"cloud-readiness"
				]
			  }`,
		},
	}

	// Test using "richclient" methods (preffered way).
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			// Create the application.
			assert.Should(t, RichClient.Application.Create(&tc.Application))

			// Prepare and submit the analyze task.
			json.Unmarshal([]byte(tc.TaskData), &tc.Task.Data)
			tc.Task.Application = &api.Ref{ID: tc.Application.ID}
			assert.Should(t, RichClient.Task.Create(&tc.Task))

			// Wait until task finishes
			var task *api.Task
			var err error
			for i := 0; i < Retry; i++ {
				task, err = RichClient.Task.Get(tc.Task.ID)
				if err != nil || task.State == "Succeeded" || task.State == "Failed" {
					break
				}
				time.Sleep(Wait)
			}

			if task.State != "Succeeded" {
				t.Errorf("Analyze Task failed. Details: %+v", task)
			}

			// TODO: check the report content here.

			// Cleanup.
			assert.Must(t, RichClient.Application.Delete(tc.Application.ID))

		})
	}
}

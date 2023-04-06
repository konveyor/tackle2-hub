package analysis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/application"
	c "github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/api/task"
)

//
// Test application analysis
// "Basic" means that there no other dependencies than the application itself (no need prepare credentials, proxy, etc)
func TestBasicAnalysis(t *testing.T) {
	tests := []TC{
		{
			Name:        "Pathfinder cloud-readiness",
			Application: application.Samples()[0],
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
		t.Run(tc.Name+"_with_richclient", func(t *testing.T) {
			// Create the application.
			c.Should(t, application.Create(&tc.Application))

			// Prepare and submit the analyze task.
			json.Unmarshal([]byte(tc.TaskData), &tc.Task.Data)
			tc.Task.Application = &api.Ref{ID: tc.Application.ID}
			c.Should(t, task.Create(&tc.Task))

			// Wait until task finishes
			for i := 0; i < Retry; i++ {
				err := task.Get(&tc.Task)
				if err != nil || tc.Task.State == "Succeeded" || tc.Task.State == "Failed" {
					// Proceed to Task result check
					break
				}
				time.Sleep(Wait)
			}

			// t.Log(tc.Task)
			time.Sleep(90 * time.Second)

			if tc.Task.State != "Succeeded" {
				t.Errorf("Analyze Task failed. Details: %+v", tc.Task)
			}

			// TODO: check the report content here.

			// Cleanup.
			c.Must(t, application.Delete(&tc.Application))

		})
	}

	// The same test with plain addon client methods (for demonstration purposes).
	for _, tc := range tests {
		t.Run(tc.Name+"_plain_api", func(t *testing.T) {
			// Create the application.
			err := Client.Post(api.ApplicationsRoot, &tc.Application)
			if err != nil {
				t.Fatalf(err.Error())
			}

			// Prepare and submit the analyze task.
			json.Unmarshal([]byte(tc.TaskData), &tc.Task.Data)
			tc.Task.Application = &api.Ref{ID: tc.Application.ID}
			err = Client.Post(api.TasksRoot, &tc.Task)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Wait until task finishes
			for i := 0; i < Retry; i++ {
				err := Client.Get(c.Path(api.TaskRoot, c.Params{api.ID: tc.Task.ID}), &tc.Task)
				if err != nil || tc.Task.State == "Succeeded" || tc.Task.State == "Failed" {
					// Proceed to Task result check
					break
				}
				//t.Log(tc.Task)
				time.Sleep(Wait)
			}

			if tc.Task.State != "Succeeded" {
				t.Errorf("Analyze Task failed. Details: %+v", tc.Task)
			}

			// TODO: check the report content here.

			// Cleanup.
			err = Client.Delete(c.Path(api.ApplicationRoot, c.Params{api.ID: tc.Application.ID}))
			if err != nil {
				t.Fatalf(err.Error())
			}

		})
	}
}

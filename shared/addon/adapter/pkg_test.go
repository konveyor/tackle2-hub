package adapter

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/api/k8s"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
	"github.com/konveyor/tackle2-hub/shared/task"
	. "github.com/onsi/gomega"
)

func TestTaskLoad(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 42
				r.Name = "Test Task"
				r.Addon = "test-addon"
				r.Data = api.Map{
					"mode": api.Map{
						"binary": true,
					},
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	// Set task ID in settings
	Settings.Task = 42

	// Load task
	adapter.Load()

	// Verify task loaded
	g.Expect(adapter.task).NotTo(BeNil())
	g.Expect(adapter.task.ID).To(Equal(uint(42)))
	g.Expect(adapter.task.Name).To(Equal("Test Task"))
	g.Expect(adapter.task.Addon).To(Equal("test-addon"))
}

func TestTaskApplication(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Application = &api.Ref{ID: 100}
			case *api.Application:
				r.ID = 100
				r.Name = "Test Application"
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get application
	app, err := adapter.Task.Application()
	g.Expect(err).To(BeNil())
	g.Expect(app).NotTo(BeNil())
	g.Expect(app.ID).To(Equal(uint(100)))
	g.Expect(app.Name).To(Equal("Test Application"))
}

func TestTaskApplicationNotSpecified(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Application = nil
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get application - should fail
	app, err := adapter.Task.Application()
	g.Expect(err).NotTo(BeNil())
	g.Expect(app).To(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("application not specified"))
}

func TestTaskPlatform(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Platform = &api.Ref{ID: 200}
			case *api.Platform:
				r.ID = 200
				r.Name = "Test Platform"
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get platform
	platform, err := adapter.Task.Platform()
	g.Expect(err).To(BeNil())
	g.Expect(platform).NotTo(BeNil())
	g.Expect(platform.ID).To(Equal(uint(200)))
	g.Expect(platform.Name).To(Equal("Test Platform"))
}

func TestTaskAddon(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Addon = "test-addon"
				r.Extensions = []string{"ext1", "ext2"}
			case *api.Addon:
				r.Name = "test-addon"
				r.Extensions = []api.Extension{
					{Name: "ext1", Container: k8s.Container{}},
					{Name: "ext2", Container: k8s.Container{}},
					{Name: "ext3", Container: k8s.Container{}},
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon without injection
	addon, err := adapter.Task.Addon(false)
	g.Expect(err).To(BeNil())
	g.Expect(addon).NotTo(BeNil())
	g.Expect(addon.Name).To(Equal("test-addon"))
	g.Expect(len(addon.Extensions)).To(Equal(2))
	g.Expect(addon.Extensions[0].Name).To(Equal("ext1"))
	g.Expect(addon.Extensions[1].Name).To(Equal("ext2"))
}

func TestTaskAddonWithInjection(t *testing.T) {
	g := NewGomegaWithT(t)

	// Set environment variable for injection
	os.Setenv("_EXT_EXT1_VAR1", "value1")
	t.Cleanup(func() {
		os.Unsetenv("_EXT_EXT1_VAR1")
	})

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Addon = "test-addon"
				r.Extensions = []string{"ext1"}
			case *api.Addon:
				r.Name = "test-addon"
				r.Extensions = []api.Extension{
					{
						Name: "ext1",
						Container: k8s.Container{
							Env: []k8s.EnvVar{
								{Name: "VAR1"},
							},
						},
						Metadata: api.Map{
							"key": "$(VAR1)",
						},
					},
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon with injection
	addon, err := adapter.Task.Addon(true)
	g.Expect(err).To(BeNil())
	g.Expect(addon).NotTo(BeNil())
	metadata := addon.Extensions[0].Metadata.(map[string]any)
	g.Expect(metadata["key"]).To(Equal("value1"))
}

func TestTaskDataWith(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Data = api.Map{
					"name":  "test",
					"value": 123,
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Unmarshal data
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var data TestData
	err := adapter.Task.DataWith(&data)
	g.Expect(err).To(BeNil())
	g.Expect(data.Name).To(Equal("test"))
	g.Expect(data.Value).To(Equal(123))
}

func TestTaskReportLifecycle(t *testing.T) {
	g := NewGomegaWithT(t)

	var deleteCalled bool
	var createCalled bool
	var updateCalled bool
	var createdReport *api.TaskReport
	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			deleteCalled = true
			return
		},
		DoPost: func(path string, object any) (err error) {
			createCalled = true
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
				createdReport = r
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			updateCalled = true
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Started - should delete old report and create new one
	adapter.Started()
	g.Expect(deleteCalled).To(BeTrue())
	g.Expect(createCalled).To(BeTrue())
	g.Expect(createdReport).NotTo(BeNil())
	g.Expect(createdReport.Status).To(Equal(task.Running))
	g.Expect(adapter.report.ID).To(Equal(uint(100)))

	// Activity - should update report
	updateCalled = false
	adapter.Activity("Processing item 1")
	g.Expect(updateCalled).To(BeTrue())
	g.Expect(updatedReport).NotTo(BeNil())
	g.Expect(len(updatedReport.Activity)).To(Equal(1))
	g.Expect(updatedReport.Activity[0]).To(Equal("Processing item 1"))

	// Total - should update report
	updateCalled = false
	adapter.Total(10)
	g.Expect(updateCalled).To(BeTrue())
	g.Expect(updatedReport.Total).To(Equal(10))

	// Increment - should update report
	updateCalled = false
	adapter.Increment()
	g.Expect(updateCalled).To(BeTrue())
	g.Expect(updatedReport.Completed).To(Equal(1))

	// Completed - should update report
	updateCalled = false
	adapter.Completed(5)
	g.Expect(updateCalled).To(BeTrue())
	g.Expect(updatedReport.Completed).To(Equal(5))

	// Error - should update report
	updateCalled = false
	adapter.Error(api.TaskError{
		Severity:    "Warning",
		Description: "Test warning",
	})
	g.Expect(updateCalled).To(BeTrue())
	g.Expect(len(updatedReport.Errors)).To(Equal(1))
	g.Expect(updatedReport.Errors[0].Severity).To(Equal("Warning"))

	// Succeeded - should update report
	updateCalled = false
	adapter.Succeeded()
	g.Expect(updateCalled).To(BeTrue())
	g.Expect(updatedReport.Status).To(Equal(task.Succeeded))
	g.Expect(updatedReport.Completed).To(Equal(updatedReport.Total))
}

func TestTaskFailed(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Failed - should update report
	adapter.Failed("Test failure: %s", "error message")
	g.Expect(updatedReport).NotTo(BeNil())
	g.Expect(updatedReport.Status).To(Equal(task.Failed))
	g.Expect(len(updatedReport.Errors)).To(Equal(1))
	g.Expect(updatedReport.Errors[0].Severity).To(Equal("Error"))
	g.Expect(updatedReport.Errors[0].Description).To(Equal("Test failure: error message"))
}

func TestTaskAttach(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Add activity
	adapter.Activity("Processing")
	adapter.Activity("Analyzing")

	// Attach to last activity
	file := &api.File{Name: "test.txt"}
	file.ID = 42
	adapter.Attach(file)
	g.Expect(len(updatedReport.Attached)).To(Equal(1))
	g.Expect(updatedReport.Attached[0].ID).To(Equal(uint(42)))
	g.Expect(updatedReport.Attached[0].Name).To(Equal("test.txt"))
	g.Expect(updatedReport.Attached[0].Activity).To(Equal(2))

	// Attach at specific activity
	file2 := &api.File{Name: "test2.txt"}
	file2.ID = 43
	adapter.AttachAt(file2, 1)
	g.Expect(len(updatedReport.Attached)).To(Equal(2))
	g.Expect(updatedReport.Attached[1].ID).To(Equal(uint(43)))
	g.Expect(updatedReport.Attached[1].Activity).To(Equal(1))

	// Attach same file again - should not duplicate
	adapter.Attach(file)
	g.Expect(len(updatedReport.Attached)).To(Equal(2))
}

func TestTaskResult(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Set result
	result := api.Map{
		"findings": 42,
		"status":   "complete",
	}
	adapter.Result(result)
	g.Expect(updatedReport).NotTo(BeNil())
	g.Expect(updatedReport.Result).To(Equal(result))
}

func TestEnvInjector(t *testing.T) {
	g := NewGomegaWithT(t)

	// Set environment variables
	os.Setenv("_EXT_EXT1_DB_HOST", "localhost")
	os.Setenv("_EXT_EXT1_DB_PORT", "5432")
	t.Cleanup(func() {
		os.Unsetenv("_EXT_EXT1_DB_HOST")
		os.Unsetenv("_EXT_EXT1_DB_PORT")
	})

	// Create extension with metadata containing env references
	extension := &api.Extension{
		Name: "ext1",
		Container: k8s.Container{
			Env: []k8s.EnvVar{
				{Name: "DB_HOST"},
				{Name: "DB_PORT"},
			},
		},
		Metadata: api.Map{
			"database": api.Map{
				"host": "$(DB_HOST)",
				"port": "$(DB_PORT)",
			},
			"description": "Connect to $(DB_HOST):$(DB_PORT)",
		},
	}

	// Inject environment variables
	injector := EnvInjector{}
	injector.Inject(extension)

	// Verify injection
	metadata := extension.Metadata.(map[string]any)
	g.Expect(metadata["database"].(map[string]any)["host"]).To(Equal("localhost"))
	g.Expect(metadata["database"].(map[string]any)["port"]).To(Equal("5432"))
	g.Expect(metadata["description"]).To(Equal("Connect to localhost:5432"))
}

func TestEnvInjectorNested(t *testing.T) {
	g := NewGomegaWithT(t)

	// Set environment variables
	os.Setenv("_EXT_EXT1_TOKEN", "secret123")
	t.Cleanup(func() {
		os.Unsetenv("_EXT_EXT1_TOKEN")
	})

	// Create extension with nested metadata
	extension := &api.Extension{
		Name: "ext1",
		Container: k8s.Container{
			Env: []k8s.EnvVar{
				{Name: "TOKEN"},
			},
		},
		Metadata: api.Map{
			"auth": api.Map{
				"type":  "bearer",
				"token": "$(TOKEN)",
				"headers": []any{
					"Authorization: Bearer $(TOKEN)",
					"X-Custom: value",
				},
			},
			"items": []any{
				api.Map{
					"name":  "item1",
					"value": "$(TOKEN)",
				},
				api.Map{
					"name":  "item2",
					"value": "static",
				},
			},
		},
	}

	// Inject environment variables
	injector := EnvInjector{}
	injector.Inject(extension)

	// Verify nested injection
	metadata := extension.Metadata.(map[string]any)
	auth := metadata["auth"].(map[string]any)
	g.Expect(auth["token"]).To(Equal("secret123"))
	headers := auth["headers"].([]any)
	g.Expect(headers[0]).To(Equal("Authorization: Bearer secret123"))

	items := metadata["items"].([]any)
	item1 := items[0].(map[string]any)
	g.Expect(item1["value"]).To(Equal("secret123"))
}

func TestEnvInjectorMissingVariable(t *testing.T) {
	g := NewGomegaWithT(t)

	// Do not set environment variable

	// Create extension with metadata containing env reference
	extension := &api.Extension{
		Name: "ext1",
		Container: k8s.Container{
			Env: []k8s.EnvVar{
				{Name: "MISSING_VAR"},
			},
		},
		Metadata: api.Map{
			"value": "$(MISSING_VAR)",
		},
	}

	// Inject environment variables
	injector := EnvInjector{}
	injector.Inject(extension)

	// Verify no injection occurred (variable reference remains)
	metadata := extension.Metadata.(map[string]any)
	g.Expect(metadata["value"]).To(Equal("$(MISSING_VAR)"))
}

func TestTaskActivityMultiline(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Add multiline activity
	adapter.Activity("First line\nSecond line\nThird line")
	g.Expect(len(updatedReport.Activity)).To(Equal(3))
	g.Expect(updatedReport.Activity[0]).To(Equal("First line"))
	g.Expect(updatedReport.Activity[1]).To(Equal("> Second line"))
	g.Expect(updatedReport.Activity[2]).To(Equal("> Third line"))
}

func TestTaskData(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	taskData := api.Map{
		"key1": "value1",
		"key2": 123,
	}
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Data = taskData
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get data
	data := adapter.Task.Data()
	g.Expect(data).To(Equal(taskData))
}

func TestEnvInjectorComplexTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create extension with various data types
	extension := &api.Extension{
		Name: "ext1",
		Container: k8s.Container{
			Env: []k8s.EnvVar{},
		},
		Metadata: api.Map{
			"string": "text",
			"number": 42,
			"float":  3.14,
			"bool":   true,
			"null":   nil,
			"array":  []any{1, 2, 3},
			"map": api.Map{
				"nested": "value",
			},
		},
	}

	// Inject (should not modify non-string values)
	injector := EnvInjector{}
	injector.Inject(extension)

	// Verify types preserved
	metadata := extension.Metadata.(map[string]any)
	g.Expect(metadata["string"]).To(Equal("text"))
	g.Expect(metadata["number"]).To(Equal(float64(42))) // JSON unmarshaling converts to float64
	g.Expect(metadata["float"]).To(Equal(3.14))
	g.Expect(metadata["bool"]).To(Equal(true))
	g.Expect(metadata["null"]).To(BeNil())
}

func TestTaskErrorf(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Errorf with formatting
	adapter.Errorf("Warning", "Item %d failed: %s", 5, "timeout")
	g.Expect(updatedReport).NotTo(BeNil())
	g.Expect(len(updatedReport.Errors)).To(Equal(1))
	g.Expect(updatedReport.Errors[0].Severity).To(Equal("Warning"))
	g.Expect(updatedReport.Errors[0].Description).To(Equal("Item 5 failed: timeout"))
}

func TestTaskBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 123
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 123
	adapter.Load()

	// Get bucket
	bucket := adapter.Task.Bucket()
	g.Expect(bucket).NotTo(BeNil())
}

func TestDataWithInvalidJSON(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				// Set data that doesn't match expected structure
				r.Data = api.Map{
					"field": "string_value",
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Try to unmarshal into incompatible type
	type TestData struct {
		Field int `json:"field"` // Expects int but data has string
	}
	var data TestData
	err := adapter.Task.DataWith(&data)
	// Should error because string can't convert to int
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("cannot unmarshal string"))
}

func TestAddonNotSpecified(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Addon = "" // No addon specified
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon - should fail
	addon, err := adapter.Task.Addon(false)
	g.Expect(err).NotTo(BeNil())
	g.Expect(addon).To(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("addon not specified"))
}

func TestPlatformNotSpecified(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Platform = nil
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get platform - should fail
	platform, err := adapter.Task.Platform()
	g.Expect(err).NotTo(BeNil())
	g.Expect(platform).To(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("platform not specified"))
}

func TestEnvInjectorWithPartialMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	// Set environment variable
	os.Setenv("_EXT_EXT1_VAR1", "value1")
	t.Cleanup(func() {
		os.Unsetenv("_EXT_EXT1_VAR1")
	})

	// Create extension with partial variable reference (should not inject)
	extension := &api.Extension{
		Name: "ext1",
		Container: k8s.Container{
			Env: []k8s.EnvVar{
				{Name: "VAR1"},
			},
		},
		Metadata: api.Map{
			"incomplete": "$(VAR",               // Incomplete pattern
			"wrong":      "$VAR1)",              // Wrong pattern
			"correct":    "$(VAR1)",             // Correct pattern
			"multiple":   "$(VAR1) and $(VAR1)", // Multiple references
		},
	}

	// Inject environment variables
	injector := EnvInjector{}
	injector.Inject(extension)

	// Verify only correct pattern is replaced
	metadata := extension.Metadata.(map[string]any)
	g.Expect(metadata["incomplete"]).To(Equal("$(VAR"))
	g.Expect(metadata["wrong"]).To(Equal("$VAR1)"))
	g.Expect(metadata["correct"]).To(Equal("value1"))
	g.Expect(metadata["multiple"]).To(Equal("value1 and value1"))
}

func TestTaskReportStatusPreserved(t *testing.T) {
	g := NewGomegaWithT(t)

	var lastReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
				// Make a copy to track state
				b, _ := json.Marshal(r)
				lastReport = &api.TaskReport{}
				_ = json.Unmarshal(b, lastReport)
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				// Make a copy to track state
				b, _ := json.Marshal(r)
				lastReport = &api.TaskReport{}
				_ = json.Unmarshal(b, lastReport)
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Start task
	adapter.Started()
	g.Expect(lastReport.Status).To(Equal(task.Running))

	// Add activity - status should remain Running
	adapter.Activity("test")
	g.Expect(lastReport.Status).To(Equal(task.Running))

	// Fail task
	adapter.Failed("error")
	g.Expect(lastReport.Status).To(Equal(task.Failed))

	// Additional activity after failure should not change status
	adapter.Activity("more activity")
	g.Expect(lastReport.Status).To(Equal(task.Failed))
}

func TestSucceededSetsCompleted(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Set total
	adapter.Total(100)

	// Call Succeeded - should set Completed = Total
	adapter.Succeeded()
	g.Expect(updatedReport.Status).To(Equal(task.Succeeded))
	g.Expect(updatedReport.Completed).To(Equal(100))
	g.Expect(updatedReport.Completed).To(Equal(updatedReport.Total))
}

// HTTP Failure Tests

func TestTaskLoadHTTPError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client that returns error
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			err = &RestError{Reason: "Internal Server Error"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1

	// Load should panic with error
	g.Expect(func() {
		adapter.Load()
	}).To(Panic())
}

func TestTaskApplicationHTTPError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch object.(type) {
			case *api.Task:
				r := object.(*api.Task)
				r.ID = 1
				r.Application = &api.Ref{ID: 100}
			case *api.Application:
				// Simulate HTTP error fetching application
				err = &RestError{Reason: "Service Unavailable"}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get application - should fail with HTTP error
	_, err := adapter.Task.Application()
	g.Expect(err).NotTo(BeNil())
}

func TestTaskPlatformHTTPError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch object.(type) {
			case *api.Task:
				r := object.(*api.Task)
				r.ID = 1
				r.Platform = &api.Ref{ID: 200}
			case *api.Platform:
				// Simulate HTTP error fetching platform
				err = &NotFound{}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get platform - should fail with HTTP error
	_, err := adapter.Task.Platform()
	g.Expect(err).NotTo(BeNil())
}

func TestTaskAddonHTTPError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch object.(type) {
			case *api.Task:
				r := object.(*api.Task)
				r.ID = 1
				r.Addon = "test-addon"
			case *api.Addon:
				// Simulate HTTP error fetching addon
				err = &RestError{Reason: "Gateway Timeout"}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon - should fail with HTTP error
	_, err := adapter.Task.Addon(false)
	g.Expect(err).NotTo(BeNil())
}

func TestStartedDeleteReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where delete fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			// Simulate delete error
			err = &RestError{Reason: "Forbidden"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Started should panic when delete fails
	g.Expect(func() {
		adapter.Started()
	}).To(Panic())
}

func TestStartedCreateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where create fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			// Simulate create error
			err = &RestError{Reason: "Bad Request"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Started should panic when create fails
	g.Expect(func() {
		adapter.Started()
	}).To(Panic())
}

func TestActivityUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	var createCalled bool

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			createCalled = true
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &RestError{Reason: "Conflict"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()
	g.Expect(createCalled).To(BeTrue())

	// Activity should panic when update fails
	g.Expect(func() {
		adapter.Activity("test")
	}).To(Panic())
}

func TestTotalUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &NotFound{}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Total should panic when update fails
	g.Expect(func() {
		adapter.Total(10)
	}).To(Panic())
}

func TestSucceededUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &RestError{Reason: "Internal Server Error"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Succeeded should panic when update fails
	g.Expect(func() {
		adapter.Succeeded()
	}).To(Panic())
}

func TestFailedUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &RestError{Reason: "Service Unavailable"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Failed should panic when update fails
	g.Expect(func() {
		adapter.Failed("test error")
	}).To(Panic())
}

// Adapter.Run() Tests

func TestRunSuccess(t *testing.T) {
	g := NewGomegaWithT(t)

	var addonCalled bool
	var finalStatus string

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				finalStatus = r.Status
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1

	// Run successful addon
	addon := func() (err error) {
		addonCalled = true
		return
	}

	// Note: We can't actually call Run() because it calls os.Exit(1) on error
	// and os.Exit() on success would also terminate the test.
	// Instead, we test the logic flow manually.
	adapter.Load()
	adapter.Started()
	err := addon()
	g.Expect(err).To(BeNil())
	g.Expect(addonCalled).To(BeTrue())

	// Simulate the success logic from Run()
	adapter.Succeeded()
	g.Expect(finalStatus).To(Equal(task.Succeeded))
}

func TestRunAddonReturnsError(t *testing.T) {
	g := NewGomegaWithT(t)

	var addonCalled bool
	var finalStatus string
	var finalErrors []api.TaskError

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				finalStatus = r.Status
				finalErrors = r.Errors
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1

	// Run addon that returns error
	addon := func() (err error) {
		addonCalled = true
		err = &RestError{Reason: "addon failed"}
		return
	}

	// Simulate Run() logic
	adapter.Load()
	adapter.Started()
	err := addon()
	g.Expect(err).NotTo(BeNil())
	g.Expect(addonCalled).To(BeTrue())

	// Simulate error handling from Run()
	adapter.Failed(err.Error())
	g.Expect(finalStatus).To(Equal(task.Failed))
	g.Expect(len(finalErrors)).To(Equal(1))
	g.Expect(finalErrors[0].Description).To(ContainSubstring("addon failed"))
}

func TestRunAddonPanics(t *testing.T) {
	g := NewGomegaWithT(t)

	var addonCalled bool

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1

	// Run addon that panics
	addon := func() (err error) {
		addonCalled = true
		panic("something went wrong")
	}

	adapter.Load()
	adapter.Started()

	// Capture panic
	var panicValue any
	func() {
		defer func() {
			panicValue = recover()
		}()
		_ = addon()
	}()

	g.Expect(addonCalled).To(BeTrue())
	g.Expect(panicValue).To(Equal("something went wrong"))
}

func TestRunAddonPanicsWithError(t *testing.T) {
	g := NewGomegaWithT(t)

	var addonCalled bool
	var finalStatus string

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				finalStatus = r.Status
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1

	// Run addon that panics with an error
	addon := func() (err error) {
		addonCalled = true
		panic(&RestError{Reason: "panic error"})
	}

	adapter.Load()
	adapter.Started()

	// Capture panic and handle it like Run() does
	var recoveredErr error
	func() {
		defer func() {
			r := recover()
			if r != nil {
				if pErr, ok := r.(error); ok {
					recoveredErr = pErr
				}
			}
		}()
		_ = addon()
	}()

	g.Expect(addonCalled).To(BeTrue())
	g.Expect(recoveredErr).NotTo(BeNil())
	g.Expect(recoveredErr.Error()).To(ContainSubstring("panic error"))

	// Simulate Failed() call from Run()
	adapter.Failed(recoveredErr.Error())
	g.Expect(finalStatus).To(Equal(task.Failed))
}

func TestRunStatusAlreadySet(t *testing.T) {
	g := NewGomegaWithT(t)

	var addonCalled bool
	var succeededCalled bool

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				if r.Status == task.Succeeded {
					succeededCalled = true
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1

	// Run addon that explicitly sets status
	addon := func() (err error) {
		addonCalled = true
		adapter.Failed("custom failure")
		return
	}

	adapter.Load()
	adapter.Started()
	_ = addon()

	g.Expect(addonCalled).To(BeTrue())

	// Simulate Run() logic - should NOT call Succeeded() when status already set
	switch adapter.report.Status {
	case task.Failed, task.Succeeded:
		// Don't call Succeeded()
	default:
		adapter.Succeeded()
	}

	// Succeeded should not have been called since status was already Failed
	g.Expect(succeededCalled).To(BeFalse())
	g.Expect(adapter.report.Status).To(Equal(task.Failed))
}

// Additional Coverage Tests

func TestAdapterClient(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	adapter.Use(richClient)

	// Get client
	client := adapter.Client()
	g.Expect(client).NotTo(BeNil())
	g.Expect(client).To(Equal(richClient))
}

func TestErrorMultiple(t *testing.T) {
	g := NewGomegaWithT(t)

	var updatedReport *api.TaskReport

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				updatedReport = r
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Report multiple errors in single call
	adapter.Error(
		api.TaskError{
			Severity:    "Warning",
			Description: "First warning",
		},
		api.TaskError{
			Severity:    "Error",
			Description: "First error",
		},
		api.TaskError{
			Severity:    "Info",
			Description: "Information message",
		},
	)

	g.Expect(updatedReport).NotTo(BeNil())
	g.Expect(len(updatedReport.Errors)).To(Equal(3))
	g.Expect(updatedReport.Errors[0].Severity).To(Equal("Warning"))
	g.Expect(updatedReport.Errors[0].Description).To(Equal("First warning"))
	g.Expect(updatedReport.Errors[1].Severity).To(Equal("Error"))
	g.Expect(updatedReport.Errors[1].Description).To(Equal("First error"))
	g.Expect(updatedReport.Errors[2].Severity).To(Equal("Info"))
	g.Expect(updatedReport.Errors[2].Description).To(Equal("Information message"))
}

func TestIncrementUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &RestError{Reason: "Update failed"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Increment should panic when update fails
	g.Expect(func() {
		adapter.Increment()
	}).To(Panic())
}

func TestCompletedUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &RestError{Reason: "Update failed"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Completed should panic when update fails
	g.Expect(func() {
		adapter.Completed(5)
	}).To(Panic())
}

func TestResultUpdateReportError(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client where update fails
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
			}
			return
		},
		DoDelete: func(path string, params ...client.Param) (err error) {
			return
		},
		DoPost: func(path string, object any) (err error) {
			if r, ok := object.(*api.TaskReport); ok {
				r.ID = 100
			}
			return
		},
		DoPut: func(path string, object any, params ...client.Param) (err error) {
			// Simulate update error
			err = &RestError{Reason: "Update failed"}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()
	adapter.Started()

	// Result should panic when update fails
	g.Expect(func() {
		adapter.Result(api.Map{"key": "value"})
	}).To(Panic())
}

func TestAddonExtensionFiltering(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Addon = "test-addon"
				// Only request ext2 and ext4
				r.Extensions = []string{"ext2", "ext4"}
			case *api.Addon:
				r.Name = "test-addon"
				// Addon has ext1, ext2, ext3, ext4
				r.Extensions = []api.Extension{
					{Name: "ext1", Container: k8s.Container{}},
					{Name: "ext2", Container: k8s.Container{}},
					{Name: "ext3", Container: k8s.Container{}},
					{Name: "ext4", Container: k8s.Container{}},
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon - should only include ext2 and ext4
	addon, err := adapter.Task.Addon(false)
	g.Expect(err).To(BeNil())
	g.Expect(addon).NotTo(BeNil())
	g.Expect(len(addon.Extensions)).To(Equal(2))
	g.Expect(addon.Extensions[0].Name).To(Equal("ext2"))
	g.Expect(addon.Extensions[1].Name).To(Equal("ext4"))
}

func TestAddonNoExtensionsRequested(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Addon = "test-addon"
				// No extensions requested
				r.Extensions = []string{}
			case *api.Addon:
				r.Name = "test-addon"
				r.Extensions = []api.Extension{
					{Name: "ext1", Container: k8s.Container{}},
					{Name: "ext2", Container: k8s.Container{}},
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon - should have no extensions
	addon, err := adapter.Task.Addon(false)
	g.Expect(err).To(BeNil())
	g.Expect(addon).NotTo(BeNil())
	g.Expect(len(addon.Extensions)).To(Equal(0))
}

func TestAddonAllExtensionsFiltered(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create adapter with stub client
	adapter := &Adapter{}
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.ID = 1
				r.Addon = "test-addon"
				// Request extensions that don't exist
				r.Extensions = []string{"ext99", "ext100"}
			case *api.Addon:
				r.Name = "test-addon"
				r.Extensions = []api.Extension{
					{Name: "ext1", Container: k8s.Container{}},
					{Name: "ext2", Container: k8s.Container{}},
				}
			}
			return
		},
	})
	adapter.Use(richClient)

	Settings.Task = 1
	adapter.Load()

	// Get addon - should have no extensions (all filtered)
	addon, err := adapter.Task.Addon(false)
	g.Expect(err).To(BeNil())
	g.Expect(addon).NotTo(BeNil())
	g.Expect(len(addon.Extensions)).To(Equal(0))
}

// SCM Resource Injection Tests
// Note: SCM.New() and proxyMap() cannot be easily unit tested because:
// 1. SCM.New() calls r.Validate() which requires actual git/svn binaries
// 2. proxyMap() uses the global Addon variable which cannot be mocked
// These functions are better suited for integration tests.

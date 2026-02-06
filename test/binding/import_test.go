package binding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestImport(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a temporary test CSV file to upload
	testDir, err := os.MkdirTemp("", "test-import-*")
	g.Expect(err).To(BeNil())

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
	})

	csvFile := filepath.Join(testDir, "test_import.csv")
	csvLine1 := []string{
		"Record Type 1",
		"Application Name",
		"Description",
		"Comments",
		"Business Service",
		"Dependency",
		"Dependency Direction",
		"Binary Group",
		"Binary Artifact",
		"Binary Version",
		"Binary Packaging",
		"Repository Type",
		"Repository URL",
		"Repository Branch",
		"Repository Path",
		"Owner",
		"Contributors",
		"Tag Category 1",
		"Tag 1",
		"Tag Category 2",
		"Tag 2",
		"Tag Category 3",
		"Tag 3",
		"Tag Category 4",
		"Tag 4",
		"Tag Category 5",
		"Tag 5",
		"Tag Category 6",
		"Tag 6",
		"Tag Category 7",
		"Tag 7",
		"Tag Category 8",
		"Tag 8",
		"Tag Category 9",
		"Tag 9",
		"Tag Category 10",
		"Tag 10",
		"Tag Category 11",
		"Tag 11",
		"Tag Category 12",
		"Tag 12",
		"Tag Category 13",
		"Tag 13",
		"Tag Category 14",
		"Tag 14",
		"Tag Category 15",
		"Tag 15",
		"Tag Category 16",
		"Tag 16",
		"Tag Category 17",
		"Tag 17",
		"Tag Category 18",
		"Tag 18",
		"Tag Category 19",
		"Tag 19",
		"Tag Category 20",
		"Tag 20",
	}
	csvLine2 := []string{
		"1",
		"TestApp",
		"Test application",
		"",
		"TestService",
		"",
		"",
		"com.test",
		"testapp",
		"1.0.0",
		"jar",
		"git",
		"https://git.example.com/testapp.git",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	}
	csvContent := strings.Join(csvLine1, ",")
	csvContent += "\n"
	csvContent += strings.Join(csvLine2, ",")
	csvContent += "\n"

	err = os.WriteFile(csvFile, []byte(csvContent), 0644)
	g.Expect(err).To(BeNil())

	// CREATE: Upload the CSV file
	uploaded, err := client.Import.Upload(csvFile)
	g.Expect(err).To(BeNil())
	g.Expect(uploaded).NotTo(BeNil())
	g.Expect(uploaded.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Import.Summary().Delete(uploaded.ID)
	})

	// Wait for the import processing to complete
	time.Sleep(2 * time.Second)

	// Clean up resources created by the import
	t.Cleanup(func() {
		// Delete applications and their associated resources
		apps, _ := client.Application.List()
		for _, app := range apps {
			if app.Name == "TestApp" {
				// Delete owner stakeholder if exists
				if app.Owner != nil {
					_ = client.Stakeholder.Delete(app.Owner.ID)
				}
				// Delete contributor stakeholders if exist
				for _, contributor := range app.Contributors {
					_ = client.Stakeholder.Delete(contributor.ID)
				}
				// Delete the application
				_ = client.Application.Delete(app.ID)
			}
		}

		// Delete dependencies created by the import
		deps, _ := client.Dependency.List()
		for _, dep := range deps {
			// Delete dependencies where From or To is TestApp
			if dep.From.Name == "TestApp" || dep.To.Name == "TestApp" {
				_ = client.Dependency.Delete(dep.ID)
			}
		}

		// Delete the business service created by the import
		services, _ := client.BusinessService.List()
		for _, svc := range services {
			if svc.Name == "TestService" {
				_ = client.BusinessService.Delete(svc.ID)
				break
			}
		}
	})

	// GET: List import summaries
	list, err := client.Import.Summary().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list) >= 1).To(BeTrue())

	// Find our uploaded summary in the list
	var found *api.ImportSummary
	for i := range list {
		if list[i].ID == uploaded.ID {
			found = &list[i]
			break
		}
	}

	ignoredPaths := []string{
		"ImportTime",
		"ImportStatus",
		"ValidCount",
		"InvalidCount",
	}

	g.Expect(found).NotTo(BeNil())
	eq, report := cmp.Eq(uploaded, found, ignoredPaths...)
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the import summary and verify it matches
	retrieved, err := client.Import.Summary().Get(uploaded.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(uploaded, retrieved, ignoredPaths...)
	g.Expect(eq).To(BeTrue(), report)

	// GET: Download the CSV file
	downloadFile := filepath.Join(testDir, "downloaded.csv")
	err = client.Import.Summary().Download(downloadFile)
	g.Expect(err).To(BeNil())

	// Verify the downloaded file exists and has content
	g.Expect(assert.EqualFileContent(downloadFile, csvFile)).To(BeTrue())

	// GET: List imports
	imports, err := client.Import.Summary().List()
	g.Expect(err).To(BeNil())
	g.Expect(len(imports) >= 1).To(BeTrue())

	// DELETE: Remove the import summary
	err = client.Import.Summary().Delete(uploaded.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail with NotFound
	_, err = client.Import.Summary().Get(uploaded.ID)
	g.Expect(err).NotTo(BeNil())
}

package reaper

import (
	"os"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/task"
	"github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TestBucketReaper_OrphanDetection tests bucket orphan detection and deletion.
func TestBucketReaper_OrphanDetection(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create test buckets
	bucket1 := &model.Bucket{Path: "/tmp/bucket1"}
	bucket2 := &model.Bucket{Path: "/tmp/bucket2"}
	bucket3 := &model.Bucket{Path: "/tmp/bucket3"}

	err = db.Create(bucket1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(bucket2).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(bucket3).Error
	g.Expect(err).To(gomega.BeNil())

	// Create application referencing bucket1
	app := &model.Application{Name: "TestApp"}
	app.BucketID = &bucket1.ID
	err = db.Create(app).Error
	g.Expect(err).To(gomega.BeNil())

	// Create task referencing bucket2
	task := &model.Task{Name: "TestTask"}
	task.BucketID = &bucket2.ID
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should mark bucket3 as orphan
	reaper := &BucketReaper{DB: db}
	reaper.Run()

	// Verify bucket1 and bucket2 still have no expiration (referenced)
	var b1 model.Bucket
	err = db.First(&b1, bucket1.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b1.Expiration).To(gomega.BeNil())

	var b2 model.Bucket
	err = db.First(&b2, bucket2.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b2.Expiration).To(gomega.BeNil())

	// Verify bucket3 has expiration set (orphan)
	var b3 model.Bucket
	err = db.First(&b3, bucket3.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b3.Expiration).NotTo(gomega.BeNil())
}

// TestBucketReaper_ExpirationDeletion tests bucket deletion after expiration.
func TestBucketReaper_ExpirationDeletion(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create temp directory for bucket
	bucketPath := t.TempDir() + "/expired-bucket"
	err = os.Mkdir(bucketPath, 0755)
	g.Expect(err).To(gomega.BeNil())

	// Create orphan bucket with expired TTL
	expiredTime := time.Now().Add(-1 * time.Hour)
	bucket := &model.Bucket{
		Path:       bucketPath,
		Expiration: &expiredTime,
	}
	err = db.Create(bucket).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should delete expired bucket
	reaper := &BucketReaper{DB: db}
	reaper.Run()

	// Verify bucket was deleted from database
	var b model.Bucket
	err = db.First(&b, bucket.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))

	// Verify bucket was deleted from filesystem
	_, err = os.Stat(bucketPath)
	g.Expect(os.IsNotExist(err)).To(gomega.BeTrue())
}

// TestBucketReaper_ReferenceRestoresExpiration tests that references clear expiration.
func TestBucketReaper_ReferenceRestoresExpiration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create temp directory for bucket
	bucketPath := t.TempDir() + "/referenced-bucket"
	err = os.Mkdir(bucketPath, 0755)
	g.Expect(err).To(gomega.BeNil())

	// Create bucket with expiration
	futureTime := time.Now().Add(1 * time.Hour)
	bucket := &model.Bucket{
		Path:       bucketPath,
		Expiration: &futureTime,
	}
	err = db.Create(bucket).Error
	g.Expect(err).To(gomega.BeNil())

	// Create application referencing the bucket
	app := &model.Application{Name: "TestApp"}
	app.BucketID = &bucket.ID
	err = db.Create(app).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should clear expiration
	reaper := &BucketReaper{DB: db}
	reaper.Run()

	// Verify expiration was cleared
	var b model.Bucket
	err = db.First(&b, bucket.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b.Expiration).To(gomega.BeNil())

	// Verify bucket still exists on filesystem
	_, err = os.Stat(bucketPath)
	g.Expect(err).To(gomega.BeNil())
}

// TestFileReaper_OrphanDetection tests file orphan detection and deletion.
func TestFileReaper_OrphanDetection(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create file records (BeforeCreate will assign paths)
	file1 := &model.File{}
	file2 := &model.File{}
	file3 := &model.File{}

	err = db.Create(file1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(file2).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(file3).Error
	g.Expect(err).To(gomega.BeNil())

	// Create actual files at the assigned paths
	err = os.WriteFile(file1.Path, []byte("test1"), 0644)
	g.Expect(err).To(gomega.BeNil())
	t.Cleanup(func() { os.Remove(file1.Path) })
	err = os.WriteFile(file2.Path, []byte("test2"), 0644)
	g.Expect(err).To(gomega.BeNil())
	t.Cleanup(func() { os.Remove(file2.Path) })
	err = os.WriteFile(file3.Path, []byte("test3"), 0644)
	g.Expect(err).To(gomega.BeNil())
	t.Cleanup(func() { os.Remove(file3.Path) })

	// Create task referencing file1
	task := &model.Task{Name: "TestTask"}
	task.Attached = []model.Attachment{{ID: file1.ID}}
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Create task report referencing file2
	report := &model.TaskReport{}
	report.Attached = []model.Attachment{{ID: file2.ID}}
	// Need to create a task first since TaskReport requires a TaskID
	taskForReport := &model.Task{Name: "TaskForReport"}
	err = db.Create(taskForReport).Error
	g.Expect(err).To(gomega.BeNil())
	report.TaskID = taskForReport.ID
	err = db.Create(report).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should mark file3 as orphan
	reaper := &FileReaper{DB: db}
	reaper.Run()

	// Verify file1 and file2 have no expiration (referenced)
	var f1 model.File
	err = db.First(&f1, file1.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f1.Expiration).To(gomega.BeNil())

	var f2 model.File
	err = db.First(&f2, file2.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f2.Expiration).To(gomega.BeNil())

	// Verify file3 has expiration set (orphan)
	var f3 model.File
	err = db.First(&f3, file3.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f3.Expiration).NotTo(gomega.BeNil())
}

// TestFileReaper_ExpirationDeletion tests file deletion after expiration.
func TestFileReaper_ExpirationDeletion(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create orphan file with expired TTL
	expiredTime := time.Now().Add(-1 * time.Hour)
	file := &model.File{
		Expiration: &expiredTime,
	}
	err = db.Create(file).Error
	g.Expect(err).To(gomega.BeNil())

	// Create the actual file at the path assigned by BeforeCreate
	err = os.WriteFile(file.Path, []byte("test"), 0644)
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should delete expired file
	reaper := &FileReaper{DB: db}
	reaper.Run()

	// Verify file was deleted from database
	var f model.File
	err = db.First(&f, file.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))

	// Verify file was deleted from filesystem
	_, err = os.Stat(file.Path)
	g.Expect(os.IsNotExist(err)).To(gomega.BeTrue())
}

// TestFileReaper_ReferenceRestoresExpiration tests that references clear expiration.
func TestFileReaper_ReferenceRestoresExpiration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create file with expiration
	futureTime := time.Now().Add(1 * time.Hour)
	file := &model.File{
		Expiration: &futureTime,
	}
	err = db.Create(file).Error
	g.Expect(err).To(gomega.BeNil())

	// Create actual file at the path assigned by BeforeCreate
	err = os.WriteFile(file.Path, []byte("test"), 0644)
	g.Expect(err).To(gomega.BeNil())
	t.Cleanup(func() { os.Remove(file.Path) })

	// Create task referencing the file
	task := &model.Task{Name: "TestTask"}
	task.Attached = []model.Attachment{{ID: file.ID}}
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should clear expiration
	reaper := &FileReaper{DB: db}
	reaper.Run()

	// Verify expiration was cleared
	var f model.File
	err = db.First(&f, file.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f.Expiration).To(gomega.BeNil())
}

// TestGroupReaper_CreatedState tests group reaping in Created state.
func TestGroupReaper_CreatedState(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Save original settings
	reaperCreated := Settings.Hub.Task.Reaper.Created
	t.Cleanup(func() {
		Settings.Hub.Task.Reaper.Created = reaperCreated
	})

	// Create old group in Created state
	oldTime := time.Now().Add(-2 * Settings.Hub.Task.Reaper.Created)
	group := &model.TaskGroup{
		Name:  "OldGroup",
		State: task.Created,
	}
	group.CreateTime = oldTime
	err = db.Create(group).Error
	g.Expect(err).To(gomega.BeNil())

	// Create recent group in Created state
	recentGroup := &model.TaskGroup{
		Name:  "RecentGroup",
		State: task.Created,
	}
	recentGroup.CreateTime = time.Now()
	err = db.Create(recentGroup).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should delete old group
	reaper := &GroupReaper{DB: db}
	reaper.Run()

	// Verify old group was deleted
	var g1 model.TaskGroup
	err = db.First(&g1, group.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))

	// Verify recent group still exists
	var g2 model.TaskGroup
	err = db.First(&g2, recentGroup.ID).Error
	g.Expect(err).To(gomega.BeNil())
}

// TestGroupReaper_ReadyStateWithNoTasks tests group deletion when all tasks are reaped.
func TestGroupReaper_ReadyStateWithNoTasks(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old group in Ready state with no tasks
	oldTime := time.Now().Add(-2 * time.Hour)
	group := &model.TaskGroup{
		Name:  "EmptyGroup",
		State: task.Ready,
	}
	group.CreateTime = oldTime
	err = db.Create(group).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should delete group after 1 hour
	reaper := &GroupReaper{DB: db}
	reaper.Run()

	// Verify group was deleted
	var tg model.TaskGroup
	err = db.First(&tg, group.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestTaskReaper_CreatedState tests task reaping in Created state.
func TestTaskReaper_CreatedState(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Created state with custom TTL
	oldTime := time.Now().Add(-2 * time.Hour)
	taskWithTTL := &model.Task{
		Name:  "TaskWithTTL",
		State: task.Created,
	}
	taskWithTTL.CreateTime = oldTime
	taskWithTTL.TTL.Created = 60 // 60 minutes
	err = db.Create(taskWithTTL).Error
	g.Expect(err).To(gomega.BeNil())

	// Create recent task in Created state
	recentTask := &model.Task{
		Name:  "RecentTask",
		State: task.Created,
	}
	recentTask.CreateTime = time.Now()
	err = db.Create(recentTask).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should delete old task
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify old task was deleted
	var t1 model.Task
	err = db.First(&t1, taskWithTTL.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))

	// Verify recent task still exists
	var t2 model.Task
	err = db.First(&t2, recentTask.ID).Error
	g.Expect(err).To(gomega.BeNil())
}

// TestTaskReaper_SucceededState tests task resource release in Succeeded state.
func TestTaskReaper_SucceededState(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Save original settings
	reaperSucceeded := Settings.Hub.Task.Reaper.Succeeded
	t.Cleanup(func() {
		Settings.Hub.Task.Reaper.Succeeded = reaperSucceeded
	})

	// Create bucket for task
	bucket := &model.Bucket{Path: "/tmp/task-bucket"}
	err = db.Create(bucket).Error
	g.Expect(err).To(gomega.BeNil())

	// Create succeeded task with bucket
	terminatedTime := time.Now().Add(-2 * Settings.Hub.Task.Reaper.Succeeded)
	task := &model.Task{
		Name:       "SucceededTask",
		State:      task.Succeeded,
		Terminated: &terminatedTime,
		Reaped:     false,
	}
	task.CreateTime = time.Now().Add(-3 * Settings.Hub.Task.Reaper.Succeeded)
	task.BucketID = &bucket.ID
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should release bucket
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task bucket was released and marked as reaped
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(t1.BucketID).To(gomega.BeNil())
	g.Expect(t1.Reaped).To(gomega.BeTrue())
}

// TestTaskReaper_FailedStateWithTTL tests task deletion in Failed state with TTL.
func TestTaskReaper_FailedStateWithTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create failed task with expired TTL
	terminatedTime := time.Now().Add(-2 * time.Hour)
	task := &model.Task{
		Name:       "FailedTask",
		State:      task.Failed,
		Terminated: &terminatedTime,
	}
	task.CreateTime = time.Now().Add(-3 * time.Hour)
	task.TTL.Failed = 60 // 60 minutes
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper - should delete task
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was deleted
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestRefFinder_FindsReferences tests the RefFinder for bucket and file references.
func TestRefFinder_FindsReferences(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create buckets
	bucket1 := &model.Bucket{Path: "/tmp/bucket1"}
	bucket2 := &model.Bucket{Path: "/tmp/bucket2"}
	err = db.Create(bucket1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(bucket2).Error
	g.Expect(err).To(gomega.BeNil())

	// Create tasks referencing buckets
	task1 := &model.Task{Name: "Task1"}
	task1.BucketID = &bucket1.ID
	task2 := &model.Task{Name: "Task2"}
	task2.BucketID = &bucket1.ID
	task3 := &model.Task{Name: "Task3"}
	task3.BucketID = &bucket2.ID
	err = db.Create(task1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(task2).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(task3).Error
	g.Expect(err).To(gomega.BeNil())

	// Find bucket references
	ids := make(map[uint]byte)
	finder := RefFinder{DB: db}
	err = finder.Find(&model.Task{}, "bucket", ids)
	g.Expect(err).To(gomega.BeNil())

	// Verify both buckets were found
	g.Expect(ids).To(gomega.HaveKey(bucket1.ID))
	g.Expect(ids).To(gomega.HaveKey(bucket2.ID))
	g.Expect(len(ids)).To(gomega.Equal(2))
}

// TestManager_Iterate tests the Manager Iterate method.
func TestManager_Iterate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create orphan bucket
	bucket := &model.Bucket{Path: "/tmp/test-bucket"}
	err = db.Create(bucket).Error
	g.Expect(err).To(gomega.BeNil())

	// Create manager and run iterate
	manager := &Manager{DB: db}
	manager.Iterate()

	// Verify bucket was marked as orphan
	var b model.Bucket
	err = db.First(&b, bucket.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b.Expiration).NotTo(gomega.BeNil())
}

// TestTaskReaper_ReadyStateWithoutTTL tests that tasks without TTL are NOT reaped.
func TestTaskReaper_ReadyStateWithoutTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Ready state without TTL
	oldTime := time.Now().Add(-48 * time.Hour)
	task := &model.Task{
		Name:  "ReadyTaskNoTTL",
		State: task.Ready,
	}
	task.CreateTime = oldTime
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was NOT reaped (intentional design)
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(t1.ID).To(gomega.Equal(task.ID))
}

// TestTaskReaper_ReadyStateWithTTL tests that tasks with TTL ARE reaped.
func TestTaskReaper_ReadyStateWithTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Ready state with expired TTL
	oldTime := time.Now().Add(-2 * time.Hour)
	task := &model.Task{
		Name:  "ReadyTaskWithTTL",
		State: task.Ready,
	}
	task.CreateTime = oldTime
	task.TTL.Pending = 60 // 60 minutes
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was deleted
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestTaskReaper_PendingStateWithTTL tests pending state reaping with TTL.
func TestTaskReaper_PendingStateWithTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Pending state with expired TTL
	oldTime := time.Now().Add(-3 * time.Hour)
	task := &model.Task{
		Name:  "PendingTaskWithTTL",
		State: task.Pending,
	}
	task.CreateTime = oldTime
	task.TTL.Pending = 120 // 120 minutes
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was deleted
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestTaskReaper_RunningStateWithoutTTL tests that running tasks without TTL are NOT reaped.
func TestTaskReaper_RunningStateWithoutTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Running state without TTL
	oldTime := time.Now().Add(-48 * time.Hour)
	startedTime := time.Now().Add(-47 * time.Hour)
	task := &model.Task{
		Name:    "RunningTaskNoTTL",
		State:   task.Running,
		Started: &startedTime,
	}
	task.CreateTime = oldTime
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was NOT reaped (intentional design)
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(t1.ID).To(gomega.Equal(task.ID))
}

// TestTaskReaper_RunningStateWithTTL tests that running tasks with TTL ARE reaped.
func TestTaskReaper_RunningStateWithTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Running state with expired TTL
	oldTime := time.Now().Add(-5 * time.Hour)
	startedTime := time.Now().Add(-4 * time.Hour)
	task := &model.Task{
		Name:    "RunningTaskWithTTL",
		State:   task.Running,
		Started: &startedTime,
	}
	task.CreateTime = oldTime
	task.TTL.Running = 180 // 180 minutes (3 hours)
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was deleted
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestTaskReaper_RunningStateWithStarted tests the bug fix for Started timestamp.
func TestTaskReaper_RunningStateWithStarted(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create task with Started set (after the bug fix)
	oldTime := time.Now().Add(-5 * time.Hour)
	startedTime := time.Now().Add(-4 * time.Hour)
	task := &model.Task{
		Name:    "RunningWithStarted",
		State:   task.Running,
		Started: &startedTime,
	}
	task.CreateTime = oldTime
	task.TTL.Running = 180 // 180 minutes
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was deleted based on Started time (4 hours ago > 3 hour TTL)
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestTaskReaper_CanceledState tests reaping of canceled tasks.
func TestTaskReaper_CanceledState(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create canceled task without TTL (should not be reaped)
	oldTime := time.Now().Add(-48 * time.Hour)
	taskNoTTL := &model.Task{
		Name:  "CanceledTaskNoTTL",
		State: task.Canceled,
	}
	taskNoTTL.CreateTime = oldTime
	err = db.Create(taskNoTTL).Error
	g.Expect(err).To(gomega.BeNil())

	// Create canceled task with TTL (currently no logic handles Canceled state)
	taskWithTTL := &model.Task{
		Name:  "CanceledTaskWithTTL",
		State: task.Canceled,
	}
	taskWithTTL.CreateTime = oldTime
	taskWithTTL.TTL.Failed = 60 // 60 minutes
	err = db.Create(taskWithTTL).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify both tasks still exist (Canceled state not handled in reaper)
	var t1 model.Task
	err = db.First(&t1, taskNoTTL.ID).Error
	g.Expect(err).To(gomega.BeNil())

	var t2 model.Task
	err = db.First(&t2, taskWithTTL.ID).Error
	g.Expect(err).To(gomega.BeNil())
}

// TestTaskReaper_PipelineProtection tests that pipeline tasks are NOT reaped.
func TestTaskReaper_PipelineProtection(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create a task group (pipeline) - Mode must be "Pipeline" (capital P)
	group := &model.TaskGroup{
		Name:  "TestPipeline",
		Mode:  "Pipeline",
		State: task.Ready,
		Kind:  "analyzer",
	}
	err = db.Create(group).Error
	g.Expect(err).To(gomega.BeNil())

	// Create old task in pipeline (should be protected)
	oldTime := time.Now().Add(-100 * time.Hour)
	pipelineTask := &model.Task{
		Name:        "PipelineTask",
		State:       task.Created,
		TaskGroupID: &group.ID,
	}
	pipelineTask.CreateTime = oldTime
	pipelineTask.TTL.Created = 60 // 60 minutes (expired)
	err = db.Create(pipelineTask).Error
	g.Expect(err).To(gomega.BeNil())

	// Create old non-pipeline task (should be reaped)
	nonPipelineTask := &model.Task{
		Name:  "NonPipelineTask",
		State: task.Created,
	}
	nonPipelineTask.CreateTime = oldTime
	nonPipelineTask.TTL.Created = 60 // 60 minutes (expired)
	err = db.Create(nonPipelineTask).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify pipeline task was NOT deleted (protected)
	var t1 model.Task
	err = db.First(&t1, pipelineTask.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(t1.ID).To(gomega.Equal(pipelineTask.ID))

	// Verify non-pipeline task WAS deleted
	var t2 model.Task
	err = db.First(&t2, nonPipelineTask.ID).Error
	g.Expect(err).NotTo(gomega.BeNil())
	g.Expect(err).To(gomega.Equal(gorm.ErrRecordNotFound))
}

// TestFileReaper_AllReferenceTypes tests file references from Rule, Target, AnalysisProfile.
func TestFileReaper_AllReferenceTypes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create files
	file1 := &model.File{}
	file2 := &model.File{}
	file3 := &model.File{}
	file4 := &model.File{}

	err = db.Create(file1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(file2).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(file3).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(file4).Error
	g.Expect(err).To(gomega.BeNil())

	// Create a RuleSet first (required for Rule)
	ruleSet := &model.RuleSet{Name: "TestRuleSet"}
	err = db.Create(ruleSet).Error
	g.Expect(err).To(gomega.BeNil())

	// Create Rule referencing file1
	rule := &model.Rule{
		Name:      "TestRule",
		RuleSetID: ruleSet.ID,
		FileID:    &file1.ID,
	}
	err = db.Create(rule).Error
	g.Expect(err).To(gomega.BeNil())

	// Create Target referencing file2 (ImageID field)
	target := &model.Target{
		Name:    "TestTarget",
		ImageID: file2.ID,
	}
	err = db.Create(target).Error
	g.Expect(err).To(gomega.BeNil())

	// Create AnalysisProfile referencing file3 (Files array field)
	profile := &model.AnalysisProfile{
		Name:  "TestProfile",
		Files: []model.Ref{{ID: file3.ID}},
	}
	err = db.Create(profile).Error
	g.Expect(err).To(gomega.BeNil())

	// file4 has no references (orphan)

	// Run reaper
	reaper := &FileReaper{DB: db}
	reaper.Run()

	// Verify file1, file2, file3 have no expiration (referenced)
	var f1 model.File
	err = db.First(&f1, file1.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f1.Expiration).To(gomega.BeNil())

	var f2 model.File
	err = db.First(&f2, file2.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f2.Expiration).To(gomega.BeNil())

	var f3 model.File
	err = db.First(&f3, file3.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f3.Expiration).To(gomega.BeNil())

	// Verify file4 has expiration set (orphan)
	var f4 model.File
	err = db.First(&f4, file4.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(f4.Expiration).NotTo(gomega.BeNil())
}

// TestBucketReaper_TaskGroupReference tests bucket referenced by TaskGroup.
func TestBucketReaper_TaskGroupReference(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create buckets
	bucket1 := &model.Bucket{Path: "/tmp/bucket1"}
	bucket2 := &model.Bucket{Path: "/tmp/bucket2"}

	err = db.Create(bucket1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(bucket2).Error
	g.Expect(err).To(gomega.BeNil())

	// Create TaskGroup referencing bucket1
	group := &model.TaskGroup{
		Name: "TestGroup",
	}
	group.BucketID = &bucket1.ID
	err = db.Create(group).Error
	g.Expect(err).To(gomega.BeNil())

	// bucket2 has no references (orphan)

	// Run reaper
	reaper := &BucketReaper{DB: db}
	reaper.Run()

	// Verify bucket1 has no expiration (referenced by TaskGroup)
	var b1 model.Bucket
	err = db.First(&b1, bucket1.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b1.Expiration).To(gomega.BeNil())

	// Verify bucket2 has expiration set (orphan)
	var b2 model.Bucket
	err = db.First(&b2, bucket2.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(b2.Expiration).NotTo(gomega.BeNil())
}

// TestEdgeCases_TTLZero tests behavior when TTL is set to zero.
func TestEdgeCases_TTLZero(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Save original settings
	reaperFailed := Settings.Hub.Task.Reaper.Failed
	t.Cleanup(func() {
		Settings.Hub.Task.Reaper.Failed = reaperFailed
	})

	// Create task with TTL = 0 (falls back to default settings)
	task := &model.Task{
		Name:  "TaskWithTTLZero",
		State: task.Failed,
	}
	task.TTL.Failed = 0
	terminatedTime := time.Now().Add(-48 * time.Hour)
	task.Terminated = &terminatedTime
	task.CreateTime = time.Now().Add(-49 * time.Hour)
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// With TTL=0, it falls through to settings-based reaping
	// Settings.Hub.Task.Reaper.Failed = 30 days (default)
	// 48 hours < 30 days, so should NOT be reaped yet
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(t1.Reaped).To(gomega.BeFalse())
}

// TestTaskReaper_CreatedStateWithoutTTL tests Created state release behavior.
func TestTaskReaper_CreatedStateWithoutTTL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Save original settings
	reaperCreated := Settings.Hub.Task.Reaper.Created
	t.Cleanup(func() {
		Settings.Hub.Task.Reaper.Created = reaperCreated
	})

	// Create bucket for task
	bucket := &model.Bucket{Path: "/tmp/task-bucket"}
	err = db.Create(bucket).Error
	g.Expect(err).To(gomega.BeNil())

	// Create old task in Created state without TTL
	oldTime := time.Now().Add(-2 * Settings.Hub.Task.Reaper.Created)
	task := &model.Task{
		Name:  "CreatedTaskNoTTL",
		State: task.Created,
	}
	task.CreateTime = oldTime
	task.BucketID = &bucket.ID
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &TaskReaper{DB: db}
	reaper.Run()

	// Verify task was released (bucket cleared) but NOT deleted
	var t1 model.Task
	err = db.First(&t1, task.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(t1.BucketID).To(gomega.BeNil())
	g.Expect(t1.Reaped).To(gomega.BeTrue())
}

// TestGroupReaper_ReadyStateWithTasks tests that groups with tasks are not deleted.
func TestGroupReaper_ReadyStateWithTasks(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db, err := setupDB()
	g.Expect(err).To(gomega.BeNil())

	// Create bucket for group
	bucket := &model.Bucket{Path: "/tmp/group-bucket"}
	err = db.Create(bucket).Error
	g.Expect(err).To(gomega.BeNil())

	// Create old group in Ready state
	oldTime := time.Now().Add(-2 * time.Hour)
	group := &model.TaskGroup{
		Name:  "GroupWithTasks",
		State: task.Ready,
	}
	group.CreateTime = oldTime
	group.BucketID = &bucket.ID
	err = db.Create(group).Error
	g.Expect(err).To(gomega.BeNil())

	// Create task belonging to group
	task := &model.Task{
		Name:        "TaskInGroup",
		TaskGroupID: &group.ID,
	}
	err = db.Create(task).Error
	g.Expect(err).To(gomega.BeNil())

	// Run reaper
	reaper := &GroupReaper{DB: db}
	reaper.Run()

	// Verify group was released (bucket cleared) but NOT deleted
	var g1 model.TaskGroup
	err = db.Preload("Tasks").First(&g1, group.ID).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(g1.BucketID).To(gomega.BeNil())
	g.Expect(len(g1.Tasks)).To(gomega.Equal(1))
}

// setupDB creates an in-memory SQLite database for testing.
func setupDB() (db *gorm.DB, err error) {
	db, err = gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	if err != nil {
		return
	}

	// Auto-migrate all required tables
	err = db.AutoMigrate(
		&model.Application{},
		&model.Task{},
		&model.TaskGroup{},
		&model.TaskReport{},
		&model.Bucket{},
		&model.File{},
		&model.RuleSet{},
		&model.Rule{},
		&model.Target{},
		&model.AnalysisProfile{},
	)

	return
}

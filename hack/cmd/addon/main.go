/*
TEST addon adapter.
This is an example of an addon adapter that lists files
and creates an application bucket for each. Error handling is
deliberately minimized to reduce code clutter.
*/
package main

import (
	"bytes"
	"errors"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"os"
	"os/exec"
	pathlib "path"
	"strings"
	"time"
)

var (
	// hub integration.
	addon = hub.Addon
	Log   = hub.Log
)

const (
	BucketDir = "list"
	TmpDir = "/tmp/list"
)

type SoftError = hub.SoftError

//
// main
func main() {
	addon.Run(func() (err error) {
		//
		// Get the addon data associated with the task.
		d := &Data{}
		_ = addon.DataWith(d)
		if err != nil {
			return
		}
		//
		// Get application.
		application, err := addon.Task.Application()
		if err != nil {
			return
		}
		//
		// Find files.
		paths, _ := find(d.Path, 25)
		//
		// List directory.
		err = listDir(d, application, paths)
		if err != nil {
			return
		}
		//
		// Set fact.
		application.Facts["Listed"] = true
		err = addon.Application.Update(application)
		if err != nil {
			return
		}
		//
		// Add tags.
		err = addTags(application, "LISTED", "TEST", "OTHER")
		if err != nil {
			return
		}
		return
	})
}

//
// listDir builds and populates the bucket.
func listDir(d *Data, application *api.Application, paths []string) (err error) {
	//
	// Task update: Update the task with total number of
	// items to be processed by the addon.
	addon.Total(len(paths))
	//
	// List directory.
	output := TmpDir
	_ = os.RemoveAll(output)
	_ = os.MkdirAll(output, 0777)
	//
	// Write files.
	for _, p := range paths {
		var b []byte
		//
		// Read file.
		b, err = os.ReadFile(p)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				continue
			}
			return
		}
		//
		// Task update: The current addon activity.
		target := pathlib.Join(
			output,
			pathlib.Base(p))
		addon.Activity("writing: %s", p)
		//
		// Write file.
		err = os.WriteFile(
			target,
			b,
			0666)
		if err != nil {
			return
		}
		time.Sleep(time.Second)
		//
		// Task update: Increment the number of completed
		// items processed by the addon.
		addon.Increment()
	}
	//
	// Build the index.
	err = buildIndex(output)
	if err != nil {
		return
	}
	//
	// Upload list directory.
	addon.Activity("[BUCKET] uploading %s => bucket/%s.", output, BucketDir)
	bucket := addon.Application.Bucket(application.ID)
	err = bucket.Put(output, BucketDir)
	if err != nil {
		return
	}
	//
	// play with buckets.
	err = playWithBucket(bucket)
	if err != nil {
		return
	}
	//
	// play with files.
	err = playWithFiles()
	if err != nil {
		return
	}
	//
	// Task update: update the current addon activity.
	addon.Activity("done")
	return
}

//
// playWithBucket
func playWithBucket(bucket *hub.Bucket) (err error) {
	tmpDir2 := "/tmp/list2"
	_ = os.MkdirAll(tmpDir2, 0777)
	//
	// Download list directory.
	err = bucket.Get(BucketDir, tmpDir2)
	if err != nil {
		return
	}
	//
	// Delete the index.
	err = bucket.Delete(BucketDir + "/index.html")
	if err != nil {
		return
	}
	//
	// Upload the index.
	err = bucket.Put(TmpDir + "/index.html", BucketDir + "/index.html")
	if err != nil {
		return
	}
	//
	// Add file.
	elmer := tmpDir2 + "/elmer"
	_, _ = os.Create(elmer)
	err = bucket.Put(elmer, BucketDir + "/networks")
	if err != nil {
		return
	}

	return
}

//
// Build index.html
func buildIndex(output string) (err error) {
	addon.Activity("Building index.")
	time.Sleep(time.Second)
	dir := output
	path := pathlib.Join(dir, "index.html")
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	body := []string{"<ul>"}
	list, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, name := range list {
		body = append(
			body,
			"<li><a href=\""+name.Name()+"\">"+name.Name()+"</a></li>")
	}

	body = append(body, "</ul>")

	_, _ = f.WriteString(strings.Join(body, "\n"))

	return
}

//
// find files.
func find(path string, max int) (paths []string, err error) {
	Log.Info("Listing.", "path", path)
	cmd := exec.Command(
		"find",
		path,
		"-maxdepth",
		"1",
		"-type",
		"f",
		"-readable")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		Log.Info(stderr.String())
		return
	}

	paths = strings.Fields(stdout.String())
	if len(paths) > max {
		paths = paths[:max]
	}

	Log.Info("List found.", "paths", paths)

	return
}

//
// addTags ensure tags created and associated with application.
// Ensure tag exists and associated with the application.
func addTags(application *api.Application, names ...string) (err error) {
	addon.Activity("Adding tags: %v", names)
	var wanted []uint
	//
	// Ensure type exists.
	tp := &api.TagType{
		Name: "DIRECTORY",
		Color: "#2b9af3",
		Rank: 3,
	}
	err = addon.TagType.Ensure(tp)
	if err != nil {
		return
	}
	//
	// Ensure tags exist.
	for _, name := range names {
		tag := &api.Tag{
			Name: name,
			TagType: api.Ref{
				ID: tp.ID,
			}}
		err = addon.Tag.Ensure(tag)
		if err == nil {
			wanted = append(wanted, tag.ID)
		} else {
			return
		}
	}
	//
	// Associate tags.
	tags := addon.Application.Tags(application.ID)
	for _, id := range wanted {
		err = tags.Add(id)
		if err != nil {
			return
		}
	}
	return
}

//
// Play with files.
func playWithFiles() (err error) {
	f, err := addon.File.Put("/etc/hosts")
	if err != nil {
		return
	}
	err = addon.File.Get(f.ID, "/tmp/hosts2")
	if err != nil {
		return
	}
	err = addon.File.Get(f.ID, "/tmp")
	if err != nil {
		return
	}
	err = addon.File.Delete(f.ID)
	if err != nil {
		return
	}
	return
}

//
// Data Addon input.
type Data struct {
	// Path to be listed.
	Path string `json:"path"`
}

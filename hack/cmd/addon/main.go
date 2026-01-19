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
	"os"
	"os/exec"
	pathlib "path"
	"strconv"
	"strings"
	"time"

	hub "github.com/konveyor/tackle2-hub/shared/addon"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"k8s.io/apimachinery/pkg/util/rand"
)

var (
	// hub integration.
	addon = hub.Addon
	Log   = hub.Log
)

const (
	BucketDir = "list"
	TmpDir    = "/tmp/list"
)

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
		facts := addon.Application.Facts(application.ID)
		facts.Source("addon")
		err = facts.Set("Listed", true)
		if err != nil {
			return
		}
		//
		// Get fact.
		var factValue bool
		err = facts.Get("Listed", &factValue)
		if err != nil {
			return
		}
		//
		// Replace facts.
		err = facts.Replace(
			api.Map{
				"Listed": true,
				"Color":  "blue",
				"Length": 100,
			})
		if err != nil {
			return
		}
		//
		// Add tags.
		err = addTags(application, "addon", "LISTED", "TEST", "OTHER")
		if err != nil {
			return
		}
		//
		// Replace tags.
		err = replaceTags(application, "addon", "TEST", "EXAMPLE", "REPLACED")
		if err != nil {
			return
		}
		return
	})
}

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

// playWithBucket
func playWithBucket(bucket hub.BucketContent) (err error) {
	tmpDir := tmpDir()
	defer func() {
		_ = nas.RmDir(tmpDir)
	}()
	//
	// Download list directory.
	err = bucket.Get(BucketDir, tmpDir)
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
	err = bucket.Put(TmpDir+"/index.html", BucketDir+"/index.html")
	if err != nil {
		return
	}
	//
	// Add file.
	elmer := tmpDir + "/elmer"
	_, _ = os.Create(elmer)
	err = bucket.Put(elmer, BucketDir+"/networks")
	if err != nil {
		return
	}

	return
}

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

// addTags ensure tags created and associated with application.
// Ensure tag exists and associated with the application.
func addTags(application *api.Application, source string, names ...string) (err error) {
	addon.Activity("Adding tags: %v", names)
	var wanted []uint
	//
	// Ensure type exists.
	tp := &api.TagCategory{
		Name:  "DIRECTORY",
		Color: "#2b9af3",
	}
	err = addon.TagCategory.Ensure(tp)
	if err != nil {
		return
	}
	//
	// Ensure tags exist.
	wanted, err = ensureTags(tp.ID, names...)
	if err != nil {
		return
	}
	//
	// Associate tags.
	tags := addon.Application.Tags(application.ID)
	tags.Source(source)
	for _, id := range wanted {
		err = tags.Ensure(id)
		if err != nil {
			return
		}
	}
	return
}

// replaceTags replaces current set of tags for the source with a new set.
// Ensures desired tags exist before replacing.
func replaceTags(application *api.Application, source string, names ...string) (err error) {
	addon.Activity("Replacing tags: %v", names)
	var wanted []uint
	//
	// Ensure type exists.
	tp := &api.TagCategory{
		Name:  "DIRECTORY",
		Color: "#2b9af3",
	}
	err = addon.TagCategory.Ensure(tp)
	if err != nil {
		return
	}
	//
	// Ensure tags exist.
	wanted, err = ensureTags(tp.ID, names...)
	if err != nil {
		return
	}
	//
	// Associate tags.
	tags := addon.Application.Tags(application.ID)
	tags.Source(source)
	err = tags.Replace(wanted)
	if err != nil {
		return
	}

	return
}

func ensureTags(category uint, names ...string) (ids []uint, err error) {
	for _, name := range names {
		tag := &api.Tag{
			Name: name,
			Category: api.Ref{
				ID: category,
			}}
		err = addon.Tag.Ensure(tag)
		if err == nil {
			ids = append(ids, tag.ID)
		} else {
			return
		}
	}
	return
}

func tmpDir() (p string) {
	p = "/tmp/pid-" + strconv.Itoa(rand.Int())
	_ = os.MkdirAll(p, 0777)
	return
}

// Data Addon input.
type Data struct {
	// Path to be listed.
	Path string `json:"path"`
}

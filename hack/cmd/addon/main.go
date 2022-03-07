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
	"strconv"
	"strings"
	"time"
)

var (
	// hub integration.
	addon = hub.Addon
	Log   = hub.Log
)

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
		// Find files.
		paths, _ := find(d.Path, 25)
		//
		// Ensure bucket.
		err = ensureBucket(d, paths)
		if err != nil {
			return
		}
		return
	})
}

//
// ensureBucket builds and populates the bucket.
func ensureBucket(d *Data, paths []string) (err error) {
	//
	// Task update: Update the task with total number of
	// items to be processed by the addon.
	addon.Total(len(paths))
	//
	// Ensure the bucket.
	addon.Activity("Ensuring bucket.")
	bucket, err := addon.Bucket.Ensure(d.Application, "Listing")
	if err == nil {
		addon.Activity("Using bucket: id=%d", bucket.ID)
	} else {
		return
	}
	defer func() {
		if err != nil {
			_ = addon.Bucket.Delete(bucket)
		}
	}()
	addon.Activity("Purging bucket.")
	err = addon.Bucket.Purge(bucket)
	if err != nil {
		return
	}
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
			bucket.Path,
			pathlib.Base(p))
		addon.Activity("writing: %s", p)
		//
		// Write file.
		err = os.WriteFile(
			target,
			b,
			0644)
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
	err = buildIndex(bucket)
	if err != nil {
		return
	}
	//
	// Task update: update the current addon activity.
	addon.Activity("done")
	return
}

//
// Build index.html
func buildIndex(bucket *api.Bucket) (err error) {
	addon.Activity("Building index.")
	time.Sleep(time.Second)
	dir := bucket.Path
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
			"<li><a href=\""+name.Name()+"\">"+name.Name()+"</a>")
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
// Tag application.
func tag(d *Data) (err error) {
	//
	// Fetch application.
	application, _ := addon.Application.Get(d.Application)
	//
	// Create tag.
	tag := &api.Tag{}
	tag.Name = "MyTag"
	tag.TagType.ID = 1
	err = addon.Tag.Create(tag)
	if err != nil {
		return
	}
	//
	// append tag.
	application.Tags = append(
		application.Tags,
		strconv.Itoa(int(tag.ID)))
	//
	// Update application.
	err = addon.Application.Update(application)
	return
}

//
// Data Addon data passed in the secret.
type Data struct {
	// Application ID.
	Application uint `json:"application"`
	// Path to be listed.
	Path string `json:"path"`
	// Delay on error (minutes).
	Delay int `json:"delay"`
}

//
// Delay as specified.
func (d *Data) delay() {
	if d.Delay > 0 {
		duration := time.Minute * time.Duration(d.Delay)
		time.Sleep(duration)
	}
}

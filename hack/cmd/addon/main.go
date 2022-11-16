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
		// Tag
		err = tag(d, application)
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
	output := pathlib.Join(application.Bucket, "list")
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
	// Task update: update the current addon activity.
	addon.Activity("done")
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
func tag(d *Data, application *api.Application) (err error) {
	//
	// Create tag.
	tag := &api.Tag{}
	tag.Name = "LISTED"
	tag.TagType.ID = 1
	err = addon.Tag.Create(tag)
	if err != nil {
		if errors.Is(err, &hub.Conflict{}) {
			err = nil
		} else {
			return
		}
	}
	//
	// add tag.
	for _, ref := range application.Tags {
		if ref.Name == tag.Name {
			return
		}
	}
	application.Tags = append(
		application.Tags,
		api.Ref{ID: tag.ID})
	//
	// Update application.
	err = addon.Application.Update(application)
	return
}

//
// Data Addon input.
type Data struct {
	// Path to be listed.
	Path string `json:"path"`
}

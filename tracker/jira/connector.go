package jira

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/konveyor/tackle2-hub/model"
	"io"
	"net/http"
	"strings"
	"time"
)

type Connector struct {
	tracker *model.Tracker
}

func (r *Connector) With(t *model.Tracker) {
	r.tracker = t
	_ = r.tracker.Identity.Decrypt()
}

// Create the ticket in JIRA.
func (r *Connector) Create(t *model.Ticket) (err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Summary:     fmt.Sprintf("Migrate %s", t.Application.Name),
			Description: "Created by Konveyor.",
			Type:        jira.IssueType{Name: t.Kind},
			Project:     jira.Project{Key: t.Parent},
		},
	}
	issue, response, err := client.Issue.Create(&i)
	err = handleJiraError(response, err)
	if err != nil {
		t.Error = true
		t.Message = err.Error()
		t.LastUpdated = time.Now()
		err = nil
		return
	}
	t.Created = true
	t.Error = false
	t.Message = ""
	t.Reference = issue.Key
	t.Link = issue.Self
	t.LastUpdated = time.Now()

	return
}

func (r *Connector) RefreshAll() (tickets map[*model.Ticket]bool, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	tickets = make(map[*model.Ticket]bool)
	var keys []string
	for i := range r.tracker.Tickets {
		t := &r.tracker.Tickets[i]
		if t.Reference != "" {
			keys = append(keys, t.Reference)
			tickets[t] = false
		}
	}

	if len(keys) == 0 {
		return
	}

	jql := fmt.Sprintf("key IN (%s)", strings.Join(keys, ","))
	issues, response, err := client.Issue.Search(jql, &jira.SearchOptions{Expand: "status"})
	err = handleJiraError(response, err)
	if err != nil {
		// JIRA returns a 400 if the search returned no results.
		if response.StatusCode == http.StatusBadRequest {
			err = nil
		}
		return
	}
	issuesByKey := make(map[string]*jira.Issue)
	for i := range issues {
		issue := &issues[i]
		issuesByKey[issue.Key] = issue
	}
	lastUpdated := time.Now()
	for i := range r.tracker.Tickets {
		t := &r.tracker.Tickets[i]
		issue, found := issuesByKey[t.Reference]
		if !found {
			continue
		}
		t.LastUpdated = lastUpdated
		if issue.Fields != nil && issue.Fields.Status != nil {
			t.Status = issue.Fields.Status.Name
		}
		tickets[t] = true
	}
	return
}

func (r *Connector) GetMetadata() (metadata model.Metadata, err error) {
	client, err := r.client()
	if err != nil {
		return
	}
	meta, response, err := client.Issue.GetCreateMetaWithOptions(nil)
	err = handleJiraError(response, err)
	if err != nil {
		return
	}

	for _, p := range meta.Projects {
		project := model.Project{ID: p.Id, Key: p.Key, Name: p.Name}
		for _, it := range p.IssueTypes {
			issueType := model.IssueType{ID: it.Id, Name: it.Name}
			project.IssueTypes = append(project.IssueTypes, issueType)
		}
		metadata.Projects = append(metadata.Projects, project)
	}

	return
}

func (r *Connector) client() (client *jira.Client, err error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if r.tracker.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	jiraTransport := jira.BasicAuthTransport{
		Username:  r.tracker.Identity.User,
		Password:  r.tracker.Identity.Password,
		Transport: transport,
	}

	client, err = jira.NewClient(jiraTransport.Client(), r.tracker.URL)
	if err != nil {
		return
	}
	return
}

// TestConnection to Jira Cloud.
func (r *Connector) TestConnection() (connected bool, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	_, response, err := client.User.GetSelf()
	err = handleJiraError(response, err)
	if err != nil {
		return
	}

	connected = true
	return
}

func handleJiraError(response *jira.Response, in error) (out error) {
	if in == nil {
		return
	}

	if response == nil {
		out = in
		return
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		out = in
		return
	}

	var jiraErr jira.Error
	err = json.Unmarshal(body, &jiraErr)
	if err != nil {
		out = in
		return
	}

	out = errors.New(strings.Join(jiraErr.ErrorMessages, " "))
	return
}

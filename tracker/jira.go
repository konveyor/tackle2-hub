package tracker

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andygrunwald/go-jira"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/metrics"
	"github.com/konveyor/tackle2-hub/model"
	"io"
	"net/http"
	"strings"
	"time"
)

//
// JiraConnector for the Jira Cloud API
type JiraConnector struct {
	tracker *model.Tracker
}

//
// With updates the connector with the Tracker model.
func (r *JiraConnector) With(t *model.Tracker) {
	r.tracker = t
	_ = r.tracker.Identity.Decrypt()
}

//
// Create the ticket in Jira.
func (r *JiraConnector) Create(t *model.Ticket) (err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Summary:     fmt.Sprintf("Migrate %s", t.Application.Name),
			Description: "Created by Konveyor.",
			Type:        jira.IssueType{ID: t.Kind},
			Project:     jira.Project{ID: t.Parent},
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
	metrics.IssuesExported.Inc()

	return
}

//
// RefreshAll retrieves fresh status information for all the tracker's tickets.
func (r *JiraConnector) RefreshAll() (tickets map[*model.Ticket]bool, err error) {
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
		if response != nil && response.StatusCode == http.StatusBadRequest {
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
		t.Status = status(issue)
		tickets[t] = true
	}
	return
}

//
// Projects returns a list of Projects.
func (r *JiraConnector) Projects() (projects []Project, err error) {
	client, err := r.client()
	if err != nil {
		return
	}
	list, response, err := client.Project.GetList()
	err = handleJiraError(response, err)
	if err != nil {
		return
	}
	for _, p := range *list {
		project := Project{
			ID:   p.ID,
			Name: p.Name,
		}
		projects = append(projects, project)
	}
	return
}

//
// Project returns a Project.
func (r *JiraConnector) Project(id string) (project Project, err error) {
	client, err := r.client()
	if err != nil {
		return
	}
	p, response, err := client.Project.Get(id)
	err = handleJiraError(response, err)
	if err != nil {
		return
	}
	project = Project{
		ID:   p.ID,
		Name: p.Name,
	}

	return
}

//
// IssueTypes returns a list of IssueTypes for a Project.
func (r *JiraConnector) IssueTypes(id string) (issueTypes []IssueType, err error) {
	client, err := r.client()
	if err != nil {
		return
	}
	project, response, err := client.Project.Get(id)
	err = handleJiraError(response, err)
	if err != nil {
		return
	}
	for _, i := range project.IssueTypes {
		issueType := IssueType{
			ID:   i.ID,
			Name: i.Name,
		}
		issueTypes = append(issueTypes, issueType)
	}
	return
}

//
// client builds a Jira API client for the tracker.
func (r *JiraConnector) client() (client *jira.Client, err error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if r.tracker.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var httpclient *http.Client
	switch r.tracker.Identity.Kind {
	case BearerAuth:
		jiraTransport := jira.BearerAuthTransport{
			Token:     r.tracker.Identity.Key,
			Transport: transport,
		}
		httpclient = jiraTransport.Client()
	case BasicAuth:
		jiraTransport := jira.BasicAuthTransport{
			Username:  r.tracker.Identity.User,
			Password:  r.tracker.Identity.Password,
			Transport: transport,
		}
		httpclient = jiraTransport.Client()
	default:
		err = liberr.New("unsupported identity kind", "kind", r.tracker.Identity.Kind)
		return
	}

	client, err = jira.NewClient(httpclient, r.tracker.URL)
	if err != nil {
		return
	}
	return
}

//
// TestConnection to Jira Cloud.
func (r *JiraConnector) TestConnection() (connected bool, err error) {
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

//
// status returns a normalized status based on the issue status category.
func status(issue *jira.Issue) (s string) {
	key := ""
	if issue.Fields != nil && issue.Fields.Status != nil {
		key = issue.Fields.Status.StatusCategory.Key
	}

	switch key {
	case jira.StatusCategoryToDo:
		s = New
	case jira.StatusCategoryInProgress:
		s = InProgress
	case jira.StatusCategoryComplete:
		s = Done
	default:
		s = Unknown
	}
	return
}

//
// handleJiraError simplifies dealing with errors from the Jira API.
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

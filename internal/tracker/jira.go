package tracker

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/metrics"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
)

const IssueTypeEpic = "Epic"

const (
	JiraEndpointBase    = "rest/api/2"
	JiraEndpointProject = JiraEndpointBase + "/project"
	JiraEndpointIssue   = JiraEndpointBase + "/issue"
	JiraEndpointSearch  = JiraEndpointBase + "/search"
	JiraEndpointMyself  = JiraEndpointBase + "/myself"
)

// JiraConnector for the Jira Cloud API
type JiraConnector struct {
	tracker *model.Tracker
}

// With updates the connector with the Tracker model.
func (r *JiraConnector) With(t *model.Tracker) {
	r.tracker = t
	_ = secret.Decrypt(r.tracker.Identity)
}

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

	req, err := client.NewRequest(http.MethodPost, JiraEndpointIssue, &i)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	response, err := client.Do(req, &i)
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
	t.Reference = i.Key
	t.Link = fmt.Sprintf("%s/browse/%s", r.tracker.URL, i.Key)
	t.LastUpdated = time.Now()
	metrics.IssuesExported.Inc()

	return
}

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
	query := url.Values{}
	query.Add("expand", "status")
	query.Add("jql", jql)
	req, err := client.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", JiraEndpointSearch, query.Encode()), nil)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	results := struct {
		Issues []jira.Issue `json:"issues"`
	}{}
	response, err := client.Do(req, &results)
	err = handleJiraError(response, err)
	if err != nil {
		// JIRA returns a 400 if the search returned no results.
		if response != nil && response.StatusCode == http.StatusBadRequest {
			err = nil
		}
		return
	}
	issuesByKey := make(map[string]*jira.Issue)
	for i := range results.Issues {
		issue := &results.Issues[i]
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

// Projects returns a list of Projects.
func (r *JiraConnector) Projects() (projects []Project, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	req, err := client.NewRequest(http.MethodGet, JiraEndpointProject, nil)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	var list jira.ProjectList
	response, err := client.Do(req, &list)
	err = handleJiraError(response, err)
	if err != nil {
		return
	}
	for _, p := range list {
		project := Project{
			ID:   p.ID,
			Name: p.Name,
		}
		projects = append(projects, project)
	}
	return
}

// Project returns a Project.
func (r *JiraConnector) Project(id string) (project Project, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	req, err := client.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", JiraEndpointProject, id), nil)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	var p jira.Project
	response, err := client.Do(req, &p)
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

// IssueTypes returns a list of IssueTypes for a Project.
func (r *JiraConnector) IssueTypes(id string) (issueTypes []IssueType, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	req, err := client.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", JiraEndpointProject, id), nil)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	var project jira.Project
	response, err := client.Do(req, &project)
	err = handleJiraError(response, err)
	if err != nil {
		return
	}
	for _, i := range project.IssueTypes {
		if i.Subtask || i.Name == IssueTypeEpic {
			continue
		}
		issueType := IssueType{
			ID:   i.ID,
			Name: i.Name,
		}
		issueTypes = append(issueTypes, issueType)
	}
	return
}

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
	wrapped := clientWrapper{client: httpclient}
	client, err = jira.NewClient(&wrapped, r.tracker.URL)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// TestConnection to Jira Cloud.
func (r *JiraConnector) TestConnection() (connected bool, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	req, err := client.NewRequest(http.MethodGet, JiraEndpointMyself, nil)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	var user jira.User
	resp, err := client.Do(req, &user)
	if err != nil {
		err = handleJiraError(resp, err)
		return
	}

	connected = true
	return
}

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

	var jiraErr jiraError
	err = json.Unmarshal(body, &jiraErr)
	if err != nil {
		out = in
		return
	}

	out = &jiraErr
	return
}

// clientWrapper wraps the http client used by the jira client.
type clientWrapper struct {
	client *http.Client
}

// Do applies an Accept header before performing the request.
func (r *clientWrapper) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/json")
	resp, err := r.client.Do(req)
	return resp, err
}

// jiraError is a catch-all structure for the various ways that the
// Jira API can report an error in JSON format.
type jiraError struct {
	Message       string            `json:"message"`
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

// Error reports the consolidated error message.
func (r *jiraError) Error() (error string) {
	if len(r.Message) > 0 {
		error = r.Message
		return
	}
	if len(r.Errors) > 0 {
		for k, v := range r.Errors {
			r.ErrorMessages = append(r.ErrorMessages, fmt.Sprintf("%s: %s", k, v))
		}
	}
	if len(r.ErrorMessages) > 0 {
		error = strings.Join(r.ErrorMessages, ", ")
		return
	}
	return
}

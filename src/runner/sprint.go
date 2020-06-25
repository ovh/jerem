package runner

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	warp "github.com/PierreZ/Warp10Exporter"
	jira "github.com/andygrunwald/go-jira"
	"github.com/ovh/jerem/src/core"
	log "github.com/sirupsen/logrus"
)

const impedimentField = "customfield_11028"

// SprintRunner runner handling sprint metrics
func SprintRunner(config core.Config) {
	tp := jira.BasicAuthTransport{
		Username: config.Jira.Username,
		Password: config.Jira.Password,
	}
	jiraClient, err := jira.NewClient(tp.Client(), config.Jira.URL)
	if err != nil {
		log.WithError(err).Error("Fail to get jira client")
		return
	}

	batch := warp.NewBatch()

	for _, project := range config.Projects {
		options := &jira.GetAllSprintsOptions{State: "active"}
		sprints, _, err := jiraClient.Board.GetAllSprintsWithOptions(project.Board, options)
		if err != nil {
			log.WithField("project", project.Name).WithError(err).Warn("Fail to get sprints")
			return
		}

		log.Debug(config.Jira.ClosedStatuses)
		closed := fmt.Sprintf("(%s)", strings.Join(config.Jira.ClosedStatuses, ","))

		for _, sprint := range sprints.Values {
			processSprint(jiraClient, sprint, project, batch, closed)
		}

		// Get last day closed impediment and set issue timespent at its creation date
		var closedImpediments []jira.Issue

		err = jiraClient.Issue.SearchPages(fmt.Sprintf("(project = \"%s\" %s) AND status in %s AND labels in (Impediment, impediment) AND updated >= -1d AND timespent is not EMPTY", project.Name, project.Jql, closed), &jira.SearchOptions{
			Fields: []string{"id", "key", "project", "created", "timespent"},
		}, func(issue jira.Issue) error {
			closedImpediments = append(closedImpediments, issue)
			return nil
		})
		if err != nil {
			log.WithField("project", project.Name).WithError(err).Warn("Fail to get sprint issues")
			return
		}

		if len(closedImpediments) > 0 {
			gts := warp.NewGTS(fmt.Sprintf("jerem.jira.impediment.total.created")).WithLabels(warp.Labels{
				"project": project.Label,
				"type":    "daily",
				"value":   "timespent",
			})
			for _, impediment := range closedImpediments {
				gts.AddDatapoint(time.Time(impediment.Fields.Created), impediment.Fields.TimeSpent)
			}
			batch.Register(gts)
		}
	}

	var b bytes.Buffer
	batch.Print(&b)
	log.Debug(b.String())
	if len(*batch) != 0 {
		err = batch.Push(config.Metrics.URL, config.Metrics.Token)
		if err != nil {
			log.WithError(err).Error("Fail to push metrics")
		}
	}
}

func getSprintMetric(name string, projectLabel, sprint string) *warp.GTS {
	return warp.NewGTS(fmt.Sprintf("jerem.jira.sprint.%s", name)).WithLabels(warp.Labels{
		"project": projectLabel,
		"sprint":  sprint,
	})
}

func getImpedimentType(field string, issue jira.Issue) (string, error) {

	v, ok := issue.Fields.Unknowns.Value(field)
	if !ok || v == nil {
		return "unknown", nil
	}
	switch items := v.(type) {
	case []interface{}:
		for _, item := range items {
			switch castMap := item.(type) {
			case map[string]interface{}:
				if val, ok := castMap["value"]; ok {
					return val.(string), nil
				}
			}
		}

	}

	return "unknown", nil
}

func getImpedimentSprintMetric(name, projectLabel, sprint string) *warp.GTS {
	return warp.NewGTS(fmt.Sprintf("jerem.jira.impediment.%s", name)).WithLabels(warp.Labels{
		"project": projectLabel,
		"type":    "sprint",
		"sprint":  sprint,
	})
}

//GetIssuesForSprint overrides go-jira GetIssuesForSprint to apply jql filter on issues
func GetIssuesForSprint(jiraClient *jira.Client, sprintID int, jql string) ([]jira.Issue, *jira.Response, error) {
	if jql != "" {
		apiEndpoint := fmt.Sprintf("rest/agile/1.0/sprint/%d/issue?jql=%s", sprintID, url.QueryEscape(jql))
		req, err := jiraClient.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return nil, nil, err
		}
		result := new(jira.IssuesInSprintResult)
		resp, err := jiraClient.Do(req, result)
		if err != nil {
			err = jira.NewJiraError(resp, err)
		}
		log.WithFields(log.Fields{"project": jql, "count": len(result.Issues)}).Debug("Sprint issue")
		return result.Issues, resp, err
	}
	return jiraClient.Sprint.GetIssuesForSprint(sprintID)
}

func processSprint(jiraClient *jira.Client, sprint jira.Sprint, project core.Project, batch *warp.Batch, jiraCloseStatus string) {
	jql := ""
	if project.Jql != "" {
		jql = fmt.Sprintf("project=%s %s", project.Name, project.Jql)
	}
	issues, _, err := GetIssuesForSprint(jiraClient, sprint.ID, jql)
	if err != nil {
		log.WithFields(log.Fields{"sprint": sprint.Name, "project": project.Label}).
			WithError(err).Warn("Fail to get issue for sprint")
		return
	}

	storyPoints, _, _ := computeStoryPoints(issues, storyPointField)

	// Gen metrics
	now := time.Now().UTC()
	gts := getSprintMetric("storypoint.total", project.Label, "current").AddDatapoint(now, storyPoints["total"])
	batch.Register(gts)
	gts = getSprintMetric("storypoint.total", project.Label, sprint.Name).AddDatapoint(now, storyPoints["total"])
	batch.Register(gts)
	gts = getSprintMetric("storypoint.inprogress", project.Label, "current").AddDatapoint(now, storyPoints["indeterminate"])
	batch.Register(gts)
	gts = getSprintMetric("storypoint.inprogress", project.Label, sprint.Name).AddDatapoint(now, storyPoints["indeterminate"])
	batch.Register(gts)
	gts = getSprintMetric("storypoint.done", project.Label, "current").AddDatapoint(now, storyPoints["done"])
	batch.Register(gts)
	gts = getSprintMetric("storypoint.done", project.Label, sprint.Name).AddDatapoint(now, storyPoints["done"])
	batch.Register(gts)

	// Add start and end date in sprint events series
	gts = getSprintMetric("events", project.Label, "current").AddDatapoint(*sprint.StartDate, "start").AddDatapoint(*sprint.EndDate, "end")
	batch.Register(gts)
	gts = getSprintMetric("events", project.Label, sprint.Name).AddDatapoint(*sprint.StartDate, "start").AddDatapoint(*sprint.EndDate, "end")
	batch.Register(gts)

	// Get current sprint closed impediments
	var impediments []jira.Issue
	err = jiraClient.Issue.SearchPages(fmt.Sprintf("(project = \"%s\" %s) AND status in %s AND labels in (Impediment, impediment) AND updated >= %s AND updated <= %s AND timespent is not EMPTY", project.Name, project.Jql, jiraCloseStatus, sprint.StartDate.Format("2006-01-02"), sprint.EndDate.Format("2006-01-02")), &jira.SearchOptions{
		Fields: []string{"id", "key", "project", "labels", "summary", "status", "timespent", impedimentField},
	}, func(issue jira.Issue) error {
		impediments = append(impediments, issue)
		return nil
	})
	if err != nil {
		log.WithFields(log.Fields{"sprint": sprint.Name, "project": project.Label}).
			WithError(err).Warn("Fail to get sprint issues")
		return
	}

	impedimentCount := make(map[string]int)
	impedimentSecond := make(map[string]int)
	for _, impediment := range impediments {
		impedimentType, err := getImpedimentType(impedimentField, impediment)
		if err != nil {
			log.WithField("key", impediment.Key).WithError(err).Warn("Fail to get impediment type")
			continue
		}
		impedimentCount["total"] = impedimentCount["total"] + 1
		impedimentCount[impedimentType] = impedimentCount[impedimentType] + 1
		impedimentSecond["total"] = impedimentSecond["total"] + impediment.Fields.TimeSpent
		impedimentSecond[impedimentType] = impedimentSecond[impedimentType] + impediment.Fields.TimeSpent
	}

	gts = getImpedimentSprintMetric("total.count", project.Label, "current").AddDatapoint(now, impedimentCount["total"])
	batch.Register(gts)
	gts = getImpedimentSprintMetric("total.count", project.Label, sprint.Name).AddDatapoint(now, impedimentCount["total"])
	batch.Register(gts)
	gts = getImpedimentSprintMetric("total.timespent", project.Label, sprint.Name).AddDatapoint(now, impedimentSecond["total"])
	batch.Register(gts)
	gts = getImpedimentSprintMetric("total.timespent", project.Label, "current").AddDatapoint(now, impedimentSecond["total"])
	batch.Register(gts)

	for impedimentType, v := range impedimentCount {
		gts = getImpedimentSprintMetric(fmt.Sprintf("%s.count", impedimentType), project.Label, "current").AddDatapoint(now, v)
		batch.Register(gts)
		gts = getImpedimentSprintMetric(fmt.Sprintf("%s.count", impedimentType), project.Label, sprint.Name).AddDatapoint(now, v)
		batch.Register(gts)
	}
	for impedimentType, v := range impedimentSecond {
		gts = getImpedimentSprintMetric(fmt.Sprintf("%s.timespent", impedimentType), project.Label, "current").AddDatapoint(now, v)
		batch.Register(gts)
		gts = getImpedimentSprintMetric(fmt.Sprintf("%s.timespent", impedimentType), project.Label, sprint.Name).AddDatapoint(now, v)
		batch.Register(gts)
	}
}

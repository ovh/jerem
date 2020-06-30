package runner

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	warp "github.com/PierreZ/Warp10Exporter"
	jira "github.com/andygrunwald/go-jira"
	log "github.com/sirupsen/logrus"

	"github.com/ovh/jerem/src/core"
)

const storyPointField = "customfield_10006"

var quarterRegex = regexp.MustCompile(`^Q[1-4]-\d{2}$`)
var projectPrefix = "Project_"

// EpicRunner runner handling epic metrics
func EpicRunner(config core.Config) {
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

	// Get epics per project
	for _, project := range config.Projects {
		epics, err := getEpics(jiraClient, project)
		if err != nil {
			log.WithError(err).Error("Fail to get jira epics")
			return
		}

		// Count storypoints per epic
		for _, epic := range epics {

			status := getStatus(epic) // [undefined, new, indeterminate, done]

			if status == jira.StatusCategoryComplete {
				continue
			}

			// Get global project related to current epic
			global := "None"
			for _, label := range epic.Fields.Labels {
				if strings.HasPrefix(label, projectPrefix) {
					global = strings.TrimLeft(label, projectPrefix)
				}
			}

			// Search for quarter label
			for _, label := range epic.Fields.Labels {
				if quarterRegex.MatchString(label) {
					processEpic(jiraClient, epic, label, project.Label, global, batch)
				}

			}
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

func getEpics(jiraClient *jira.Client, project core.Project) ([]jira.Issue, error) {
	query := getEpicQuery(project)

	var epics []jira.Issue
	err := jiraClient.Issue.SearchPages(query, &jira.SearchOptions{
		Fields: []string{"id", "key", "project", "labels", "summary", "status"},
	}, func(issue jira.Issue) error {
		epics = append(epics, issue)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return epics, nil
}

func getEpicQuery(project core.Project) string {
	return fmt.Sprintf("(project = \"%s\" %s) AND issuetype = Epic", project.Name, project.Jql)
}

func processEpic(jiraClient *jira.Client, epic jira.Issue, quarter, projectLabel, global string, batch *warp.Batch) {
	issues, err := getIssues(jiraClient, epic.Key)
	if err != nil {
		log.WithField("key", epic.Key).WithError(err).Warn("Fail to get jira issues")
		return
	}

	storyPoints, unestimated, dependency := computeStoryPoints(issues, storyPointField)

	// Gen metrics
	now := time.Now().UTC()
	gts := getEpicMetric("storypoint", epic, quarter, projectLabel, global).AddDatapoint(now, storyPoints["total"])
	batch.Register(gts)
	gts = getEpicMetric("unestimated", epic, quarter, projectLabel, global).AddDatapoint(now, float64(unestimated))
	batch.Register(gts)
	gts = getEpicMetric("dependency", epic, quarter, projectLabel, global).AddDatapoint(now, float64(dependency))
	batch.Register(gts)
	gts = getEpicMetric("storypoint.inprogress", epic, quarter, projectLabel, global).AddDatapoint(now, storyPoints["indeterminate"])
	batch.Register(gts)
	gts = getEpicMetric("storypoint.done", epic, quarter, projectLabel, global).AddDatapoint(now, storyPoints["done"])
	batch.Register(gts)
}

func getIssues(jiraClient *jira.Client, epic string) ([]jira.Issue, error) {
	var issues []jira.Issue
	err := jiraClient.Issue.SearchPages(fmt.Sprintf("\"Epic Link\" = %s", epic), &jira.SearchOptions{
		Fields: []string{"id", "key", "labels", "summary", "status", storyPointField},
	}, func(issue jira.Issue) error {
		issues = append(issues, issue)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return issues, nil
}

func getStatus(issue jira.Issue) string {
	// /api/2/statuscategory => [undefined, new, indeterminate, done]
	if issue.Fields.Status == nil {
		log.WithField("key", issue.Key).Warn("No status")
		return "undefined"
	}
	return issue.Fields.Status.StatusCategory.Key
}

func getEpicMetric(name string, epic jira.Issue, quarter, projectLabel, global string) *warp.GTS {
	return warp.NewGTS(fmt.Sprintf("jerem.jira.epic.%s", name)).WithLabels(warp.Labels{
		"project": projectLabel,
		"key":     epic.Key,
		"summary": epic.Fields.Summary,
		"quarter": quarter,
		"global":  global,
	})
}

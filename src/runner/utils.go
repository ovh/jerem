package runner

import (
	jira "github.com/andygrunwald/go-jira"
	log "github.com/sirupsen/logrus"
)

var dependencyLabel = "dependency"

func computeStoryPoints(issues []jira.Issue, field string) (map[string]float64, int, int) {
	unestimated := 0
	dependency := 0
	storyPoints := make(map[string]float64)

	for _, issue := range issues {
		sp, err := getStoryPoints(field, issue)
		if err != nil {
			log.WithField("key", issue.Key).WithError(err).Warn("Fail to get story points")
			continue
		}
		storyPoints["total"] = storyPoints["total"] + sp

		for _, label := range issue.Fields.Labels {
			if label == dependencyLabel {
				dependency++
				break
			}
		}

		if sp == 0.0 {
			unestimated++
			continue
		}

		status := getStatus(issue) // [undefined, new, indeterminate, done]
		storyPoints[status] = storyPoints[status] + sp
	}

	return storyPoints, unestimated, dependency
}

func getStoryPoints(field string, issue jira.Issue) (float64, error) {
	v, ok := issue.Fields.Unknowns.Value(field)
	if !ok || v == nil {
		return 0, nil
	}

	sp, err := issue.Fields.Unknowns.Float(field)
	if err != nil {
		return 0, err
	}
	return sp, nil
}

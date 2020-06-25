package core

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config is jerem root config
type Config struct {
	Projects []Project
	Jira     Jira
	Metrics  Metrics
}

// Project define a jira project
type Project struct {
	Name  string
	Board int
	Jql   string
	Label string
}

// Jira define jira params
type Jira struct {
	Username       string
	Password       string
	URL            string
	ClosedStatuses []string
}

// Metrics define metrics params
type Metrics struct {
	Token string
	URL   string
}

// LoadConfig read config from viper
func LoadConfig() (Config, error) {
	config := Config{}

	jira, err := loadJira()
	if err != nil {
		return config, err
	}
	config.Jira = jira

	metrics, err := loadMetrics()
	if err != nil {
		return config, err
	}
	config.Metrics = metrics

	projects, err := loadProjects()
	if err != nil {
		return config, err
	}
	config.Projects = projects

	return config, nil
}

func loadProjects() ([]Project, error) {
	if !viper.IsSet("projects") {
		return nil, fmt.Errorf("projects is required")
	}

	settings := viper.AllSettings()

	projects, ok := settings["projects"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("projects should be an array")
	}

	var res []Project
	for idx, p := range projects {
		project, ok := p.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("project %d should be a map", idx)
		}

		if _, ok = project["name"]; !ok {
			return nil, fmt.Errorf("project %d name is required", idx)
		}
		name, ok := project["name"].(string)
		if !ok {
			return nil, fmt.Errorf("project %d name should be a string", idx)
		}

		if _, ok = project["board"]; !ok {
			return nil, fmt.Errorf("project %d board is required", idx)
		}
		board, ok := project["board"].(int)
		if !ok {
			return nil, fmt.Errorf("project %d board should be a number", idx)
		}

		jql := ""
		if _, ok = project["jql_filter"]; ok {
			jql, ok = project["jql_filter"].(string)
			if !ok {
				return nil, fmt.Errorf("jql filter '%d' should be a string", idx)
			}
			jql = fmt.Sprintf("AND (%s)", jql)
		}

		//If not label is provided, the project name is used
		label := name
		if _, ok = project["label"]; ok {
			label, ok = project["label"].(string)
			if !ok {
				return nil, fmt.Errorf("label '%d' should be a string", idx)
			}
		}
		label = strings.TrimSpace(label)
		res = append(res, Project{Name: name, Board: board, Jql: jql, Label: label})
	}

	return res, nil
}

func loadJira() (Jira, error) {
	jira := Jira{}
	if !viper.IsSet("jira") {
		return jira, fmt.Errorf("jira is required")
	}

	if !viper.IsSet("jira.username") {
		return jira, fmt.Errorf("jira username is required")
	}
	jira.Username = viper.GetString("jira.username")

	if !viper.IsSet("jira.password") {
		return jira, fmt.Errorf("jira password is required")
	}
	jira.Password = viper.GetString("jira.password")

	if !viper.IsSet("jira.url") {
		return jira, fmt.Errorf("jira url is required")
	}
	jira.URL = viper.GetString("jira.url")

	viper.SetDefault("jira.closed.statuses", []string{"Resolved", "Closed", "Done"})

	jira.ClosedStatuses = viper.GetStringSlice("jira.closed.statuses")
	return jira, nil
}

func loadMetrics() (Metrics, error) {
	metrics := Metrics{}
	if !viper.IsSet("metrics") {
		return metrics, fmt.Errorf("metrics is required")
	}

	if !viper.IsSet("metrics.token") {
		return metrics, fmt.Errorf("metrics token is required")
	}
	metrics.Token = viper.GetString("metrics.token")

	if !viper.IsSet("metrics.url") {
		return metrics, fmt.Errorf("metrics url is required")
	}
	metrics.URL = viper.GetString("metrics.url")

	return metrics, nil
}

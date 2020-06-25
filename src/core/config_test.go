package core

import (
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/require"
)

func loadConfig(assert *require.Assertions, config string) {
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(err)
}
func TestMissingProjects(t *testing.T) {
	assert := require.New(t)

	// Load an empty config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "projects is required")
}
func TestNoProjects(t *testing.T) {
	assert := require.New(t)

	// Load an empty projects config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "projects is required")
}
func TestNoProjectName(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - board: 96`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "project 0 name is required")
}
func TestNoProjectBoard(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "project 0 board is required")
}
func TestBadProjectBoard(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: abc`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "project 0 board should be a number")
}

func TestProjectOptionalLabelAndJqlLoad(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 94
    jql_filter: component = test
    label: test`
	loadConfig(assert, config)

	conf, err := LoadConfig()
	assert.Empty(err)
	assert.Equal(conf.Projects[0].Jql, "AND (component = test)", "The Jql filter should be surrouned by AND()")
	assert.Equal(conf.Projects[0].Label, "test", "Label should be loaded")
}

func TestProjectEmptyqlLoad(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 94`
	loadConfig(assert, config)

	conf, err := LoadConfig()
	assert.Empty(err)
	assert.Equal(conf.Projects[0].Jql, "", "The Jql filter should be empty")
}

func TestMissingJira(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "jira is required")
}
func TestMissingJiraUsername(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "jira username is required")
}
func TestMissingJiraPassword(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "jira password is required")
}
func TestMissingJiraURL(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "jira url is required")
}
func TestMissingMetrics(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "metrics is required")
}
func TestMissingMetricsURL(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  token: mytoken
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "metrics url is required")
}
func TestMissingMetricsToken(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	_, err := LoadConfig()
	assert.EqualError(err, "metrics token is required")
}
func TestParseConfig(t *testing.T) {
	assert := require.New(t)

	// Load config
	config := `
jira:
  username: jerem
  password: foo
  url: https://jira.com
metrics:
  url: https://metrics.ovh.net
  token: mytoken
projects:
  - name: K8S
    board: 96
  - name: OB
    board: 407`
	loadConfig(assert, config)

	cfg, err := LoadConfig()
	assert.NoError(err)
	assert.Len(cfg.Projects, 2)
	assert.Equal(cfg.Projects[0].Name, "K8S")
	assert.Equal(cfg.Projects[0].Board, 96)
	assert.Equal(cfg.Projects[1].Name, "OB")
	assert.Equal(cfg.Projects[1].Board, 407)
	assert.Equal(cfg.Jira.Username, "jerem")
	assert.Equal(cfg.Jira.Password, "foo")
	assert.Equal(cfg.Jira.URL, "https://jira.com")
}

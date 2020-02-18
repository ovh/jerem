package runner

import (
	"testing"

	"github.com/ovh/jerem/src/core"
	"github.com/stretchr/testify/require"
)

func TestGetEpicQuerysWithNoProjects(t *testing.T) {
	assert := require.New(t)

	projects := []core.Project{}

	query := getEpicQuery(projects)

	assert.Equal(query, "")
}
func TestGetEpicQuerysWithOneProjects(t *testing.T) {
	assert := require.New(t)

	projects := []core.Project{
		core.Project{
			Name: "PJ1",
		},
	}

	query := getEpicQuery(projects)

	assert.Equal(query, "(project = \"PJ1\") AND issuetype = Epic")
}
func TestGetEpicQuerysWithMultipleProjects(t *testing.T) {
	assert := require.New(t)

	projects := []core.Project{
		core.Project{
			Name: "PJ1",
		},
		core.Project{
			Name: "PJ2",
		},
	}

	query := getEpicQuery(projects)

	assert.Equal(query, "(project = \"PJ1\" OR project = \"PJ2\") AND issuetype = Epic")
}

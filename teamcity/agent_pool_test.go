package teamcity_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/leonj1/teamcity/teamcity"
	"github.com/stretchr/testify/assert"
)

func TestAgentPools_GetDefaultProject(t *testing.T) {
	client := setup()
	assert := assert.New(t)

	// this is hard-coded in TeamCity, so we may as well do the same
	defaultAgentPoolId := 0

	retrievedPool, err := client.AgentPools.GetByID(defaultAgentPoolId)
	assert.NoError(err)
	assert.Equal("Default", retrievedPool.Name)
	assert.Nil(retrievedPool.MaxAgents)
	// count of projects determined by content of integration_tests/teamcity_data.tar.gz
	// inside archive path 'data_dir/config/projects'
	assert.Equal(3, len(retrievedPool.Projects.Project))
}

func TestAgentPools_GetDefaultProjectByName(t *testing.T) {
	client := setup()
	assert := assert.New(t)

	retrievedPool, err := client.AgentPools.GetByName("Default")
	assert.NoError(err)
	assert.Equal(0, retrievedPool.Id)
	assert.Equal("Default", retrievedPool.Name)
	assert.Nil(retrievedPool.MaxAgents)
}

func TestAgentPools_Lifecycle(t *testing.T) {
	client := setup()
	assert := assert.New(t)

	agentPool := teamcity.CreateAgentPool{
		Name: fmt.Sprintf("test-%d", time.Now().Unix()),
	}
	createdPool, err := client.AgentPools.Create(agentPool)
	assert.NoError(err)
	assert.NotEmpty(createdPool.Id)
	assert.Equal(agentPool.Name, createdPool.Name)

	retrievedPool, err := client.AgentPools.GetByID(createdPool.Id)
	assert.NoError(err)
	assert.Equal(agentPool.Name, retrievedPool.Name)
	assert.Nil(retrievedPool.MaxAgents)

	assert.NoError(client.AgentPools.Delete(createdPool.Id))

	// confirm it's gone
	agentPools, err := client.AgentPools.List()
	assert.NoError(err)
	for _, pool := range agentPools.AgentPools {
		if pool.Name == agentPool.Name {
			t.Fatalf("Created agent pool still exists!")
		}
	}
}

func TestAgentPools_List(t *testing.T) {
	client := setup()
	assert := assert.New(t)

	agentPools, err := client.AgentPools.List()
	assert.NoError(err)

	// whilst other pools may have been added by other tests - the Default pool
	// cannot be removed, so can be used as test data
	assert.True(len(agentPools.AgentPools) > 0, "At least one agent pool should exist")

	found := false
	for _, pool := range agentPools.AgentPools {
		if pool.Name == "Default" {
			found = true
		}
	}

	assert.True(found, "Default Agent Pool was not found")
}

func TestAgentPools_ListForProject(t *testing.T) {
	client := setup()
	assert := assert.New(t)

	firstProjectData := getTestProjectData(fmt.Sprintf("Project %d", time.Now().Unix()), "")
	firstProject, err := client.Projects.Create(firstProjectData)
	assert.NoError(err)

	agentPool := teamcity.CreateAgentPool{
		Name: fmt.Sprintf("test-%d", time.Now().Unix()),
	}
	createdPool, err := client.AgentPools.Create(agentPool)
	assert.NoError(err)
	assert.NotEmpty(createdPool.Id)
	assert.Equal(agentPool.Name, createdPool.Name)

	retrievedPool, err := client.AgentPools.GetByID(createdPool.Id)
	assert.NoError(err)
	assert.Equal(agentPool.Name, retrievedPool.Name)
	assert.Nil(retrievedPool.MaxAgents)

	assignments, err := client.AgentPools.ListForProject(firstProject.ID)
	assert.NoError(err)
	assert.Equal(1, assignments.Count)
	assert.Equal("Default", assignments.AgentPools[0].Name)

	// assign the build
	assert.NoError(client.AgentPools.AssignProject(createdPool.Id, firstProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))

	assignments, err = client.AgentPools.ListForProject(firstProject.ID)
	assert.NoError(err)
	assert.Equal(2, assignments.Count)

	// remove it
	assert.NoError(client.AgentPools.UnassignProject(createdPool.Id, firstProject.ID))
	assert.False(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))

	// and confirm
	assignments, err = client.AgentPools.ListForProject(firstProject.ID)
	assert.NoError(err)
	assert.Equal(1, assignments.Count)
	assert.Equal("Default", assignments.AgentPools[0].Name)

	assert.NoError(client.Projects.Delete(firstProject.ID))
	assert.NoError(client.AgentPools.Delete(createdPool.Id))
}

func TestAgentPools_ProjectAssignment(t *testing.T) {
	client := setup()
	assert := assert.New(t)

	firstProjectData := getTestProjectData("First Project", "")
	secondProjectData := getTestProjectData("Second Project", "")

	firstProject, err := client.Projects.Create(firstProjectData)
	assert.NoError(err)
	secondProject, err := client.Projects.Create(secondProjectData)
	assert.NoError(err)

	agentPool := teamcity.CreateAgentPool{
		Name: fmt.Sprintf("test-%d", time.Now().Unix()),
	}
	createdPool, err := client.AgentPools.Create(agentPool)
	assert.NoError(err)
	assert.NotEmpty(createdPool.Id)
	assert.Equal(agentPool.Name, createdPool.Name)

	retrievedPool, err := client.AgentPools.GetByID(createdPool.Id)
	assert.NoError(err)
	assert.Equal(agentPool.Name, retrievedPool.Name)
	assert.Nil(retrievedPool.MaxAgents)

	// assign the build
	assert.NoError(client.AgentPools.AssignProject(createdPool.Id, firstProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))

	// assign another
	assert.NoError(client.AgentPools.AssignProject(createdPool.Id, secondProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, secondProject.ID))

	// remove the first
	assert.NoError(client.AgentPools.UnassignProject(createdPool.Id, firstProject.ID))
	assert.False(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, secondProject.ID))

	// re-assign the first
	assert.NoError(client.AgentPools.AssignProject(createdPool.Id, firstProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))
	assert.True(validateContainsProject(assert, client, createdPool.Id, secondProject.ID))

	// then remove everything
	assert.NoError(client.AgentPools.UnassignProject(createdPool.Id, firstProject.ID))
	assert.NoError(client.AgentPools.UnassignProject(createdPool.Id, secondProject.ID))
	assert.False(validateContainsProject(assert, client, createdPool.Id, firstProject.ID))
	assert.False(validateContainsProject(assert, client, createdPool.Id, secondProject.ID))

	assert.NoError(client.Projects.Delete(firstProject.ID))
	assert.NoError(client.Projects.Delete(secondProject.ID))
	assert.NoError(client.AgentPools.Delete(createdPool.Id))
}

func validateContainsProject(assert *assert.Assertions, client *teamcity.Client, poolId int, projectId string) bool {
	agentPool, err := client.AgentPools.GetByID(poolId)
	assert.NoError(err)

	if agentPool.Projects == nil {
		return false
	}

	for _, v := range agentPool.Projects.Project {
		if v.ID == projectId {
			return true
		}
	}

	return false
}

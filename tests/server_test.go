package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	client = initTest()
}

func TestGetServer(t *testing.T) {

	server, err := client.GetServer()
	if err != nil {
		t.Fatalf("Failed to GetServer: %s", err)
	}

	if server == nil {
		t.Fatalf("GetServer did not return a server or serialization failure.")
	}

	assert.Equal(t, int32(2017), server.VersionMajor)
	assert.NotEmpty(t, server.WebURL)
	assert.NotEmpty(t, server.Projects)
	assert.NotEmpty(t, server.VcsRoots)
	assert.NotEmpty(t, server.Builds)
	assert.NotEmpty(t, server.Users)
	assert.NotEmpty(t, server.UserGroups)
	assert.NotEmpty(t, server.Agents)
	assert.NotEmpty(t, server.BuildQueue)
	assert.NotEmpty(t, server.Investigations)
	assert.NotEmpty(t, server.Mutes)
}

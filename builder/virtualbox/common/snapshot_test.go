package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestData() string {
	return `SnapshotName="Imported"
SnapshotUUID="7e5b4165-91ec-4091-a74c-a5709d584530"
SnapshotName-1="Snapshot 1"
SnapshotUUID-1="5fc461ec-da7a-40a8-a168-03134d7cdf5c"
SnapshotName-1-1="Snapshot 2"
SnapshotUUID-1-1="8e12833b-c6b5-4cbd-b42b-09eff8ffc173"
SnapshotName-1-1-1="Snapshot 3"
SnapshotUUID-1-1-1="eb342b39-b4bd-47b0-afd8-dcd1cc5c5929"
SnapshotName-1-1-2="Snapshot 4"
SnapshotUUID-1-1-2="17df1668-e79a-4ed6-a86b-713913699846"
SnapshotName-1-2="Snapshot-Export"
SnapshotUUID-1-2="c857d1b8-4fd6-4044-9d2c-c6e465b3cdd4"
CurrentSnapshotName="Snapshot-Export"
CurrentSnapshotUUID="c857d1b8-4fd6-4044-9d2c-c6e465b3cdd4"
CurrentSnapshotNode="SnapshotName-1-2"
SnapshotName-2="Snapshot 5"
SnapshotUUID-2="85646c6a-fb86-4112-b15e-cab090670778"
SnapshotName-2-1="Snapshot 2"
SnapshotUUID-2-1="7b093686-2981-4ada-8b0f-4c03ae23cd1a"
SnapshotName-3="Snapshot 7"
SnapshotUUID-3="0d977a1f-c9ef-412c-a08d-7c0707b3b18f"
SnapshotName-3-1="Snapshot 8"
SnapshotUUID-3-1="f4ed75b3-afc1-42d4-9e02-8df6f053d07e"
SnapshotName-3-2="Snapshot 9"
SnapshotUUID-3-2="a5903505-9261-4bd3-9972-bacd0064d667"`
}

func TestSnapshot_ParseFullTree(t *testing.T) {
	rootNode, err := ParseSnapshotData(getTestData())
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)
	assert.Equal(t, rootNode.Name, "Imported")
	assert.Equal(t, rootNode.UUID, "7e5b4165-91ec-4091-a74c-a5709d584530")
	assert.Equal(t, 3, len(rootNode.Children))
	assert.Nil(t, rootNode.Parent)
}

func TestSnapshot_FindCurrent(t *testing.T) {
	rootNode, err := ParseSnapshotData(getTestData())
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)

	current := rootNode.GetCurrentSnapshot()
	assert.NotNil(t, current)
	assert.Equal(t, current.UUID, "c857d1b8-4fd6-4044-9d2c-c6e465b3cdd4")
	assert.Equal(t, current.Name, "Snapshot-Export")
	assert.NotNil(t, current.Parent)
	assert.Equal(t, current.Parent.UUID, "5fc461ec-da7a-40a8-a168-03134d7cdf5c")
	assert.Equal(t, current.Parent.Name, "Snapshot 1")
}

func TestSnapshot_FindNodeByUUID(t *testing.T) {
	rootNode, err := ParseSnapshotData(getTestData())
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)

	node := rootNode.GetSnapshotByUUID("7b093686-2981-4ada-8b0f-4c03ae23cd1a")
	assert.NotNil(t, node)
	assert.Equal(t, "Snapshot 2", node.Name)
	assert.Equal(t, "7b093686-2981-4ada-8b0f-4c03ae23cd1a", node.UUID)
	assert.Equal(t, 0, len(node.Children))
	assert.Nil(t, rootNode.Parent)

	otherNode := rootNode.GetSnapshotByUUID("f4ed75b3-afc1-42d4-9e02-8df6f053d07e")
	assert.NotNil(t, otherNode)
	assert.True(t, otherNode.IsChildOf(rootNode))
	assert.False(t, node.IsChildOf(otherNode))
	assert.False(t, otherNode.IsChildOf(node))
}

func TestSnapshot_FindNodesByName(t *testing.T) {
	rootNode, err := ParseSnapshotData(getTestData())
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)

	nodes := rootNode.GetSnapshotsByName("Snapshot 2")
	assert.NotNil(t, nodes)
	assert.Equal(t, 2, len(nodes))
}

func TestSnapshot_IsChildOf(t *testing.T) {
	rootNode, err := ParseSnapshotData(getTestData())
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)

	child := rootNode.GetSnapshotByUUID("c857d1b8-4fd6-4044-9d2c-c6e465b3cdd4")
	assert.NotNil(t, child)
	assert.True(t, child.IsChildOf(rootNode))
	assert.True(t, child.IsChildOf(child.Parent))
	assert.PanicsWithValue(t, "Missing parameter value: candidate", func() { child.IsChildOf(nil) })
}

func TestSnapshot_SingleSnapshot(t *testing.T) {
	snapData := `SnapshotName="Imported"
	SnapshotUUID="7e5b4165-91ec-4091-a74c-a5709d584530"`

	rootNode, err := ParseSnapshotData(snapData)
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)

	assert.Equal(t, rootNode.Name, "Imported")
	assert.Equal(t, rootNode.UUID, "7e5b4165-91ec-4091-a74c-a5709d584530")
	assert.Equal(t, len(rootNode.Children), 0)
	assert.Nil(t, rootNode.Parent)
}

func TestSnapshot_EmptySnapshotData(t *testing.T) {
	snapData := ``

	rootNode, err := ParseSnapshotData(snapData)
	assert.NoError(t, err)
	assert.Nil(t, rootNode)
}

func TestSnapshot_EnsureParents(t *testing.T) {
	rootNode, err := ParseSnapshotData(getTestData())
	assert.NoError(t, err)
	assert.NotNil(t, rootNode)

	for _, snapshot := range rootNode.GetSnapshots() {
		if snapshot == rootNode {
			assert.Nil(t, snapshot.Parent)
		} else {
			assert.NotNil(t, snapshot.Parent)
		}
	}
}

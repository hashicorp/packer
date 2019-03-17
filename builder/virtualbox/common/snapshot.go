package common

import (
	"strings"
)

// VBoxSnapshot stores the hierarchy of snapshots for a VM instance
type VBoxSnapshot struct {
	Name      string
	UUID      string
	IsCurrent bool
	Parent    *VBoxSnapshot // nil if topmost (root) snapshot
	Children  []*VBoxSnapshot
}

// IsChildOf verifies if the current snaphot is a child of the passed as argument
func (sn *VBoxSnapshot) IsChildOf(candidate *VBoxSnapshot) bool {
	if nil == candidate {
		panic("Missing parameter value: candidate")
	}
	node := sn
	for nil != node {
		if candidate.UUID == node.UUID {
			break
		}
		node = node.Parent
	}
	return nil != node
}

// the walker uses a channel to return nodes from a snapshot tree in breadth approach
func walk(sn *VBoxSnapshot, ch chan *VBoxSnapshot) {
	if nil == sn {
		return
	}
	if 0 < len(sn.Children) {
		for _, child := range sn.Children {
			walk(child, ch)
		}
	} else {
		ch <- sn
	}
}

func walker(sn *VBoxSnapshot) <-chan *VBoxSnapshot {
	if nil == sn {
		panic("Argument null exception: sn")
	}

	ch := make(chan *VBoxSnapshot)
	go func() {
		walk(sn, ch)
		close(ch)
	}()
	return ch
}

// GetRoot returns the top-most (root) snapshot for a given snapshot
func (sn *VBoxSnapshot) GetRoot() *VBoxSnapshot {
	if nil == sn {
		panic("Argument null exception: sn")
	}

	node := sn
	for nil != node.Parent {
		node = node.Parent
	}
	return node
}

// GetSnapshotsByName find all snapshots with a given name
func (sn *VBoxSnapshot) GetSnapshotsByName(name string) []*VBoxSnapshot {
	var result []*VBoxSnapshot
	root := sn.GetRoot()
	ch := walker(root)
	for {
		node, ok := <-ch
		if !ok {
			panic("Internal channel error while traversing the snapshot tree")
		}
		if strings.EqualFold(node.Name, name) {
			result = append(result, node)
		}
	}
	return result
}

// GetSnapshotByUUID returns a snapshot by it's UUID
func (sn *VBoxSnapshot) GetSnapshotByUUID(uuid string) *VBoxSnapshot {
	root := sn.GetRoot()
	ch := walker(root)
	for {
		node, ok := <-ch
		if !ok {
			panic("Internal channel error while traversing the snapshot tree")
		}
		if strings.EqualFold(node.UUID, uuid) {
			return node
		}
	}
	return nil
}

// GetCurrentSnapshot returns the currently attached snapshot
func (sn *VBoxSnapshot) GetCurrentSnapshot() *VBoxSnapshot {
	root := sn.GetRoot()
	ch := walker(root)
	for {
		node, ok := <-ch
		if !ok {
			panic("Internal channel error while traversing the snapshot tree")
		}
		if node.IsCurrent {
			return node
		}
	}
	return nil
}

func (sn *VBoxSnapshot) GetChildWithName(name string) *VBoxSnapshot {
	for _, child := range sn.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

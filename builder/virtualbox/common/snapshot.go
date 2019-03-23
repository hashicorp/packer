package common

import (
	"bufio"
	"log"
	"regexp"
	"strings"

	"github.com/golang-collections/collections/stack"
)

// VBoxSnapshot stores the hierarchy of snapshots for a VM instance
type VBoxSnapshot struct {
	Name      string
	UUID      string
	IsCurrent bool
	Parent    *VBoxSnapshot // nil if topmost (root) snapshot
	Children  []*VBoxSnapshot
}

// ParseSnapshotData parses the machinereadable representation of a virtualbox snapshot tree
func ParseSnapshotData(snapshotData string) (*VBoxSnapshot, error) {
	scanner := bufio.NewScanner(strings.NewReader(snapshotData))
	SnapshotNamePartsRe := regexp.MustCompile("Snapshot(?P<Type>Name|UUID)(?P<Path>(-[1-9]+)*)=\"(?P<Value>[^\"]*)\"")
	var currentIndicator string
	parentStack := stack.New()
	var node *VBoxSnapshot
	var rootNode *VBoxSnapshot

	for scanner.Scan() {
		txt := scanner.Text()
		idx := strings.Index(txt, "=")
		if idx > 0 {
			if strings.HasPrefix(txt, "Current") {
				node.IsCurrent = true
			} else {
				matches := SnapshotNamePartsRe.FindStringSubmatch(txt)
				log.Printf("************ Snapshot %s name parts", txt)
				log.Printf("Matches %#v\n", matches)
				log.Printf("Node %s\n", matches[0])
				log.Printf("Type %s\n", matches[1])
				log.Printf("Path %s\n", matches[2])
				log.Printf("Leaf %s\n", matches[3])
				log.Printf("Value %s\n", matches[4])
				if matches[1] == "Name" {
					if nil == rootNode {
						node = new(VBoxSnapshot)
						rootNode = node
						currentIndicator = matches[2]
					} else {
						pathLenCur := strings.Count(currentIndicator, "-")
						pathLen := strings.Count(matches[2], "-")
						if pathLen > pathLenCur {
							currentIndicator = matches[2]
							parentStack.Push(node)
						} else if pathLen < pathLenCur {
							currentIndicator = matches[2]
							for i := 0; i < pathLenCur-1; i++ {
								parentStack.Pop()
							}
						}
						node = new(VBoxSnapshot)
						parent := parentStack.Peek().(*VBoxSnapshot)
						if nil != parent {
							node.Parent = parent
							parent.Children = append(parent.Children, node)
						}
					}
					node.Name = matches[4]
				} else if matches[1] == "UUID" {
					node.UUID = matches[4]
				}
			}
		} else {
			log.Printf("Invalid key,value pair [%s]", txt)
		}
	}
	return rootNode, nil
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
	}
	ch <- sn
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
			break
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
			break
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
			break
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

package main

import (
	"bufio"
	"context"
	"io"
	"regexp"
)

type Tree struct {
	Value    string
	Samples  int64
	Children map[string]*Tree
}

func BuildTree(ctx context.Context, reader io.Reader, separator *regexp.Regexp) (*Tree, error) {
	bufReader := bufio.NewReader(reader)
	tree := Tree{}
	for {
		line, err := bufReader.ReadString('\n')
		if err == io.EOF && line == "" {
			break
		}
		if err != nil && err != io.EOF {
			return nil, err
		}

		if line == "" {
			continue
		}

		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		separatorLocations := separator.FindAllStringSubmatchIndex(line, -1)
		parts := make([]string, len(separatorLocations)+1)
		lineIdx := 0
		for idx, startEnd := range separatorLocations {
			end := startEnd[1]
			parts[idx] = line[lineIdx:end]
			lineIdx = end
		}
		parts[len(parts)-1] = line[lineIdx:]

		tree.Add(parts)
	}
	return &tree, nil
}

func (t *Tree) Add(parts []string) {
	t.Samples++

	if len(parts) == 0 {
		return
	}

	value := parts[0]
	right := parts[1:]

	child, ok := t.Children[value]
	if !ok {
		child = &Tree{Value: value}
		if t.Children == nil {
			t.Children = map[string]*Tree{}
		}
		t.Children[value] = child
	}
	child.Add(right)
}

func (t *Tree) Size() int64 {
	var size int64 = 1
	for _, child := range t.Children {
		size += child.Size()
	}
	return size
}

type FlameGraph struct {
	Name     string        `json:"name"`
	Value    int64         `json:"value"`
	Children []*FlameGraph `json:"children,omitempty"`
}

func (t *Tree) ToFlameGraph() *FlameGraph {
	var children []*FlameGraph
	if len(t.Children) > 0 {
		children = make([]*FlameGraph, len(t.Children))
		idx := 0
		for _, child := range t.Children {
			children[idx] = child.ToFlameGraph()
			idx++
		}
	}

	return &FlameGraph{
		Name:     t.Value,
		Value:    t.Samples,
		Children: children,
	}
}

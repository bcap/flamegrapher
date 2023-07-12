package main

import (
	"bufio"
	"context"
	"io"
	"regexp"
)

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

		if line[len(line)-1] == '\n' {
			line = line[:len(line)-2]
		}

		parts := separator.Split(line, -1)
		tree.Add(parts)
	}
	return &tree, nil
}

type Tree struct {
	Value    string
	Count    int64
	Children map[string]*Tree
}

func (t *Tree) Add(parts []string) {
	if len(parts) == 0 {
		return
	}

	t.Count++

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

type FlameGraph struct {
	Name     string        `json:"name"`
	Value    int64         `json:"value"`
	Children []*FlameGraph `json:"children"`
}

func (t *Tree) ToFlameGraph() *FlameGraph {
	children := make([]*FlameGraph, len(t.Children))
	idx := 0
	for _, child := range t.Children {
		children[idx] = child.ToFlameGraph()
		idx++
	}

	return &FlameGraph{
		Name:     t.Value,
		Value:    t.Count,
		Children: children,
	}
}

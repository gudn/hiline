package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type DocumentPath string

type Document struct {
	ClassName string   `json:"className,omitempty"`
	Group     string   `json:"group"`
	Id        string   `json:"id"`
	Content   string   `json:"content"`
	Start     string   `json:"start"`
	End       string   `json:"end,omitempty"`
	Type      string   `json:"type"`
	Groups    []string `json:"groups"`
}

func (doc *Document) IsValid() bool {
	if doc.Content == "" {
		return false
	}
	if doc.Start == "" {
		return false
	}
	if doc.Type != "box" && doc.Type != "range" {
		return false
	}
	if doc.Type == "range" && doc.End == "" {
		return false
	}
	if doc.Group != "" {
		return false
	}
	return true
}

func (dp DocumentPath) String() string {
	return filepath.Join(*root, string(dp))
}

func (dp DocumentPath) Open() (*os.File, error) {
	return os.Open(dp.String())
}

func (dp DocumentPath) Write(contents []byte) error {
	return os.WriteFile(dp.String(), contents, 0o777)
}

func (dp DocumentPath) Read(group string) (Document, error) {
	f, err := dp.Open()
	if err != nil {
		return Document{}, err
	}
	defer f.Close()

	var doc Document
	err = json.NewDecoder(f).Decode(&doc)
	doc.Id = string(dp)
	doc.Group = group
	return doc, err
}

type Group struct {
	Content      string   `json:"content"`
	Id           string   `json:"id"`
	NestedGroups []string `json:"nestedGroups,omitempty"`

	parent    *Group         `json:"-"`
	Documents []DocumentPath `json:"-"`
}

type Groups struct {
	index map[string]*Group
}

func NewGroups() *Groups {
	return &Groups{index: make(map[string]*Group)}
}

func (g *Groups) Group(group string) *Group {
	if _, ok := g.index[group]; ok {
		return g.index[group]
	}

	colon := strings.LastIndexByte(group, '/')
	var instance *Group
	if colon != -1 {
		parent := g.Group(group[:colon])
		instance = &Group{
			Content: group[colon+1:],
			Id:      group,
			parent:  parent,
		}
	} else {
		instance = &Group{
			Content: group,
			Id:      group,
			parent:  nil,
		}
	}

	g.index[group] = instance
	return instance
}

func (g *Groups) AddDocument(path DocumentPath, doc Document) {
	for _, group := range doc.Groups {
		gg := g.Group(group)
		gg.Documents = append(gg.Documents, path)
	}
}

func (g *Groups) Get(group string) []Group {
	var groups []Group

	for gg := g.index[group]; gg != nil; gg = gg.parent {
		groups = append(groups, *gg)
	}

	return groups
}

func (g *Groups) AllGroups() []string {
	allGroups := make([]string, len(g.index))
	i := 0
	for group := range g.index {
		allGroups[i] = group
		i++
	}

	return allGroups
}

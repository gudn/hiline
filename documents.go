package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type DocumentPath string

type Document struct {
	ClassName string `json:"className,omitempty"`
	Group     string `json:"group"`
	Id        string `json:"id,omitempty"`
	Content   string `json:"content"`
	Start     string `json:"start"`
	End       string `json:"end,omitempty"`
	Type      string `json:"type"`
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
	if doc.Group == "" {
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
	return os.WriteFile(dp.String(), contents, 0o664)
}

func (dp DocumentPath) Read() (Document, error) {
	f, err := dp.Open()
	if err != nil {
		return Document{}, err
	}
	defer f.Close()

	var doc Document
	err = json.NewDecoder(f).Decode(&doc)
	doc.Id = string(dp)
	return doc, err
}

type Group struct {
	Content      string   `json:"content"`
	Id           string   `json:"id"`
	NestedGroups []string `json:"nestedGroups,omitempty"`

	parent    *Group                    `json:"-"`
	Documents map[DocumentPath]struct{} `json:"-"`
}

type Groups struct {
	index        map[string]*Group
	reverseIndex map[DocumentPath]*Group
}

func NewGroups() *Groups {
	return &Groups{index: make(map[string]*Group), reverseIndex: make(map[DocumentPath]*Group)}
}

func SplitGroupName(group string) (parent, content string) {
	slash := strings.LastIndexByte(group, '/')
	if slash != -1 {
		return group[:slash], group[slash+1:]
	}
	return "", group
}

func (g *Groups) Group(group string) *Group {
	if _, ok := g.index[group]; ok {
		return g.index[group]
	}

	parentId, content := SplitGroupName(group)
	var instance *Group
	if parentId != "" {
		parent := g.Group(parentId)
		instance = &Group{
			Content:   content,
			Id:        group,
			parent:    parent,
			Documents: make(map[DocumentPath]struct{}),
		}
	} else {
		instance = &Group{
			Content:   content,
			Id:        group,
			parent:    nil,
			Documents: make(map[DocumentPath]struct{}),
		}
	}

	g.index[group] = instance
	return instance
}

func (g *Groups) AddDocument(path DocumentPath, doc Document) {
	group := g.Group(doc.Group)
	group.Documents[path] = struct{}{}
	g.reverseIndex[path] = group
}

func (g *Groups) DeleteDocument(path DocumentPath) {
	if group := g.reverseIndex[path]; group != nil {
		delete(group.Documents, path)
	}
	delete(g.reverseIndex, path)
}

func (g *Groups) ReplaceDocument(path DocumentPath, doc Document) {
	g.DeleteDocument(path)
	g.AddDocument(path, doc)
}

func (g *Groups) Get(group string) []Group {
	var groups []Group

	for {
		_, ok := g.index[group]
		if ok {
			break
		}

		parent, _ := SplitGroupName(group)
		if parent == "" {
			break
		}
		group = parent
	}

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

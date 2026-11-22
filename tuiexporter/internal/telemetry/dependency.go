package telemetry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ymtdzzz/mermaid-ascii/pkg/drawer"
)

func (m SpanDataMap) getDependencyGraph() (string, error) {
	mermaid := m.getDependencies().getMermaid()
	props, err := drawer.MermaidFileToMap(mermaid, "cli")
	if err != nil {
		return "", err
	}
	return drawer.DrawMap(props), nil
}

func (m SpanDataMap) getDependencies() *dependencyInfo {
	// TODO: should we take an exclusive lock?
	counts := map[string]int{}
	nodeMap := map[string]*node{} // service ID to node map
	for _, span := range m {
		sn, ok := span.ResourceSpan.Resource().Attributes().Get("service.name")
		if !ok {
			continue
		}
		parentspan, ok := m[span.Span.ParentSpanID().String()]
		if !ok {
			getOrCreateNode(nodeMap, sn.AsString())
			continue
		}
		parentsn, ok := parentspan.ResourceSpan.Resource().Attributes().Get("service.name")
		if !ok {
			continue
		}
		if parentsn.AsString() == sn.AsString() {
			continue
		}
		depkey := getDepKey(parentsn.AsString(), sn.AsString())
		counts[depkey]++
		if counts[depkey] > 1 {
			continue
		}
		pn := getOrCreateNode(nodeMap, parentsn.AsString())
		cn := getOrCreateNode(nodeMap, sn.AsString())
		pn.Children = append(pn.Children, cn)
		cn.Parent = pn
	}

	// The nodes that do not have a parent are the head nodes
	heads := []*node{}
	for _, n := range nodeMap {
		if n.Parent == nil {
			n.updateDepth()
			heads = append(heads, n)
		}
	}

	return &dependencyInfo{
		HeadNodes:  heads,
		CallCounts: counts,
	}
}

func getOrCreateNode(nodeMap map[string]*node, service string) *node {
	if n, ok := nodeMap[service]; ok {
		return n
	}
	n := &node{Service: service}
	nodeMap[service] = n
	return n
}

type dependencyInfo struct {
	HeadNodes  []*node
	CallCounts map[string]int
}

func (d *dependencyInfo) getMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph LR\n")

	var traverse func(node *node, tmp string)
	traverse = func(node *node, tmp string) {
		if len(node.Children) == 0 {
			fmt.Fprintf(&sb, "%s\n", tmp)
			return
		}
		for _, child := range node.Children {
			callCount := d.CallCounts[getDepKey(node.Service, child.Service)]
			tmp = fmt.Sprintf("%s -->|%d| %s", tmp, callCount, child.Service)

			traverse(child, tmp)

			// Start new line
			tmp = node.Service
		}
	}

	for _, head := range d.HeadNodes {
		traverse(head, head.Service)
	}

	return getSortedMermaid(sb.String())
}

type node struct {
	Service  string
	Parent   *node
	Children []*node
	Depth    int
}

func (n *node) updateDepth() int {
	if n.Depth != 0 {
		// If depth is already calculated, return it
		return n.Depth
	}

	if len(n.Children) == 0 {
		n.Depth = 1
	} else {
		maxDepth := 0
		for _, cn := range n.Children {
			childDepth := cn.updateDepth()
			if childDepth > maxDepth {
				maxDepth = childDepth
			}
		}
		n.Depth = maxDepth + 1
	}

	return n.Depth
}

func getSortedMermaid(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) <= 1 {
		return input
	}

	header := lines[0]
	lines = lines[1:]

	sort.Slice(lines, func(i, j int) bool {
		return strings.Count(lines[i], "-->") > strings.Count(lines[j], "-->")
	})

	output := append([]string{header}, lines...)

	return strings.Join(output, "\n")
}

func getDepKey(parentsn, childsn string) string {
	return parentsn + "&&&" + childsn
}

package main

import (
	"bufio"
	"bytes"
	"container/list"
	"encoding/hex"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Cluster struct {
	Subgraph *Graph
	Parent   *Cluster
	Children []*Cluster
}

type GraphOfGraphs struct {
	RootCluster *Cluster
}

type Graph struct {
	Name         string
	Nodes        map[string]*Node
	Edges        map[string]map[string]*Edge
	ReverseEdges map[string]map[string]bool
}

type Node struct {
	Color     string
	FillColor string
	Id        string
	Subsystem string
	Label     string
	Shape     string
	NodeInfo  ExtraInfo
	Complete  bool
}

type Edge struct {
	Color string
}

type ExtraInfo interface{}

type KconfigSymbol struct {
	Name        string
	Type        string
	Default     string
	Prompt      string
	Description string
}

type KconfigExpr struct {
	Op      string    // Operator: "!", "||", "&&" (empty for operands)
	Left    *KconfigExpr
	Right   *KconfigExpr
	Operand *Node
}

const (
	DefaultColor string = "black"
	HighlighColor string = "green"
	Merge1Color string = "orange"
	Merge2Color string = "blue"
	MergeCommonColor string = "purple"
	ExprColor string = "red"
	OpColor string = "red"
	DepColor string = "orange"
	OpShape string = "diamond"
	IfShape string = "triangle"
	IfColor string = "grey"
)

var debug bool = false

const (
	PARSE_PLAIN int = iota
	PARSE_SUBSYS
)
const (
	PART_NONE int = iota
	LOUVAIN_SINGLE
	LOUVAIN_PARALLEL
	CHEAT_CLUSTERING
)
const (
	O_INVALID int = iota
	O_HORIZONTAL
	O_VERTICAL
)

func NewGraph() *Graph {
	return &Graph{
		Nodes:        make(map[string]*Node),
		Edges:        make(map[string]map[string]*Edge),
		ReverseEdges: make(map[string]map[string]bool),
	}
}

func (g *Graph) AddNode(id string, node *Node) {
	if _, exists := g.Nodes[id]; !exists {
		g.Nodes[id] = node
	}
}

func (g *Graph) SetNodeColor(id, color string) {
	if node, exists := g.Nodes[id]; exists {
		node.Color = color
	}
}

func (g *Graph) AddEdge(from, to string, e *Edge) {
	if _, exists := g.Edges[from]; !exists {
		g.Edges[from] = make(map[string]*Edge)
	}
	g.Edges[from][to] = e

	if _, exists := g.ReverseEdges[to]; !exists {
		g.ReverseEdges[to] = make(map[string]bool)
	}
	g.ReverseEdges[to][from] = true
}

func (g *Graph) findExclusivelyReachable(root string) (map[string]int, []string) {
	visited := make(map[string]bool)
	queue := list.New()
	queue.PushBack(root)
	visited[root] = true
	distances := make(map[string]int)
	distances[root] = 0
	exclusivelyReachable := []string{}

	for queue.Len() > 0 {
		node := queue.Remove(queue.Front()).(string)
		for neighbor := range g.Edges[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue.PushBack(neighbor)
				distances[neighbor] = distances[node] + 1
			}
		}
	}

	for node := range visited {
		if node != root {
			exclusivelyReachable = append(exclusivelyReachable, node)
		}
	}

	sortByDistance(exclusivelyReachable, distances)

	return distances, exclusivelyReachable
}

func sortByDistance(nodes []string, distances map[string]int) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return distances[nodes[i]] > distances[nodes[j]]
	})
}

func (g *Graph) OnlyReachableNodes(start, intermediate string) []string {

	allReachable := g.traverseFrom(start)
	outgoing, incoming := g.removeNodeEdges(intermediate)
	reachableWithoutIntermediate := g.traverseFrom(start)
	g.restoreNodeEdges(intermediate, outgoing, incoming)
	onlyReachable := []string{}
	distances := g.computeDistances(intermediate)
	for node := range allReachable {
		if node != intermediate {
			if _, stillReachable := reachableWithoutIntermediate[node]; !stillReachable {
				onlyReachable = append(onlyReachable, node)
			}
		}
	}

	sort.Slice(onlyReachable, func(i, j int) bool {
		return distances[onlyReachable[i]] < distances[onlyReachable[j]]
	})

	return onlyReachable
}

func (g *Graph) computeDistances(start string) map[string]int {
	distances := make(map[string]int)
	queue := []string{start}
	distances[start] = 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for neighbor := range g.Edges[current] {
			if _, visited := distances[neighbor]; !visited {
				distances[neighbor] = distances[current] + 1
				queue = append(queue, neighbor)
			}
		}
	}

	return distances
}

func (g *Graph) traverseFrom(start string) map[string]bool {
	visited := make(map[string]bool)
	var dfs func(node string)
	dfs = func(node string) {
		if visited[node] {
			return
		}
		visited[node] = true
		for neighbor := range g.Edges[node] {
			dfs(neighbor)
		}
	}
	dfs(start)
	return visited
}

func (g *Graph) removeNodeEdges(node string) (outgoing map[string]*Edge, incoming map[string]*Edge) {
	outgoing = g.Edges[node]

	incoming = make(map[string]*Edge)
	for parent := range g.ReverseEdges[node] {
		incoming[parent] = g.Edges[parent][node]
		delete(g.Edges[parent], node)
	}

	g.Edges[node] = nil
	g.ReverseEdges[node] = nil
	return outgoing, incoming
}

func (g *Graph) restoreNodeEdges(node string, outgoing map[string]*Edge, incoming map[string]*Edge) {
	g.Edges[node] = outgoing

	for parent, color := range incoming {
		g.Edges[parent][node] = color
		if g.ReverseEdges[node] == nil {
			g.ReverseEdges[node] = make(map[string]bool)
		}
		g.ReverseEdges[node][parent] = true
	}
}

func (g *Graph) RemoveEdge(from, to string) {
	delete(g.Edges[from], to)
	delete(g.ReverseEdges[to], from)
}

func (g *Graph) collapseNode(first, second string) {
	if _, exists := g.Edges[first][second]; exists {
		g.RemoveEdge(first, second)
	}

	for neighbor, color := range g.Edges[second] {
		if _, backEdgeExists := g.Edges[neighbor][second]; backEdgeExists {
			g.RemoveEdge(neighbor, second)
			g.AddEdge(neighbor, first, color)
		}

		if first != neighbor {
			g.AddEdge(first, neighbor, color)
		}

		delete(g.ReverseEdges[neighbor], second)
	}

	for parent := range g.ReverseEdges[second] {
		color := g.Edges[parent][second]
		g.AddEdge(parent, first, color)
	}

	delete(g.ReverseEdges, second)
	delete(g.Nodes, second)
	delete(g.Edges, second)
}

func RemoveNodesDC(g *Graph, nodesToProcess []string) bool {
	toProcessSet := make(map[string]bool)
	for _, node := range nodesToProcess {
		toProcessSet[node] = true
	}

	removed := make(map[string]bool)
	var markToRemove func(node string)
	markToRemove = func(node string) {
		if removed[node] {
			return
		}
		removed[node] = true

		for neighbor := range g.Edges[node] {
			if len(g.ReverseEdges[neighbor]) == 1 && g.ReverseEdges[neighbor][node] {
				markToRemove(neighbor)
			}
		}
	}

	for _, node := range nodesToProcess {
		markToRemove(node)
	}

	return true
}

func (g *Graph) DetachAndClean(start, target string) {

	g.Nodes[target].Color = HighlighColor
	removedEdges := g.Edges[target]
	g.Edges[target] = nil
	reachable := g.traverseFrom(start)

	for node := range g.Nodes {
		if !reachable[node] {
			if debug {
				fmt.Printf("removing %s\n", node)
			}
			g.removeNode(node)
		}
	}

	for neighbor := range removedEdges {
		delete(g.ReverseEdges[neighbor], target)
	}
}
func (g *Graph) removeNode(node string) {
	for parent := range g.ReverseEdges[node] {
		delete(g.Edges[parent], node)
	}
	delete(g.Edges, node)
	delete(g.ReverseEdges, node)
	delete(g.Nodes, node)
}

func ParseDOT(filename string, parseType int) (*Graph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s: %v", filename, err)
	}
	defer file.Close()
	debugIOPrintf("input file exists!\n")

	graph := NewGraph()
	graphRegex := regexp.MustCompile(`^\s*digraph\s+[a-zA-Z_\-0-9]+\s*\{\s*$`)
	nodeRegex := regexp.MustCompile(`^\s*"*([^"]+[^\s\"])"*\s*\[([^\]]*)\]\s*;`)
	edgeRegex := regexp.MustCompile(`^\s*"*([^"\[\]]+[^\s\"])"*\s*->\s*"*([^"\[\];][^\s\"\;]*)"*\s*(\[([^\]]*)\]){0,1}\s*;{0,1}`)
	attrRegex := regexp.MustCompile(`([0-9a-zA-Z]+[^\s])\s*=\s*"*([^,\"]+)"*`)

	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if line == "" || line == "}" {
			continue
		}

		if lineNum == 1 {
			if !graphRegex.MatchString(line) {
				return nil, errors.New("invalid DOT format: missing or malformed graph declaration")
			}
			debugIOPrintf("dot header found\n")
			continue
		}

		if match := nodeRegex.FindStringSubmatch(line); match != nil {
			debugIOPrintf("node line found @%d\n", lineNum)
			id := match[1]
			attrString := match[2]

			node := &Node{Id: id, Color: DefaultColor}
			for _, attr := range attrRegex.FindAllStringSubmatch(attrString, -1) {
				key, value := attr[1], attr[2]
				switch key {
				case "color":
					debugIOPrintf("match color=%s\n", value)
					node.Color = value
				case "label":
					debugIOPrintf("match label=%s\n", value)
					node.Label = value
				case "fillcolor":
					debugIOPrintf("match fillcolor=%s\n", value)
					node.FillColor = value
				case "shape":
					debugIOPrintf("match shape=%s\n", value)
					node.Shape = value
				}

			}
			if parseType == PARSE_PLAIN {
				graph.Nodes[id] = node
			} else {
				nodeData := strings.Split(id, "|")
				if len(nodeData) != 2 {
					return nil, errors.New(fmt.Sprintf("invalid node syntax at %d", lineNum))
				}
				node.Subsystem = nodeData[0]
				graph.Nodes[nodeData[1]] = node
			}
			continue
		}

		if match := edgeRegex.FindStringSubmatch(line); match != nil {
			debugIOPrintf("edge line found @%d\n", lineNum)
			src_r := match[1]
			dst_r := match[2]
			attrString := match[3]
			src := src_r
			dst := dst_r
			srcSubsys := "UNKNOWN"
			dstSubsys := "UNKNOWN"
			if parseType != PARSE_PLAIN {
				srcNodeData := strings.Split(src_r, "|")
				if len(srcNodeData) != 2 {
					return nil, errors.New(fmt.Sprintf("invalid src node syntax at %d", lineNum))
				}
				dstNodeData := strings.Split(dst_r, "|")
				if len(dstNodeData) != 2 {
					return nil, errors.New(fmt.Sprintf("invalid dst node syntax at %d", lineNum))
				}
				src = srcNodeData[1]
				dst = dstNodeData[1]
				srcSubsys = srcNodeData[0]
				dstSubsys = dstNodeData[0]
			}

			if _, ok := graph.Nodes[src]; !ok {
				graph.Nodes[src] = &Node{Id: src, Subsystem: srcSubsys}
			}
			if _, ok := graph.Nodes[dst]; !ok {
				graph.Nodes[dst] = &Node{Id: dst, Subsystem: dstSubsys}
			}

			edge := &Edge{Color: DefaultColor}
			for _, attr := range attrRegex.FindAllStringSubmatch(attrString, -1) {
				key, value := attr[1], attr[2]
				if key == "color" {
					edge.Color = value
				}
			}

			if graph.Edges[src] == nil {
				graph.Edges[src] = make(map[string]*Edge)
			}
			graph.Edges[src][dst] = edge

			if graph.ReverseEdges[dst] == nil {
				graph.ReverseEdges[dst] = make(map[string]bool)
			}
			graph.ReverseEdges[dst][src] = true
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}
	debugIOPrintf("Parsed lines = %d\n", lineNum)
	return graph, nil
}

func (g *Graph) ToDot(name string, orientation int) string {
	return g.ToDotRaw(name, orientation, false)
}

func (g *Graph) ToDotRaw(name string, orientation int, markRoots bool) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("digraph %s {\n", name))
	if orientation == O_HORIZONTAL {
		buf.WriteString("rankdir=LR;\n")
	}

	noParentNodes := make(map[string]bool)
	if markRoots {
		for source, edges := range g.Edges {
			noParentNodes[source] = true
			if _, ok := g.Nodes[source]; !ok {
				g.Nodes[source] = &Node{Color: DefaultColor}
			}
			for target := range edges {
				noParentNodes[target] = true
				if _, ok := g.Nodes[target]; !ok {
					g.Nodes[target] = &Node{Color: DefaultColor}
				}
			}
		}
		for _, edges := range g.Edges {
			for target := range edges {
				delete(noParentNodes, target)
			}
		}
	}

	for id, node := range g.Nodes {
		buf.WriteString(fmt.Sprintf("    %q [", id))
		c := DefaultColor
		if markRoots && noParentNodes[id] {
			c = HighlighColor
		} else if node.Color != "" {
			c = node.Color
		}
		buf.WriteString(fmt.Sprintf("color=%q", c))

		if node.FillColor != "" {
			buf.WriteString(fmt.Sprintf(", style=\"filled\", fillcolor=%q", node.FillColor))
		}
		if node.Shape != "" {
			buf.WriteString(fmt.Sprintf(", shape=%q", node.Shape))
		}
		if node.Label != "" {
			buf.WriteString(fmt.Sprintf(", label=%q", node.Label))
		}

		buf.WriteString("];\n")
	}

	for from, edges := range g.Edges {
		for to, edge := range edges {
			buf.WriteString(fmt.Sprintf("    %q -> %q", from, to))
			if edge.Color != "" {
				buf.WriteString(fmt.Sprintf(" [color=%q]", edge.Color))
			}
			buf.WriteString(";\n")
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}

func MergeWithHighlight(g1, g2 *Graph) *Graph {
	merged := NewGraph()

	for node, _ := range g1.Nodes {
		if _, exists := g2.Nodes[node]; exists {
			merged.AddNode(node, &Node{Color: MergeCommonColor})
		} else {
			merged.AddNode(node, &Node{Color: Merge1Color})
		}
	}
	for from, edges := range g1.Edges {
		for to := range edges {
			if _, exists := g2.Edges[from][to]; exists {
				merged.AddEdge(from, to, &Edge{Color: MergeCommonColor})
			} else {
				merged.AddEdge(from, to, &Edge{Color: Merge1Color})
			}
		}
	}

	for node, _ := range g2.Nodes {
		if _, exists := g1.Nodes[node]; !exists {
			merged.AddNode(node, &Node{Color: Merge2Color})
		}
	}
	for from, edges := range g2.Edges {
		for to := range edges {
			if _, exists := g1.Edges[from][to]; !exists {
				merged.AddEdge(from, to, &Edge{Color: Merge2Color})
			}
		}
	}

	return merged
}

func (g *Graph) RemovePathsFrom(roots []string) {
	visited := make(map[string]bool)
	for _, root := range roots {
		g.removePathsDFS(root, visited)
	}
}

func (g *Graph) removePathsDFS(node string, visited map[string]bool) {
	if visited[node] {
		return
	}
	visited[node] = true

	if _, exists := g.Edges[node]; exists {
		for child := range g.Edges[node] {
			g.removePathsDFS(child, visited)
		}
		delete(g.Edges, node)
	}
}

func (g *Graph) HasNoParent(node string) bool {
	if g.ReverseEdges[node] != nil && len(g.ReverseEdges[node])>0 {
		return false
	}
	return true
}

func (g *Graph) HasNoChild(node string) bool {
	_, exists := g.Edges[node]
	return !exists || len(g.Edges[node]) == 0
}

func ImportFtrace(filename string) (*Graph, error) {

	graph := NewGraph()
	lineRegex := regexp.MustCompile(`([a-zA-Z0-9_.]+) <-([a-zA-Z0-9_.]+)`)

	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error opening file: %v", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := lineRegex.FindStringSubmatch(line); matches != nil {
			caller := matches[2]
			callee := matches[1]
			graph.AddEdge(caller, callee, &Edge{Color: DefaultColor})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading file: %v\n", err))
	}
	return graph, nil

}

func ConnectGraphs(graph1, graph2 *Graph, srcNode, dstNode, edgeColor string) (*Graph, error) {
	if _, exists := graph1.Nodes[srcNode]; !exists {
		return nil, fmt.Errorf("source node %q does not exist in the first graph", srcNode)
	}

	if _, exists := graph2.Nodes[dstNode]; !exists {
		return nil, fmt.Errorf("destination node %q does not exist in the second graph", dstNode)
	}

	newGraph := NewGraph()
	for id, node := range graph1.Nodes {
		newGraph.Nodes[id] = &Node{Id: node.Id, Color: node.Color}
	}
	for src, edges := range graph1.Edges {
		newGraph.Edges[src] = make(map[string]*Edge)
		for dst, edge := range edges {
			newGraph.Edges[src][dst] = &Edge{Color: edge.Color}
		}
	}
	for dst, reverseEdges := range graph1.ReverseEdges {
		newGraph.ReverseEdges[dst] = make(map[string]bool)
		for src := range reverseEdges {
			newGraph.ReverseEdges[dst][src] = true
		}
	}

	for id, node := range graph2.Nodes {
		if _, exists := newGraph.Nodes[id]; !exists {
			newGraph.Nodes[id] = &Node{Id: node.Id, Color: node.Color}
		}
	}
	for src, edges := range graph2.Edges {
		if newGraph.Edges[src] == nil {
			newGraph.Edges[src] = make(map[string]*Edge)
		}
		for dst, edge := range edges {
			if _, exists := newGraph.Edges[src][dst]; !exists {
				newGraph.Edges[src][dst] = &Edge{Color: edge.Color}
			}
		}
	}
	for dst, reverseEdges := range graph2.ReverseEdges {
		if newGraph.ReverseEdges[dst] == nil {
			newGraph.ReverseEdges[dst] = make(map[string]bool)
		}
		for src := range reverseEdges {
			newGraph.ReverseEdges[dst][src] = true
		}
	}

	if newGraph.Edges[srcNode] == nil {
		newGraph.Edges[srcNode] = make(map[string]*Edge)
	}
	newGraph.Edges[srcNode][dstNode] = &Edge{
		Color: edgeColor,
	}

	if newGraph.ReverseEdges[dstNode] == nil {
		newGraph.ReverseEdges[dstNode] = make(map[string]bool)
	}
	newGraph.ReverseEdges[dstNode][srcNode] = true

	return newGraph, nil
}
func (g *Graph) NodeExists(nodeID string) bool {
	_, exists := g.Nodes[nodeID]
	return exists
}
func (g *Graph) HasNode(nodeName string) bool {
	_, exists := g.Nodes[nodeName]
	return exists
}
func (g *Graph) Subgraph(nodeName string) (*Graph, error) {
	if _, exists := g.Nodes[nodeName]; !exists {
		return nil, fmt.Errorf("node %q does not exist in the graph", nodeName)
	}

	subgraph := NewGraph()

	visited := make(map[string]bool)
	queue := []string{nodeName}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		subgraph.Nodes[current] = &Node{
			Id:    g.Nodes[current].Id,
			Color: g.Nodes[current].Color,
		}

		if g.Edges[current] != nil {
			subgraph.Edges[current] = make(map[string]*Edge)
			for neighbor, edge := range g.Edges[current] {
				subgraph.Edges[current][neighbor] = &Edge{
					Color: edge.Color,
				}
				if !visited[neighbor] {
					queue = append(queue, neighbor)
				}
			}
		}
	}

	for src, edges := range subgraph.Edges {
		for dst := range edges {
			if subgraph.ReverseEdges[dst] == nil {
				subgraph.ReverseEdges[dst] = make(map[string]bool)
			}
			subgraph.ReverseEdges[dst][src] = true
		}
	}

	return subgraph, nil
}
func (g *Graph) FindPath(startNode, endNode string) ([]string, error) {
	if _, exists := g.Nodes[startNode]; !exists {
		return nil, fmt.Errorf("start node %q does not exist in the graph", startNode)
	}
	if _, exists := g.Nodes[endNode]; !exists {
		return nil, fmt.Errorf("end node %q does not exist in the graph", endNode)
	}

	queue := [][]string{{startNode}}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		current := path[len(path)-1]
		if current == endNode {
			return path, nil
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		for neighbor := range g.Edges[current] {
			if !visited[neighbor] {
				newPath := append([]string{}, path...)
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}

	return nil, fmt.Errorf("no path exists between %q and %q", startNode, endNode)
}

func (g *Graph) SubKernelreplace(startNode, newNodeID string) (*Graph, error) {
	if _, exists := g.Nodes[startNode]; !exists {
		return nil, fmt.Errorf("node %q does not exist in the graph", startNode)
	}

	visited := make(map[string]bool)
	queue := []string{startNode}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		for neighbor := range g.Edges[current] {
			if !visited[neighbor] {
				queue = append(queue, neighbor)
			}
		}
	}

	newGraph := NewGraph()

	for nodeID, node := range g.Nodes {
		if !visited[nodeID] || nodeID == startNode {
			newGraph.Nodes[nodeID] = &Node{Id: node.Id, Color: node.Color}
		}
	}

	newGraph.Nodes[newNodeID] = &Node{Id: newNodeID, Color: HighlighColor}

	for src, edges := range g.Edges {
		for dst, edge := range edges {
			if visited[src] && visited[dst] {
				continue
			}

			if visited[dst] {
				newGraph.AddEdge(src, newNodeID, edge)
			} else if visited[src] {
				newGraph.AddEdge(newNodeID, dst, edge)
			} else {
				newGraph.AddEdge(src, dst, edge)
			}
		}
	}

	return newGraph, nil
}

func (g *Graph) CollapseKernel(newNodeID string) (*Graph, error) {
	sccs := findSCCs(g)

	var kernelNodes []string
	for _, scc := range sccs {
		if len(scc) > len(kernelNodes) {
			kernelNodes = scc
		}
	}

	newGraph := NewGraph()
	kernelSet := make(map[string]bool)
	for _, node := range kernelNodes {
		kernelSet[node] = true
	}

	for nodeID, node := range g.Nodes {
		if !kernelSet[nodeID] {
			newGraph.Nodes[nodeID] = &Node{Id: node.Id, Color: node.Color}
		}
	}

	newGraph.Nodes[newNodeID] = &Node{Id: newNodeID, Color: "kernel"}

	for src, edges := range g.Edges {
		for dst, edge := range edges {
			if kernelSet[src] && kernelSet[dst] {
				continue
			} else if kernelSet[src] {
				newGraph.AddEdge(newNodeID, dst, edge)
			} else if kernelSet[dst] {
				newGraph.AddEdge(src, newNodeID, edge)
			} else {

				newGraph.AddEdge(src, dst, edge)
			}
		}
	}

	return newGraph, nil
}

func findSCCs(g *Graph) [][]string {
	var (
		index      = 0
		stack      []string
		indices    = make(map[string]int)
		lowLinks   = make(map[string]int)
		onStack    = make(map[string]bool)
		components [][]string
	)

	var strongConnect func(v string)
	strongConnect = func(v string) {
		indices[v] = index
		lowLinks[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for neighbor := range g.Edges[v] {
			if _, visited := indices[neighbor]; !visited {
				strongConnect(neighbor)
				lowLinks[v] = min(lowLinks[v], lowLinks[neighbor])
			} else if onStack[neighbor] {
				lowLinks[v] = min(lowLinks[v], indices[neighbor])
			}
		}

		if lowLinks[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			components = append(components, scc)
		}
	}

	for node := range g.Nodes {
		if _, visited := indices[node]; !visited {
			strongConnect(node)
		}
	}

	return components
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *Graph) HierarchicalClustering(clusteringType int) (*GraphOfGraphs, error) {
	rootCluster := &Cluster{
		Subgraph: g,
		Parent:   nil,
		Children: nil,
	}

	graphOfGraphs := &GraphOfGraphs{
		RootCluster: rootCluster,
	}

	err := recursivelyCluster(rootCluster, clusteringType)
	if err != nil {
		return nil, err
	}

	return graphOfGraphs, nil
}

func recursivelyCluster(cluster *Cluster, clusteringType int) error {
	var err error
	var nodeClusters [][]string
	switch clusteringType {
	case LOUVAIN_PARALLEL:
		nodeClusters, err = LouvainClusteringWithMinSize_P(cluster.Subgraph, 3)
	case LOUVAIN_SINGLE:
		nodeClusters, err = LouvainClusteringWithMinSize(cluster.Subgraph, 3)
	case CHEAT_CLUSTERING:
		nodeClusters, err = SubsystemBasedClustering(cluster.Subgraph)
	default:
		return fmt.Errorf("clustering failed: Unknown clusteringType")
	}
	if err != nil {
		return fmt.Errorf("clustering failed: %v", err)
	}

	if len(nodeClusters) <= 1 {
		return nil
	}

	for _, subgraphNodes := range nodeClusters {
		subgraph := extractSubgraph(cluster.Subgraph, subgraphNodes)
		childCluster := &Cluster{
			Subgraph: subgraph,
			Parent:   cluster,
			Children: nil,
		}
		cluster.Children = append(cluster.Children, childCluster)

		if err := recursivelyCluster(childCluster, clusteringType); err != nil {
			return err
		}
	}

	return nil
}

func extractSubgraph(g *Graph, nodes []string) *Graph {
	subgraph := NewGraph()
	for _, node := range nodes {
		if originalNode, exists := g.Nodes[node]; exists {
			subgraph.AddNode(originalNode.Id, originalNode)
		}
	}

	for _, node := range nodes {
		if edges, exists := g.Edges[node]; exists {
			for target := range edges {
				if subgraph.HasNode(target) {
					subgraph.AddEdge(node, target, g.Edges[node][target])
				}
			}
		}
	}

	return subgraph
}
func LouvainClustering(g *Graph) ([][]string, error) {
	if g == nil || len(g.Nodes) == 0 {
		return nil, fmt.Errorf("graph is empty")
	}

	clusters := make(map[string]string)
	for node := range g.Nodes {
		clusters[node] = node
	}

	modularityGain := true
	for modularityGain {
		modularityGain = false
		for node := range g.Nodes {
			bestCluster := clusters[node]
			bestDeltaQ := 0.0

			for neighbor := range g.Edges[node] {
				currentCluster := clusters[node]
				targetCluster := clusters[neighbor]

				if currentCluster != targetCluster {
					deltaQ := calculateModularityGain(g, clusters, node, targetCluster)
					if deltaQ > bestDeltaQ {
						bestDeltaQ = deltaQ
						bestCluster = targetCluster
					}
				}
			}

			if clusters[node] != bestCluster {
				clusters[node] = bestCluster
				modularityGain = true
			}
		}

		if modularityGain {
			g = aggregateCommunities(g, clusters)
		}
	}

	finalClusters := make(map[string][]string)
	for node, cluster := range clusters {
		finalClusters[cluster] = append(finalClusters[cluster], node)
	}

	result := make([][]string, 0, len(finalClusters))
	for _, nodes := range finalClusters {
		result = append(result, nodes)
	}

	return result, nil
}
func calculateModularityGain(g *Graph, clusters map[string]string, node, targetCluster string) float64 {
	totalEdgeWeight := float64(0)
	for _, edges := range g.Edges {
		totalEdgeWeight += float64(len(edges))
	}

	nodeDegree := float64(len(g.Edges[node]))
	clusterDegree := float64(0)
	for n, c := range clusters {
		if c == targetCluster {
			clusterDegree += float64(len(g.Edges[n]))
		}
	}

	intraClusterEdges := 0.0
	for neighbor := range g.Edges[node] {
		if clusters[neighbor] == targetCluster {
			intraClusterEdges += 1.0
		}
	}

	deltaQ := (intraClusterEdges / totalEdgeWeight) - (nodeDegree*clusterDegree)/(totalEdgeWeight*totalEdgeWeight)
	return deltaQ
}
func aggregateCommunities(g *Graph, clusters map[string]string) *Graph {
	newGraph := NewGraph()
	clusterMapping := make(map[string]string)

	for _, cluster := range clusters {
		if _, exists := clusterMapping[cluster]; !exists {
			clusterMapping[cluster] = cluster
			newGraph.AddNode(cluster, &Node{Color: DefaultColor})
		}
	}

	for source, targets := range g.Edges {
		for target := range targets {
			srcCluster := clusters[source]
			tgtCluster := clusters[target]
			if srcCluster != tgtCluster {
				if _, exists := newGraph.Edges[srcCluster]; !exists {
					newGraph.Edges[srcCluster] = make(map[string]*Edge)
				}
				newGraph.Edges[srcCluster][tgtCluster] = &Edge{Color: DefaultColor}
			}
		}
	}

	return newGraph
}

func getClusterForNode(rootCluster *Cluster, node string) *Cluster {
	if rootCluster == nil || rootCluster.Subgraph == nil {
		return nil
	}

	if rootCluster.Subgraph.HasNode(node) {
		return rootCluster
	}

	for _, child := range rootCluster.Children {
		result := getClusterForNode(child, node)
		if result != nil {
			return result
		}
	}

	return nil
}
func getRandomElement(m map[string]*Node) (*Node, bool) {
	for _, value := range m {
		return value, true
	}
	return nil, false
}

func writeClusterToDOT(builder *strings.Builder, cluster *Cluster, depth int, gog *GraphOfGraphs) error {
	var color_r, color_g, color_b byte

	if cluster == nil {
		return nil
	}

	cluster_name := "main"
	color_r = 0xff
	color_g = 0xff
	color_b = 0xff
	if depth != 0 {
		node, _ := getRandomElement(cluster.Subgraph.Nodes)
		cluster_name = node.Subsystem
		color_r, color_g, color_b = getMd5Parts(cluster_name)

	}
	indent := strings.Repeat("    ", depth)
	builder.WriteString(fmt.Sprintf("%ssubgraph cluster_%p {\n", indent, cluster))
	builder.WriteString(fmt.Sprintf("%s    label=\"%s\";\n", indent, cluster_name))
	builder.WriteString(fmt.Sprintf("%s    color=\"#%02x%02x%02x\";\n", indent, color_r, color_g, color_b))
	builder.WriteString(fmt.Sprintf("%s    fillcolor=lightgrey;\n", indent))
	builder.WriteString(fmt.Sprintf("%s    fontsize=40;\n", indent))
	builder.WriteString(fmt.Sprintf("%s    penwidth=10;\n", indent))

	for nodeName, node := range cluster.Subgraph.Nodes {
		ncolor := DefaultColor
		if node.Color != "" {
			ncolor = node.Color
		}
		builder.WriteString(fmt.Sprintf("%s    %q [color=%q];\n", indent, nodeName, ncolor))
	}

	for src, targets := range cluster.Subgraph.Edges {
		for tgt := range targets {
			if cluster.Subgraph.HasNode(tgt) {
				fmtStr := "%s    %q -> %q;\n"
				if depth == 0 {
					fmtStr = fmt.Sprintf("%%s    %%q -> %%q [color=%s; penwidth=5];\n", HighlighColor)
				}
				if depth != 0 || cluster.Subgraph.Nodes[src].Subsystem != cluster.Subgraph.Nodes[tgt].Subsystem {
					builder.WriteString(fmt.Sprintf(fmtStr, indent, src, tgt))
				}
			}
		}
	}

	for _, child := range cluster.Children {
		err := writeClusterToDOT(builder, child, depth+1, gog)
		if err != nil {
			return err
		}
	}

	builder.WriteString(fmt.Sprintf("%s}\n", indent))

	if depth == 0 {
		writtenEdges := make(map[string]struct{})
		for src, targets := range gog.RootCluster.Subgraph.Edges {
			for tgt, edge := range targets {
				srcCluster := getClusterForNode(gog.RootCluster, src)
				tgtCluster := getClusterForNode(gog.RootCluster, tgt)

				if srcCluster != tgtCluster {
					edgeKey := fmt.Sprintf("%s->%s", src, tgt)
					if _, written := writtenEdges[edgeKey]; !written {
						if edge != nil && edge.Color != "" {
							builder.WriteString(fmt.Sprintf("%s%q -> %q [color=%q];\n", indent, src, tgt, edge.Color))
						} else {
							builder.WriteString(fmt.Sprintf("%s%q -> %q;\n", indent, src, tgt))
						}
						writtenEdges[edgeKey] = struct{}{}
					}
				}
			}
		}
	}

	return nil
}
func (gog *GraphOfGraphs) ToDot(name string, orientation int) (string, error) {
	if gog == nil || gog.RootCluster == nil {
		return "", fmt.Errorf("GraphOfGraphs is empty")
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("digraph %s {\n", name))
	if orientation == O_HORIZONTAL {
		builder.WriteString("rankdir=LR;\n")
	}

	err := writeClusterToDOT(&builder, gog.RootCluster, 0, gog)
	if err != nil {
		return "", err
	}

	builder.WriteString(fmt.Sprintf("} # %s\n", name))
	return builder.String(), nil
}

func convertClustersToGraphOfGraphs(graph *Graph, clusters map[string][]string) *GraphOfGraphs {
	gog := &GraphOfGraphs{
		RootCluster: &Cluster{
			Subgraph: NewGraph(),
		},
	}

	for _, nodes := range clusters {
		clusterGraph := NewGraph()
		for _, node := range nodes {
			clusterGraph.AddNode(node, graph.Nodes[node])
		}

		for _, src := range nodes {
			for tgt := range graph.Edges[src] {
				if clusterGraph.HasNode(tgt) {
					clusterGraph.AddEdge(src, tgt, graph.Edges[src][tgt])
				}
			}
		}

		gog.RootCluster.Children = append(gog.RootCluster.Children, &Cluster{Subgraph: clusterGraph})
	}

	return gog
}

func LouvainClusteringWithMinSize(graph *Graph, minClusterSize int) ([][]string, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph cannot be nil")
	}

	nodeToCluster := make(map[string]string)
	for node := range graph.Nodes {
		nodeToCluster[node] = node
	}

	improved := true
	for improved {
		improved = false
		for node := range graph.Nodes {
			bestCluster := nodeToCluster[node]
			bestGain := 0.0

			for neighbor := range graph.Edges[node] {
				targetCluster := nodeToCluster[neighbor]
				gain := calculateModularityGain(graph, nodeToCluster, node, targetCluster)
				if gain > bestGain {
					bestGain = gain
					bestCluster = targetCluster
				}
			}

			if bestCluster != nodeToCluster[node] {
				nodeToCluster[node] = bestCluster
				improved = true
			}
		}
	}

	clusters := make(map[string][]string)
	for node, cluster := range nodeToCluster {
		clusters[cluster] = append(clusters[cluster], node)
	}

	finalClusters := [][]string{}
	for _, nodes := range clusters {
		if len(nodes) >= minClusterSize {
			finalClusters = append(finalClusters, nodes)
		} else {
			if len(nodes) > 0 {
				if len(finalClusters) == 0 || finalClusters[len(finalClusters)-1][0] != "default" {
					finalClusters = append(finalClusters, []string{})
				}
				finalClusters[len(finalClusters)-1] = append(finalClusters[len(finalClusters)-1], nodes...)
			}
		}
	}

	return finalClusters, nil
}

func LouvainClusteringWithMinSize_P(graph *Graph, minClusterSize int) ([][]string, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph cannot be nil")
	}

	nodeToCluster := make(map[string]string)
	for node := range graph.Nodes {
		nodeToCluster[node] = node
	}

	var mu sync.Mutex

	improved := true
	for improved {
		improved = false

		results := make(chan struct {
			Node        string
			BestCluster string
			BestGain    float64
		})

		var wg sync.WaitGroup
		for node := range graph.Nodes {
			wg.Add(1)
			go func(node string) {
				defer wg.Done()
				var bestCluster string
				var bestGain float64

				mu.Lock()
				localClusters := make(map[string]string)
				for k, v := range nodeToCluster {
					localClusters[k] = v
				}
				mu.Unlock()

				bestCluster = localClusters[node]
				bestGain = 0.0

				for neighbor := range graph.Edges[node] {
					targetCluster := localClusters[neighbor]
					gain := calculateModularityGain(graph, localClusters, node, targetCluster)
					if gain > bestGain {
						bestGain = gain
						bestCluster = targetCluster
					}
				}

				results <- struct {
					Node        string
					BestCluster string
					BestGain    float64
				}{Node: node, BestCluster: bestCluster, BestGain: bestGain}
			}(node)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for result := range results {
			mu.Lock()
			if nodeToCluster[result.Node] != result.BestCluster {
				nodeToCluster[result.Node] = result.BestCluster
				improved = true
			}
			mu.Unlock()
		}
	}

	clusters := make(map[string][]string)
	for node, cluster := range nodeToCluster {
		clusters[cluster] = append(clusters[cluster], node)
	}

	finalClusters := [][]string{}
	for _, nodes := range clusters {
		if len(nodes) >= minClusterSize {
			finalClusters = append(finalClusters, nodes)
		} else {
			if len(nodes) > 0 {
				if len(finalClusters) == 0 || (len(finalClusters) > 0 && finalClusters[len(finalClusters)-1][0] != "default") {
					finalClusters = append(finalClusters, []string{"default"})
				}
				finalClusters[len(finalClusters)-1] = append(finalClusters[len(finalClusters)-1], nodes...)
			}
		}
	}

	return finalClusters, nil
}

func SubsystemBasedClustering(g *Graph) ([][]string, error) {
	if g == nil {
		return nil, fmt.Errorf("graph cannot be nil")
	}

	subsystemClusters := make(map[string][]string)
	for nodeName, node := range g.Nodes {
		if node.Subsystem == "" {
			return nil, fmt.Errorf("node %s has no Subsystem defined", nodeName)
		}
		subsystemClusters[node.Subsystem] = append(subsystemClusters[node.Subsystem], nodeName)
	}

	clusters := [][]string{}
	for _, clusterNodes := range subsystemClusters {
		clusters = append(clusters, clusterNodes)
	}

	return clusters, nil
}

func getMd5Parts(input string) (byte, byte, byte) {
	hash := md5.Sum([]byte(input))

	first8 := hash[0]
	middle8 := hash[len(hash)/2]
	last8 := hash[len(hash)-1]

	return first8, middle8, last8
}

func (graph *Graph) ExtractSubsystemSubgraph(nodeName string) (*Graph, error) {
	node, exists := graph.Nodes[nodeName]
	if !exists {
		return nil, fmt.Errorf("node %q not found in the graph", nodeName)
	}

	subsystem := node.Subsystem
	if subsystem == "" {
		return nil, fmt.Errorf("node %q does not have a subsystem defined", nodeName)
	}

	subgraph := NewGraph()
	subgraph.Name = fmt.Sprintf("subgraph_%s", subsystem)

	fakeNodes := make(map[string]string)

	for src, targets := range graph.Edges {
		srcNode, srcExists := graph.Nodes[src]
		if !srcExists {
			continue
		}

		if srcNode.Subsystem == subsystem {
			if _, exists := subgraph.Nodes[src]; !exists {
				subgraph.Nodes[src] = &Node{
					Color:     srcNode.Color,
					Id:        srcNode.Id,
					Subsystem: srcNode.Subsystem,
				}
			}

			for tgt, edge := range targets {
				tgtNode, tgtExists := graph.Nodes[tgt]
				if !tgtExists {
					continue
				}

				if tgtNode.Subsystem == subsystem {
					if _, exists := subgraph.Edges[src]; !exists {
						subgraph.Edges[src] = make(map[string]*Edge)
					}
					subgraph.Edges[src][tgt] = edge
				} else {
					if _, exists := fakeNodes[tgtNode.Subsystem]; !exists {
						fakeNodeName := fmt.Sprintf("subsystem_%s", tgtNode.Subsystem)
						subgraph.Nodes[fakeNodeName] = &Node{
							Color:     HighlighColor,
							Id:        fakeNodeName,
							Subsystem: tgtNode.Subsystem,
						}
						fakeNodes[tgtNode.Subsystem] = fakeNodeName
					}

					fakeNodeName := fakeNodes[tgtNode.Subsystem]
					if _, exists := subgraph.Edges[src]; !exists {
						subgraph.Edges[src] = make(map[string]*Edge)
					}
					subgraph.Edges[src][fakeNodeName] = &Edge{
						Color: HighlighColor,
					}
				}
			}
		}
	}

	return subgraph, nil
}

func (g1 *Graph) AddMissingEdges(g2 *Graph, c string) {
	color := "red"
	if c != "none" {
		color = c
	}


	for _, node1 := range g1.Nodes {
		parented := false
		if _, exists := g2.Nodes[node1.Id]; !exists {
			continue
		}
		if g1.HasNoParent(node1.Id) {
			for _, g2n1 := range g2.Nodes {
				if g2.Edges[g2n1.Id] != nil && g2.Edges[g2n1.Id][node1.Id] != nil {
						if _, exists := g1.Nodes[g2n1.Id]; !exists {
							g1.Nodes[node1.Id].Color = "orange"
							continue
						}

					parented = true
					if g1.Edges[g2n1.Id] == nil {
						g1.Edges[g2n1.Id] = make(map[string]*Edge)
					}
					if g1.Edges[g2n1.Id][node1.Id] == nil {
						g1.Edges[g2n1.Id][node1.Id] = &Edge{Color: color}
					}
				}
			}
		if parented {
			g1.Nodes[node1.Id].Color = color
		}
		}
	}
}

func (g1 *Graph) ListNoParent() []string {
	res:= []string {}

	for _, node := range g1.Nodes {
		if g1.HasNoParent((*node).Id) {
			res = append(res, (*node).Id)
		}
	}
	return res
}

func (g *Graph) FollowPath(startNode string, newColor string) {
	if g == nil {
		return
	}

	visited := make(map[string]bool)
	queue := []string{startNode}

	for len(queue) > 0 {
		nodeId := queue[0]
		queue = queue[1:]

		if visited[nodeId] {
			continue
		}
		visited[nodeId] = true

		// Change node color
		if node, exists := g.Nodes[nodeId]; exists {
			node.Color = newColor
		}

		// Change edges and enqueue adjacent nodes
		for neighbor, edge := range g.Edges[nodeId] {
			edge.Color = newColor
			queue = append(queue, neighbor)
		}
	}
}


func (g *Graph) SubGraphFromDirect(start string, color string) *Graph {
	newGraph := &Graph{
		Name:         g.Name + "_subgraph",
		Nodes:        make(map[string]*Node),
		Edges:        make(map[string]map[string]*Edge),
		ReverseEdges: make(map[string]map[string]bool),
	}

	if _, exists := g.Nodes[start]; !exists {
		return newGraph
	}

	reachable := make(map[string]bool)
	var dfs func(string)
	dfs = func(node string) {
		if reachable[node] {
			return
		}
		reachable[node] = true
		for neighbor := range g.Edges[node] {
			dfs(neighbor)
		}
	}

        dfs(start)

	for node := range reachable {
		newGraph.Nodes[node] = &Node{
			Color:     "red",
			Id:        g.Nodes[node].Id,
			Subsystem: g.Nodes[node].Subsystem,
		}
		if node == start {
			newGraph.Nodes[node].Color = color
		} else {
			newGraph.Nodes[node].Color = g.Nodes[node].Color
		}
		newGraph.Edges[node] = make(map[string]*Edge)
		for neighbor, edge := range g.Edges[node] {
			if reachable[neighbor] {
				newGraph.Edges[node][neighbor] = &Edge{Color: edge.Color}
			}
		}
	}

	return newGraph
}

func findSubGraph(g *Graph, start string) map[string]bool {
	reachable := make(map[string]bool)
	var dfs func(string)

	dfs = func(node string) {
		if reachable[node] {
			return
		}
		reachable[node] = true
		for neighbor := range g.Edges[node] {
			dfs(neighbor)
		}
		for neighbor := range g.ReverseEdges[node] {
			dfs(neighbor)
		}
	}

	dfs(start)
	return reachable
}

func (g *Graph) SubGraphFrom(start string, color string) *Graph {
	newGraph := &Graph{
		Name:         g.Name + "_subgraph",
		Nodes:        make(map[string]*Node),
		Edges:        make(map[string]map[string]*Edge),
		ReverseEdges: make(map[string]map[string]bool),
	}

	if _, exists := g.Nodes[start]; !exists {
		return newGraph
	}

	reachable := findSubGraph(g, start)

	for node := range reachable {
		newGraph.Nodes[node] = &Node{
			Color:     "red",
			Id:        g.Nodes[node].Id,
			Subsystem: g.Nodes[node].Subsystem,
		}
		if node == start {
			newGraph.Nodes[node].Color = color
		} else {
			newGraph.Nodes[node].Color = g.Nodes[node].Color
		}
		newGraph.Edges[node] = make(map[string]*Edge)
		for neighbor, edge := range g.Edges[node] {
			if reachable[neighbor] {
				newGraph.Edges[node][neighbor] = &Edge{Color: edge.Color}
			}
		}
		newGraph.ReverseEdges[node] = make(map[string]bool)
		for neighbor := range g.ReverseEdges[node] {
			if reachable[neighbor] {
				newGraph.ReverseEdges[node][neighbor] = true
			}
		}
	}

	return newGraph
}

func (graph *Graph)ParseKconfigFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentNode *Node
	var currentSymbol *KconfigSymbol
	var inHelp bool
	var helpText strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		if strings.HasPrefix(line, "depends on") {
			if currentNode != nil {
				debugIOPrintf("case depends on str='%s'\n", strings.TrimPrefix(line, "depends on "))
				expr := graph.parseKconfigExpr(strings.TrimPrefix(line, "depends on "))
				if expr != nil && expr.Operand != nil {
					graph.AddEdge(expr.Operand.Id, currentNode.Id, &Edge{Color: DepColor})
				}
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if inHelp {
			if strings.HasPrefix(line, " ") {
				helpText.WriteString(line + "\n")
				continue
			} else {
				inHelp = false
				if currentSymbol != nil {
					currentSymbol.Description = helpText.String()
				}
			}
		}

		switch fields[0] {
		case "config", "menuconfig":
			debugIOPrintf("handling config >>>>>>>>>>>>>>>>>>%s<<<<<<<<<<<<<<<<<\n", fields[1])
			name := fields[1]
			if node, exists := graph.Nodes[name]; exists {
				currentNode = node
				node.NodeInfo = &KconfigSymbol{Name: name}
				node.Complete = true
			} else {
				debugIOPrintf("Current is set to '%s'<<<<\n", name)
				currentSymbol = &KconfigSymbol{Name: name}
				currentNode = &Node{Id: name, NodeInfo: currentSymbol, Complete: true}
				graph.AddNode(name, currentNode)
			}

		case "select":
			if currentNode != nil {
                                debugIOPrintf("case select str='%s'<<<<\n", strings.Join(fields[1:], " "))
                                expr := graph.parseKconfigExpr(strings.Join(fields[1:], " "))
                                if expr != nil && expr.Operand != nil {
					debugIOPrintf("AddEdge(%s, %s)'\n", currentNode.Id, expr.Operand.Id)
                                        graph.AddEdge(currentNode.Id, expr.Operand.Id, &Edge{Color: DefaultColor})
                                }
			}

		case "default":
			if currentSymbol != nil {
				currentSymbol.Default = strings.Join(fields[1:], " ")
			}

		case "prompt":
			if currentSymbol != nil {
				currentSymbol.Prompt = strings.Join(fields[1:], " ")
			}

		case "help":
			inHelp = true
			helpText.Reset()
		}
	}

	return scanner.Err()
}

func generateOperatorId(op string, left, right *KconfigExpr) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%p-%p", op, left, right)))
	return "op-" + hex.EncodeToString(hash[:])
}

func tokenizeExpression(expr string) []string {
	var tokens []string
	var current strings.Builder
	i := 0

	for i < len(expr) {
		ch := expr[i]

		switch ch {
		case ' ', '\t':
			debugIOPrintf("case whitespace\n")
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		case '(', ')':
			debugIOPrintf("case Parentheses\n")
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(ch))
		case '!':
			debugIOPrintf("case !\n")
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(ch))
		case '&', '|':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			if i+1 < len(expr) && expr[i+1] == ch {
				debugIOPrintf("case %s\n", string(ch)+string(ch))
				tokens = append(tokens, string(ch)+string(ch))
				i++
			} else {
				tokens = append(tokens, string(ch))
			}
		case 'i':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			if i+1 < len(expr) && expr[i+1] == 'f' && i+2 < len(expr) && expr[i+2] == ' ' {
				debugIOPrintf("case if\n")
				tokens = append(tokens, "if")
				i++
			}

		default:
			current.WriteRune(rune(ch))
		}
		i++
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
func (g *Graph) FindIncompleteNodes() []string {
	incomplete := []string{}
	for id, node := range g.Nodes {
		if !node.Complete {
			incomplete = append(incomplete, id)
		}
	}
	return incomplete
}

func (graph *Graph)parseKconfigExpr(exprStr string) *KconfigExpr {
	tokens := tokenizeExpression(exprStr)
	debugIOPrintf("parseKconfigExpr: tokens=\"%v\"\n", tokens)
	pos := 0
	return parseExpr(graph, tokens, &pos)
}

func parseExpr(graph *Graph, tokens []string, pos *int) *KconfigExpr {
	if *pos >= len(tokens) {
		debugIOPrintf("pos > len(token)\n", tokens)
		return nil
	}

	var left *KconfigExpr

	token := tokens[*pos]
	debugIOPrintf("Current = '%s' (%d/%d)\n", token, *pos, len(tokens))
	*pos++

	switch token {
	case "(":
		left = parseExpr(graph, tokens, pos)
		*pos++
	case "!":
		debugIOPrintf("generate ! subtree\n")
		right := parseExpr(graph, tokens, pos)
		id := generateOperatorId("!", nil, right)
		operatorNode := &Node{Id: id, FillColor: OpColor, Label: "!", Shape: OpShape}
		graph.AddNode(id, operatorNode)
		graph.AddEdge(id, right.Operand.Id, &Edge{Color: DepColor})
		left = &KconfigExpr{Op: "!", Right: right, Operand: operatorNode}
	default:
		if _, exists := graph.Nodes[token]; !exists {
			graph.AddNode(token, &Node{Id: token, Complete: false})
		}
		left = &KconfigExpr{Operand: graph.Nodes[token]}
	}

	for *pos < len(tokens) {
		op := tokens[*pos]

		if op != "&&" && op != "||" && op != "if"{
			break
		}

		*pos++

		right := parseExpr(graph, tokens, pos)

		id := generateOperatorId(op, left, right)

		if op != "if" {
			operatorNode := &Node{Id: id, FillColor: OpColor, Label: op, Shape: OpShape}
			graph.AddNode(id, operatorNode)
			graph.AddEdge(left.Operand.Id, id, &Edge{Color: DepColor})
			graph.AddEdge(right.Operand.Id, id, &Edge{Color: DepColor})

			left = &KconfigExpr{Op: op, Left: left, Right: right, Operand: operatorNode}
		} else {
			operatorNode := &Node{Id: id, FillColor: IfColor, Label: op, Shape: IfShape}
			graph.AddNode(id, operatorNode)
			graph.AddEdge(left.Operand.Id, id, &Edge{Color: DepColor})
			graph.AddEdge(id, right.Operand.Id, &Edge{Color: DepColor})

			left = &KconfigExpr{Op: op, Left: left, Right: right, Operand: operatorNode}
		}
	}

	return left
}

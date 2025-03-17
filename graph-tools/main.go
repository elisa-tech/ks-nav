package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
)
type CmdFn func([]string, string)(int)

type Cmds struct {
	ID string
	Help string
	Fn CmdFn
}

const (
	ErrMsg string = "Error: use -h for help\n"
	ErrMsgParse string = "Error: parse error: %v\n"
	ErrMsgFile string = "Error: Can't read file: %v\n"
	ErrMsgFparse string = "can't parse %f: %v\n"
	ErrMsgSdtdinUsed string = "Stdin is not empty, are you sure you want to use input file?\n"
	ErrNodeNotExists string = "%s does not existr in %s\n"
	ErrConnGrph string = "Error connecting graphs\n"
	ErrCreateGrph string = "Error creating graph:"
)
const (
	ExitOk int = iota
	ExitErrFile
	ExitErrContent
	ExitErrArgs
	ExitErrNode
	ExitErrAlg
)

var cmds = [...]Cmds {
	{ ID: "addkconfig", 	Help: "adds Kconfig info to a graph", Fn: runAddKconfig,},
	{ ID: "ftrace2graph", 	Help: "Process ftrace log and output DOT graph", Fn: runF2G,},
	{ ID: "compare", 	Help: "Compare 2 DOT graphs and highlight differences", Fn: runCompare,},
	{ ID: "cutoff", 	Help: "Transform the graph such as the specified node is not expanded", Fn: runCutoff,},
	{ ID: "collapse", 	Help: "Collapses nodes in the graph by iteratively merging nodes reachable exclusively through each specified node from the given start point", Fn: runCollapse,},
	{ ID: "connect", 	Help: "Connect 2 DOT graph connecting the given nodes", Fn: runConnect,},
	{ ID: "subgraph", 	Help: "change color to the path originating from a given node", Fn: runSubgraph,},
	{ ID: "findpath", 	Help: "Return the path between two nodes, if exist", Fn: runFindPath,},
	{ ID: "kerncollapse", 	Help: "Replaces the subgraph of the nodes reachable from a given one with a newone", Fn: runKernCollapse,},
	{ ID: "cluster", 	Help: "Computes authomatic/manual clustering of a given graph", Fn: runCluster,},
	{ ID: "print", 		Help: "Print enhance graph", Fn: runPrint,},
	{ ID: "subsystem", 	Help: "extract a subsystem from an enanched graph", Fn: runSubsystem,},
	{ ID: "complete", 	Help: "search for missing arches existing in g2 in g1", Fn: runComplete,},
	{ ID: "noParentList", 	Help: "list no parent node of a given graph", Fn: runNoParentList,},
	{ ID: "followpath", 	Help: "subgraph starting from the given node", Fn: runFollowPath,},
	{ ID: "reachfrom", 	Help: "produce the subgraph reached from a given node", Fn: runReachFrom,},
}

func runF2G(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	excludeRoots := flags.String("exclude", "", "Comma-separated list of root nodes to exclude paths from")
	logFile := flags.String("file", "/dev/stdin", "Path to the ftrace log file ")
	flags.Parse(args)

	rootNodes := []string{}
	if *excludeRoots != "" {
		rootNodes = strings.Split(*excludeRoots, ",")
	}

	lineRegex := regexp.MustCompile(`([a-zA-Z0-9_.]+) <-([a-zA-Z0-9_.]+)`)
	graph := NewGraph()

	file, err := os.Open(*logFile)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrFile
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := lineRegex.FindStringSubmatch(line); matches != nil {
			caller := matches[2]
			callee := matches[1]
			graph.AddEdge(caller, callee, &Edge{Color: "black"})
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	graph.RemovePathsFrom(rootNodes)
	fmt.Println(graph.ToDotRaw(cmdId, O_HORIZONTAL, true))
	return ExitOk
}

func runCompare(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	f1 := flags.String("f1", "graph1.dot", "Path to first DOT graph")
	f2 := flags.String("f2", "graph2.dot", "Path to second DOT graph")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	graph1, err := ParseDOT(*f1, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	graph2, err := ParseDOT(*f2, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	merged := MergeWithHighlight(graph1, graph2)
	fmt.Println(merged.ToDot(cmdId, O_HORIZONTAL))
	return ExitOk
}

func runCutoff(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	nodes := flags.String("nodes", "", "Comma-separated list of node names")
	startNode := flags.String("start", "", "Start node")
	debugf := flags.Bool("debug", false, "Enable debug messages")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	if *graphFile == "" || *nodes == "" || *startNode == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}

	debug := *debugf
	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	actualNodeList := []string{}
	nodeList := strings.Split(*nodes, ",")
	for _, ex := range nodeList {
		for k := range graph.Nodes {
			m, _ := regexp.MatchString(ex, k)
			if m {
				if debug {
					fmt.Printf("Adding node: %s\n", k)
				}
				actualNodeList = append(actualNodeList, k)
			}
		}
	}

	for _, node := range actualNodeList {
		if _, ok := graph.Nodes[node]; ok {
			graph.DetachAndClean(*startNode, node)
		} else {
			fmt.Printf(ErrNodeNotExists, node)
		}
	}

	fmt.Println(graph.ToDot(cmdId, O_HORIZONTAL))
	return ExitOk
}

func runCollapse(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	nodes := flags.String("nodes", "", "Comma-separated list of node names")
	startNode := flags.String("start", "", "start node")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	if *graphFile == "" || *nodes == "" || *startNode == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}

	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	nodeList := strings.Split(*nodes, ",")

	for _, node := range nodeList {
		listNodes := graph.OnlyReachableNodes(*startNode, node)
		for _, current := range listNodes {
			graph.collapseNode(node, current)
		}
	}

	fmt.Println(graph.ToDot(cmdId, O_HORIZONTAL))
	return ExitOk
}
func runConnect(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	f1 := flags.String("f1", "Graph1", "Path to first DOT graph")
	f2 := flags.String("f2", "Graph2", "Path to second DOT graph")
	n1 := flags.String("n1", "nodeg1", "Node in Graph1 to connect")
	n2 := flags.String("n2", "nodeg2", "Node in Graph2 to connect")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}
	graph1, err := ParseDOT(*f1, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	graph2, err := ParseDOT(*f2, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	if !graph1.NodeExists(*n1) {
		fmt.Printf(ErrNodeNotExists, *n1, *graph1)
		return ExitErrNode
	}
	if !graph2.NodeExists(*n2) {
		fmt.Printf(ErrNodeNotExists, *n2, *graph2)
		return ExitErrNode
	}
	newGraph, err := ConnectGraphs(graph1, graph2, *n1, *n2, "red")
	if err != nil {
		fmt.Printf(ErrConnGrph)
		return ExitErrAlg
	}
	fmt.Println(newGraph.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runSubgraph(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	node := flags.String("node", "", "Node name where start")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	if *graphFile == "" || *node == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}

	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	processedGraph, err := graph.Subgraph(*node)
	if err != nil {
		fmt.Println(err)
		return ExitErrAlg
	}

	fmt.Println(processedGraph.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runFindPath(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	nodes := flags.String("start", "", "Node name where start")
	nodee := flags.String("end", "", "Node name where arrive")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	if *graphFile == "" || *nodes == "" || *nodee == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}

	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	path, err := graph.FindPath(*nodes, *nodee)
	if err != nil {
		fmt.Println(err)
		return ExitErrAlg
	}

	for _, node := range path {
		fmt.Printf("%s\n", node)
	}
	fmt.Printf("path lenght = %d\n", len(path))
        return ExitOk
}

func runKernCollapse(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	newNode := flags.String("newnode", "", "name of the node that replaces")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	if *graphFile == "" || *newNode == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}

	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	processedGraph, err := graph.CollapseKernel(*newNode)
	if err != nil {
		fmt.Println(err)
		return ExitErrAlg
	}

	fmt.Println(processedGraph.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func checkClusteringType(target string) int {
	arglist := []string{"ls", "lp", "cp"}
	vallist := []int{LOUVAIN_SINGLE, LOUVAIN_PARALLEL, CHEAT_CLUSTERING}
	for i, str := range arglist {
		if str == target {
			return vallist[i]
		}
	}
	return PART_NONE
}

func runCluster(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	clusteringType := flags.String("clstype", "", "Specifies the clustering strategy: <ls:Louvain single thread>|lp:Louvain multithread>|<cp:graph provided>")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	if *graphFile == "" || *clusteringType == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}
	ct := checkClusteringType(*clusteringType)
	if ct == PART_NONE {
		fmt.Printf(ErrMsg)
	}
	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	cluster, err := graph.HierarchicalClustering(ct)
	if err != nil {
		fmt.Println(err)
		return ExitErrAlg
	}

	output, err := cluster.ToDot(cmdId, O_HORIZONTAL)
	if err != nil {
		fmt.Println(ErrCreateGrph, err)
		return ExitErrAlg
	}
	fmt.Println(output)
        return ExitOk
}
func runPrint(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)

	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}
	graph, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	fmt.Println(graph.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runSubsystem(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	node := flags.String("node", "", "target Node name")
	flags.Parse(args)

	if *graphFile == "" || *node == "" {
		fmt.Printf(ErrMsg)
		return ExitErrArgs
	}

	graph, err := ParseDOT(*graphFile, PARSE_SUBSYS)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	processedGraph, err := graph.ExtractSubsystemSubgraph(*node)
	if err != nil {
		fmt.Println(err)
		return ExitErrAlg
	}

	fmt.Println(processedGraph.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runComplete(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	f1 := flags.String("f1", "/dev/stdin", "Path to first DOT graph")
	f2 := flags.String("f2", "graph2.dot", "Path to second DOT graph")
	color := flags.String("color", "none", "Color used for added arches")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	graph1, err := ParseDOT(*f1, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	graph2, err := ParseDOT(*f2, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}

	graph1.AddMissingEdges(graph2, *color)
	fmt.Println(graph1.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runNoParentList(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	flags.Parse(args)

	graph, err := ParseDOT(*graphFile, PARSE_PLAIN)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	for _, l := range graph.ListNoParent() {
		fmt.Println(l)
	}
        return ExitOk

}

func runFollowPath(args []string, cmdId string) int{
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	snode := flags.String("snode", "", "Node where start")
	color := flags.String("color", "none", "Color used for added arches")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	graph1, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	graph1.FollowPath(*snode, *color)
	fmt.Println(graph1.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runReachFrom(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	snode := flags.String("snode", "", "Node where start")
	color := flags.String("color", "none", "Color used for added arches")
	EnhancedDot := flags.Bool("edot", false, "Specify the dot is in enahnced syntax")
	direct := flags.Bool("direct", false, "use arch direction")
	flags.Parse(args)
	parseType := PARSE_PLAIN
	if *EnhancedDot {
		parseType = PARSE_SUBSYS
	}

	graph1, err := ParseDOT(*graphFile, parseType)
	if err != nil {
		fmt.Printf(ErrMsgFile, err)
		return ExitErrContent
	}
	res := &Graph{}
	if *direct {
		res = graph1.SubGraphFromDirect(*snode, *color)
	} else {
		res = graph1.SubGraphFrom(*snode, *color)
	}
	fmt.Println(res.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func runAddKconfig(args []string, cmdId string) int {
	flags := flag.NewFlagSet(cmdId, flag.ExitOnError)
	graphFile := flags.String("file", "/dev/stdin", "Path to the DOT graph file")
	kconfigFile := flags.String("kconfig", "Kconfig", "Path to the Kconfig file to add")
	flags.Parse(args)
	parseType := PARSE_PLAIN

	debugIOPrintf("start: graphFile=%s, kconfigFile=%s\n", *graphFile, *kconfigFile)

	if *graphFile != "/dev/stdin" && stdinUsed() {
		fmt.Printf(ErrMsgSdtdinUsed)
		os.Exit(1)
	}

	graph1 := &Graph{}
	if *graphFile != "none" {
		var err error
		debugIOPrintf("parsing dot file %s\n", *graphFile)
		graph1, err = ParseDOT(*graphFile, parseType)
		if err != nil {
			fmt.Printf(ErrMsgFile, err)
			return ExitErrContent
		}
	} else {
		graph1 = NewGraph()
	}
	debugIOPrintf("parsing Kconfig file %s\n", *kconfigFile)
	err := graph1.ParseKconfigFile(*kconfigFile)
	if err != nil {
		fmt.Printf(ErrMsgFparse, *kconfigFile, err)
		return ExitErrAlg
	}

	fmt.Println(graph1.ToDot(cmdId, O_HORIZONTAL))
        return ExitOk
}

func stdinUsed() bool {
	debugIOPrintf("Check stdin\n")
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		debugIOPrintf("stdin is piped\n")
		return true
	}
	debugIOPrintf("stdin is from a terminal")
	return false
}

func main() {
		// Stderr would be best for errors, right?
		// Using stdout because I want to poison stdout
		// if this check fails, the produced output can not
		// be used in piped commands like:
		// graph-tools something | graph-tools somethingelse | xdot -

	if len(os.Args) < 2 {
		fmt.Println("Usage: [-d] <command> [options]")
		fmt.Println("Commands:")
		w := tabwriter.NewWriter(os.Stdout, 10, 1, 1, ' ', 0)
		for i := 0; i < len(cmds); i++ {
			fmt.Fprintf(w, "  %s\t- %s\n", cmds[i].ID, cmds[i].Help)
		}
		w.Flush()
		return
	}

	cmdIndx := 1
	if os.Args[cmdIndx]=="-d" {
		DebugLevel = debugIO | (1<<debugAddFunctionName - 1)
		cmdIndx = 2
	}
	exitCode := 0
	command := os.Args[cmdIndx]
	for i := 0; i < len(cmds); i++ {
		if command == cmds[i].ID {
			exitCode = cmds[i].Fn(os.Args[(cmdIndx + 1):], cmds[i].ID)
			break
		}
	}
	if exitCode != 0 {
		fmt.Printf("Unknown command: %s\n", command)
	}
	os.Exit(exitCode)
}

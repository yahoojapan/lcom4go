package lcom4

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	reportmsg = "'%s' has low cohesion, LCOM4 is %d, pairs of methods: %v"
)

const doc = "lcom4go caluculates cohesion metrics value"

// Analyzer is the lcom4 analyzer.
var Analyzer = &analysis.Analyzer{
	Name:     "lcom4",
	Doc:      doc,
	Run:      run,
	Requires: []*analysis.Analyzer{},
}

var argsort bool

func init() {
	Analyzer.Flags.BoolVar(&argsort, "sort", false, "sort results by the lcom4 value(increasing order)")
}

const (
	field = iota
	method
)

type graphNode interface {
	typ() int
	name() string
}

type fieldNode string

func (f fieldNode) typ() int       { return field }
func (f fieldNode) name() string   { return string(f) }
func (f fieldNode) String() string { return fmt.Sprintf(".%s", string(f)) }

type methodNode string

func (m methodNode) typ() int       { return method }
func (m methodNode) name() string   { return string(m) }
func (f methodNode) String() string { return fmt.Sprintf("%s()", string(f)) }

type graph struct {
	nodes    []graphNode
	neighbor map[graphNode][]graphNode
}

type graphs map[types.Object]graph

func initGraph(pkg *types.Package) graphs {
	graphs := map[types.Object]graph{}
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		if _, ok := o.(*types.TypeName); !ok {
			continue
		}
		if _, ok := o.Type().(*types.Named); !ok {
			continue
		}
		// skip 'type xxx interface {...}'
		if _, ok := o.Type().Underlying().(*types.Interface); ok {
			continue
		}
		g := graph{nil, map[graphNode][]graphNode{}}
		ms := collectMethods(o)
		g.nodes = append(g.nodes, ms...)
		graphs[o] = g
	}
	return graphs
}

func collectMethods(o types.Object) []graphNode {
	var nodes []graphNode

	named, ok := o.Type().(*types.Named)
	if !ok {
		return nil
	}
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		nodes = append(nodes, methodNode(m.Name()))
	}
	return nodes
}

func collectComments(pass *analysis.Pass) []ast.CommentMap {
	var ret []ast.CommentMap
	for _, f := range pass.Files {
		m := ast.NewCommentMap(pass.Fset, f, f.Comments)
		ret = append(ret, m)
	}
	return ret
}

func fillNeighbor(graphs map[types.Object]graph, pass *analysis.Pass) {
	for _, f := range pass.Files {
		ast.Inspect(f, func(node ast.Node) bool {
			switch fdecl := node.(type) {
			case *ast.FuncDecl:
				if fdecl.Recv == nil {
					break
				}
				if len(fdecl.Recv.List[0].Names) == 0 {
					break
				}
				recvType := pass.TypesInfo.TypeOf(fdecl.Recv.List[0].Type)
				if p, ok := recvType.(*types.Pointer); ok {
					recvType = p.Elem()
				}
				nd, ok := recvType.(*types.Named)
				if !ok {
					break
				}
				graph := graphs[nd.Obj()]

				recvObj := pass.TypesInfo.ObjectOf(fdecl.Recv.List[0].Names[0])

				ast.Inspect(fdecl.Body, func(node ast.Node) bool {
					switch nd := node.(type) {
					case *ast.SelectorExpr:
						xx, ok := nd.X.(*ast.Ident)
						if !ok {
							break
						}
						o := pass.TypesInfo.ObjectOf(xx)
						if recvObj != o {
							break
						}
						o2 := pass.TypesInfo.ObjectOf(nd.Sel)
						src := methodNode(fdecl.Name.Name)
						var dst graphNode
						if _, ok := o2.(*types.Var); ok {
							dst = fieldNode(nd.Sel.Name)
						} else if _, ok := o2.(*types.Func); ok {
							dst = methodNode(nd.Sel.Name)
						}
						graph.neighbor[src] = append(graph.neighbor[src], dst)
						graph.neighbor[dst] = append(graph.neighbor[dst], src)
						return false
					case *ast.Ident:
						o := pass.TypesInfo.ObjectOf(nd)
						if recvObj == o {
							src := methodNode(fdecl.Name.Name)
							dst := fieldNode("__receiver__")
							graph.neighbor[src] = append(graph.neighbor[src], dst)
							graph.neighbor[dst] = append(graph.neighbor[dst], src)
						}
					}
					return true
				})
				return false
			}
			return true
		})
	}

}

func computeConnectedComponents(g graph) [][]graphNode {
	components := [][]graphNode{}

	visited := make(map[graphNode]bool)
	for _, n := range g.nodes {
		if visited[n] {
			continue
		}

		compo := collectConnectedNodes(g, n)
		for _, m := range compo {
			visited[m] = true
		}
		components = append(components, compo)
	}
	return components
}

func collectConnectedNodes(g graph, n graphNode) []graphNode {
	var nodes []graphNode
	visited := make(map[graphNode]bool)
	q := []graphNode{n}
	for len(q) > 0 {
		head := q[0]
		q = q[1:]
		if visited[head] {
			continue
		}
		nodes = append(nodes, head)
		visited[head] = true
		q = append(q, g.neighbor[head]...)
	}
	return nodes
}

// ignore comment is: '/lint:ignore lcom4[,...,...] reason'
func hasIgnoreComment(obj types.Object, fset *token.FileSet, cmaps []ast.CommentMap) bool {
	for _, cmap := range cmaps {
		for node, cgs := range cmap {
			cline := fset.File(node.Pos()).Line(node.Pos())
			oline := fset.File(obj.Pos()).Line(obj.Pos())
			if cline != oline {
				continue
			}
			for _, cg := range cgs {
				for _, cmt := range cg.List {
					if !strings.HasPrefix(cmt.Text, "//") {
						continue
					}
					spl := strings.Split(cmt.Text[2:], " ")
					if len(spl) < 3 {
						continue
					}
					if spl[0] != "lint:ignore" {
						continue
					}
					for _, checkee := range strings.Split(spl[1], ",") {
						if checkee == "lcom4" {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	graphs := initGraph(pass.Pkg)
	fillNeighbor(graphs, pass)
	cmaps := collectComments(pass)

	type data struct {
		obj        types.Object
		g          graph
		components [][]graphNode
	}
	arr := []data{}
	for obj, g := range graphs {
		components := computeConnectedComponents(g)
		if len(components) > 1 && !hasIgnoreComment(obj, pass.Fset, cmaps) {
			arr = append(arr, data{obj, g, components})
		}
	}

	if argsort {
		sort.Slice(arr, func(i, j int) bool {
			return len(arr[i].components) < len(arr[j].components)
		})
	}

	for _, d := range arr {
		pass.Reportf(d.obj.Pos(), fmt.Sprintf(reportmsg, d.obj.Id(), len(d.components), d.components))
	}

	return nil, nil
}

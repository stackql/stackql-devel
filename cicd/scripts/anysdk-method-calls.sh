cat > /tmp/dump_anysdk_methods.go <<'GO'
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/types"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Hit struct {
	RecvType string `json:"recv_type"` // static receiver type
	Method   string `json:"method"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	PkgPath  string `json:"pkg_path"` // defining package (best effort)
}

func main() {
	var prefix string
	flag.StringVar(&prefix, "prefix", "github.com/stackql/any-sdk", "match receiver type containing this import path")
	flag.Parse()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedImports,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil { panic(err) }

	var hits []Hit
	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil || pkg.Fset == nil { continue }
		for selExpr, sel := range pkg.TypesInfo.Selections {
			if sel == nil { continue }
			if sel.Kind() != types.MethodVal && sel.Kind() != types.MethodExpr { continue }

			recv := sel.Recv()
			if recv == nil { continue }

			recvStr := types.TypeString(recv, func(p *types.Package) string {
				if p == nil { return "" }
				return p.Path()
			})
			if !strings.Contains(recvStr, prefix) { continue }

			pos := pkg.Fset.Position(selExpr.Sel.Pos())

			pkgPath := ""
			if obj := sel.Obj(); obj != nil && obj.Pkg() != nil {
				pkgPath = obj.Pkg().Path()
			}

			hits = append(hits, Hit{
				RecvType: recvStr,
				Method:   sel.Obj().Name(),
				File:     pos.Filename,
				Line:     pos.Line,
				PkgPath:  pkgPath,
			})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].RecvType != hits[j].RecvType { return hits[i].RecvType < hits[j].RecvType }
		if hits[i].Method != hits[j].Method { return hits[i].Method < hits[j].Method }
		if hits[i].File != hits[j].File { return hits[i].File < hits[j].File }
		return hits[i].Line < hits[j].Line
	})

	enc := json.NewEncoder(os.Stdout)
	for _, h := range hits {
		_ = enc.Encode(h)
	}
	fmt.Fprintln(os.Stderr, "DONE")
}
GO

go run /tmp/dump_anysdk_methods.go -prefix github.com/stackql/any-sdk > /tmp/anysdk_method_calls.jsonl

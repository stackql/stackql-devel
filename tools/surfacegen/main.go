package main

import (
	"flag"
	"fmt"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	var prefix string
	flag.StringVar(&prefix, "prefix", "github.com/stackql/any-sdk", "receiver type prefix")
	flag.Parse()

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}

	methods := map[string]map[string]string{}

	for _, pkg := range pkgs {
		for _, sel := range pkg.TypesInfo.Selections {
			if sel == nil {
				continue
			}
			if sel.Kind() != types.MethodVal && sel.Kind() != types.MethodExpr {
				continue
			}
			recv := sel.Recv()
			if recv == nil {
				continue
			}
			recvStr := types.TypeString(recv, nil)
			if !strings.Contains(recvStr, prefix) {
				continue
			}

			obj := sel.Obj()
			if obj == nil {
				continue
			}

			sig := types.TypeString(obj.Type(), nil)

			if _, ok := methods[recvStr]; !ok {
				methods[recvStr] = map[string]string{}
			}
			methods[recvStr][obj.Name()] = sig
		}
	}

	recvTypes := make([]string, 0, len(methods))
	for rt := range methods {
		recvTypes = append(recvTypes, rt)
	}
	sort.Strings(recvTypes)

	for _, rt := range recvTypes {
		fmt.Println(rt)
		mm := methods[rt]
		names := make([]string, 0, len(mm))
		for n := range mm {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Printf("    %s %s\n", n, mm[n])
		}
		fmt.Println()
	}
}

#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./scripts/dump_interfaces.sh [module_dir] [out_dir]
#
# Example:
#   ./scripts/dump_interfaces.sh . /tmp/anysdk-ifaces

MOD_DIR="${1:-.}"
OUT_DIR="${2:-/tmp/anysdk-ifaces}"

mkdir -p "$OUT_DIR"

cat > "$OUT_DIR/dump_interfaces.go" <<'GO'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Method struct {
	Name       string `json:"name"`        // empty for embedded interfaces
	Signature  string `json:"signature"`   // for methods, "Func(...)(...)"; for embeds, the embedded type expr
	Exported   bool   `json:"exported"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	IsEmbedded bool   `json:"is_embedded"`
}

type Iface struct {
	Package      string   `json:"package"`
	Name         string   `json:"name"`
	Exported     bool     `json:"exported"`
	File         string   `json:"file"`
	Line         int      `json:"line"`
	Methods      []Method `json:"methods"`
	Embeds       []Method `json:"embeds"`
	HasAnySetter bool     `json:"has_any_setter"` // heuristic: exported method starting with Set
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := rune(name[0])
	return unicode.IsUpper(r)
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, e)
	return buf.String()
}

func funcTypeString(fset *token.FileSet, ft *ast.FuncType) string {
	// Print just the func signature without name prefix
	var buf bytes.Buffer
	buf.WriteString("func")
	_ = printer.Fprint(&buf, fset, ft.Params)
	if ft.Results != nil {
		_ = printer.Fprint(&buf, fset, ft.Results)
	}
	return buf.String()
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	fset := token.NewFileSet()

	// Collect .go files excluding vendor/.git and generated-ish dirs if you want
	var goFiles []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	sort.Strings(goFiles)

	// Map: package -> interface name -> iface
	out := make(map[string]map[string]*Iface)

	for _, file := range goFiles {
		parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			// Keep going; don’t hard fail the whole dump if one file is weird
			fmt.Fprintf(os.Stderr, "WARN parse %s: %v\n", file, err)
			continue
		}
		pkg := parsed.Name.Name

		if _, ok := out[pkg]; !ok {
			out[pkg] = make(map[string]*Iface)
		}

		for _, decl := range parsed.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				it, ok := ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				pos := fset.Position(ts.Pos())

				iface := &Iface{
					Package:  pkg,
					Name:     ts.Name.Name,
					Exported: isExported(ts.Name.Name),
					File:     pos.Filename,
					Line:     pos.Line,
				}

				// Methods + embeds
				if it.Methods != nil {
					for _, field := range it.Methods.List {
						fpos := fset.Position(field.Pos())

						// Embedded interface/type: Names == nil
						if len(field.Names) == 0 {
							m := Method{
								Name:       "",
								Signature:  exprString(fset, field.Type),
								Exported:   false,
								File:       fpos.Filename,
								Line:       fpos.Line,
								IsEmbedded: true,
							}
							iface.Embeds = append(iface.Embeds, m)
							continue
						}

						// Named method
						name := field.Names[0].Name
						m := Method{
							Name:       name,
							Exported:   isExported(name),
							File:       fpos.Filename,
							Line:       fpos.Line,
							IsEmbedded: false,
						}
						if ft, ok := field.Type.(*ast.FuncType); ok {
							m.Signature = funcTypeString(fset, ft)
						} else {
							// shouldn’t happen for interface methods, but keep safe
							m.Signature = exprString(fset, field.Type)
						}
						iface.Methods = append(iface.Methods, m)
						if m.Exported && strings.HasPrefix(m.Name, "Set") {
							iface.HasAnySetter = true
						}
					}
				}

				out[pkg][iface.Name] = iface
			}
		}
	}

	// Emit JSONL sorted for diff-friendly output
	var pkgs []string
	for p := range out {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)

	enc := json.NewEncoder(os.Stdout)
	for _, p := range pkgs {
		var names []string
		for n := range out[p] {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			_ = enc.Encode(out[p][n])
		}
	}
}
GO

pushd "$MOD_DIR" >/dev/null
go run "$OUT_DIR/dump_interfaces.go" "$MOD_DIR" > "$OUT_DIR/interfaces.jsonl"
popd >/dev/null

# Helpful derived views
jq -r '
  select(.exported==true) |
  [.package, .name, (.methods|length), (.embeds|length), (.has_any_setter|tostring), .file, (.line|tostring)] |
  @tsv
' "$OUT_DIR/interfaces.jsonl" \
| (echo -e "pkg\tiface\tmethods\tembeds\thasSetter\tfile\tline"; cat) \
> "$OUT_DIR/exported_interfaces.tsv"

jq -r '
  select(.exported==true) |
  . as $i |
  ($i.methods[]? | select(.exported==true) |
    [$i.package, $i.name, .name, .signature, .file, (.line|tostring)] | @tsv
  )
' "$OUT_DIR/interfaces.jsonl" \
> "$OUT_DIR/exported_methods.tsv"

echo "Wrote:"
echo "  $OUT_DIR/interfaces.jsonl         (all interfaces, JSONL)"
echo "  $OUT_DIR/exported_interfaces.tsv  (exported interfaces summary)"
echo "  $OUT_DIR/exported_methods.tsv     (exported method signatures)"
echo
echo "Tip: open exported_methods.tsv and sort/group by iface to spot fat interfaces and setters."

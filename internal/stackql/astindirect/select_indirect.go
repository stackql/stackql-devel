package astindirect

import (
	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/symtab"
	"github.com/stackql/stackql/internal/stackql/typing"
)

type parserSelectIndirect struct {
	selectObj             *sqlparser.Select
	selCtx                drm.PreparedStatementCtx
	paramCollection       internaldto.TableParameterCollection
	underlyingSymbolTable symtab.SymTab
}

func (v *parserSelectIndirect) GetType() IndirectType {
	return SubqueryType
}

func (v *parserSelectIndirect) GetAssignedParameters() (internaldto.TableParameterCollection, bool) {
	return v.paramCollection, v.paramCollection != nil
}

func (v *parserSelectIndirect) SetAssignedParameters(paramCollection internaldto.TableParameterCollection) {
	v.paramCollection = paramCollection
}

func (v *parserSelectIndirect) GetRelationalColumns() []typing.RelationalColumn {
	return nil
}

func (v *parserSelectIndirect) GetRelationalColumnByIdentifier(_ string) (typing.RelationalColumn, bool) {
	return nil, false
}

func (v *parserSelectIndirect) GetUnderlyingSymTab() symtab.SymTab {
	return v.underlyingSymbolTable
}

func (v *parserSelectIndirect) SetUnderlyingSymTab(symbolTable symtab.SymTab) {
	v.underlyingSymbolTable = symbolTable
}

func (v *parserSelectIndirect) GetName() string {
	return ""
}

func (v *parserSelectIndirect) GetColumns() []typing.ColumnMetadata {
	return v.selCtx.GetNonControlColumns()
}

func (v *parserSelectIndirect) GetOptionalParameters() map[string]openapistackql.Addressable {
	return nil
}

func (v *parserSelectIndirect) GetRequiredParameters() map[string]openapistackql.Addressable {
	return nil
}

func (v *parserSelectIndirect) GetColumnByName(name string) (typing.ColumnMetadata, bool) {
	for _, col := range v.selCtx.GetNonControlColumns() {
		if col.GetIdentifier() == name {
			return col, true
		}
	}
	return nil, false
}

func (v *parserSelectIndirect) SetSelectContext(selCtx drm.PreparedStatementCtx) {
	v.selCtx = selCtx
}

func (v *parserSelectIndirect) GetSelectContext() drm.PreparedStatementCtx {
	return v.selCtx
}

func (v *parserSelectIndirect) GetTables() sqlparser.TableExprs {
	return nil
}

func (v *parserSelectIndirect) GetSelectAST() sqlparser.SelectStatement {
	return v.selectObj
}

func (v *parserSelectIndirect) GetSelectionCtx() (drm.PreparedStatementCtx, error) {
	return v.selCtx, nil
}

func (v *parserSelectIndirect) Parse() error {
	return nil
}

func (v *parserSelectIndirect) GetTranslatedDDL() (string, bool) {
	return "", false
}

func (v *parserSelectIndirect) GetLoadDML() (string, bool) {
	return "", false
}

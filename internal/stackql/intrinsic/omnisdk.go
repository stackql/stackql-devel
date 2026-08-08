package intrinsic

// This file holds the omnisdk-backed relations. There are two kinds:
//
//   - Metadata relations ("catalog", "methods") are small, compile-time
//     constant listings. They materialise as views, so they compose with
//     arbitrary SQL.
//   - Data relations - one per omnisdk resource - are the product of running a
//     method plan. Their rows stream from Plan.Open straight to the output
//     writer and never enter the SQL backend, so they do NOT compose into
//     subqueries, joins or CTEs, and ORDER BY / GROUP BY over them is not
//     applied. That is the deliberate cost of eager streaming.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lib/pq/oid"
	"github.com/stackql-labs/omnisdk/pkg/omnisdk"
	"github.com/stackql/psql-wire/pkg/sqldata"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

// methodPredicate is the reserved WHERE key that names the omnisdk method to
// run, disambiguating when a resource has several satisfiable methods.
const methodPredicate = "method"

// streamBatchSize is the number of rows accumulated per ISQLResultStream read.
// Small, because the point is to emit rows as they arrive.
const streamBatchSize = 64

// relationName maps an omnisdk dot-path to a stackql resource name; a stackql
// relation is provider.service.resource, so the path's dots cannot survive.
func relationName(path string) string {
	return strings.ReplaceAll(path, ".", "_")
}

func catalogRows() ([][]string, error) {
	resources, err := omnisdk.Default().Resources(".*")
	if err != nil {
		return nil, fmt.Errorf("intrinsic: cannot read omnisdk catalog: %w", err)
	}
	rows := make([][]string, 0, len(resources))
	for _, resource := range resources {
		rows = append(rows, []string{resource.Path, relationName(resource.Path), resource.Summary})
	}
	return rows, nil
}

func methodRows() ([][]string, error) {
	resources, err := omnisdk.Default().Resources(".*")
	if err != nil {
		return nil, fmt.Errorf("intrinsic: cannot read omnisdk catalog: %w", err)
	}
	var rows [][]string
	for _, resource := range resources {
		methods, methodsErr := omnisdk.Default().Methods(resource.Path)
		if methodsErr != nil {
			return nil, fmt.Errorf(
				"intrinsic: cannot read omnisdk methods for '%s': %w", resource.Path, methodsErr)
		}
		for _, method := range methods {
			rows = append(rows, []string{
				method.Path,
				relationName(resource.Path),
				method.Summary,
				strings.Join(requiredParamNames(method), ", "),
			})
		}
	}
	return rows, nil
}

func requiredParamNames(method omnisdk.Method) []string {
	var names []string
	for _, param := range method.Params {
		if param.Required {
			names = append(names, param.Name)
		}
	}
	return names
}

// dataTables synthesises one streaming relation per omnisdk resource.
func dataTables() ([]table, error) {
	resources, err := omnisdk.Default().Resources(".*")
	if err != nil {
		return nil, fmt.Errorf("intrinsic: cannot read omnisdk catalog: %w", err)
	}
	out := make([]table, 0, len(resources))
	for _, resource := range resources {
		out = append(out, dataTable(resource))
	}
	return out, nil
}

func dataTable(resource omnisdk.Resource) table {
	return table{
		name:        relationName(resource.Path),
		description: resource.Summary,
		columns:     schemaColumns(resource.Schema),
		isData:      true,
	}
}

// schemaColumns reads a JSON Schema's "required" list, which omnisdk builds in
// egress column order. A schema that declares no columns (an open object)
// yields none, and the columns are then taken from the rows themselves.
func schemaColumns(schema map[string]any) []column {
	required, ok := schema["required"].([]string)
	if !ok {
		return nil
	}
	properties, _ := schema["properties"].(map[string]any)
	cols := make([]column, 0, len(required))
	for _, name := range required {
		property, _ := properties[name].(map[string]any)
		cols = append(cols, column{name: name, dataType: schemaTypeOf(property)})
	}
	return cols
}

// lookupDataRelation resolves a stackql resource name to its omnisdk resource.
func lookupDataRelation(name string) (omnisdk.Resource, bool) {
	resources, err := omnisdk.Default().Resources(".*")
	if err != nil {
		return omnisdk.Resource{}, false
	}
	for _, resource := range resources {
		if strings.EqualFold(name, relationName(resource.Path)) {
			return resource, true
		}
	}
	return omnisdk.Resource{}, false
}

// pickMethod selects the method to run for a resource. An explicit `method`
// predicate wins; otherwise exactly one method's required params must be
// satisfied by the supplied predicates. Ambiguity is an error rather than a
// guess.
func pickMethod(resourcePath string, params map[string]string) (omnisdk.Method, error) {
	methods, err := omnisdk.Default().Methods(resourcePath)
	if err != nil {
		return omnisdk.Method{}, err
	}
	if explicit, ok := params[methodPredicate]; ok {
		for _, method := range methods {
			if strings.EqualFold(method.Path, explicit) || strings.EqualFold(lastSegment(method.Path), explicit) {
				return method, nil
			}
		}
		return omnisdk.Method{}, fmt.Errorf(
			"intrinsic: resource '%s' has no method '%s'", resourcePath, explicit)
	}
	var satisfied []omnisdk.Method
	for _, method := range methods {
		if satisfiedBy(method, params) {
			satisfied = append(satisfied, method)
		}
	}
	switch len(satisfied) {
	case 1:
		return satisfied[0], nil
	case 0:
		return omnisdk.Method{}, fmt.Errorf(
			"intrinsic: no method of '%s' has its required parameters supplied; run SHOW METHODS IN %s.%s.%s",
			resourcePath, ProviderName, serviceName, relationName(resourcePath))
	default:
		return omnisdk.Method{}, fmt.Errorf(
			"intrinsic: several methods of '%s' are satisfiable (%s); disambiguate with %s = '<name>'",
			resourcePath, strings.Join(methodPaths(satisfied), ", "), methodPredicate)
	}
}

func satisfiedBy(method omnisdk.Method, params map[string]string) bool {
	for _, name := range requiredParamNames(method) {
		if params[name] == "" {
			return false
		}
	}
	return true
}

func methodPaths(methods []omnisdk.Method) []string {
	out := make([]string, 0, len(methods))
	for _, method := range methods {
		out = append(out, method.Path)
	}
	return out
}

func lastSegment(path string) string {
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

// openStream plans and opens the omnisdk method, returning a lazy cursor.
// Credentials come from stackql's auth context for the relevant cloud.
func openStream(
	ctx queryContext, resourcePath string, params map[string]string) (*rowStream, error) {
	method, err := pickMethod(resourcePath, params)
	if err != nil {
		return nil, err
	}
	args := omnisdk.Args{Params: params, Auth: omnisdkAuth(ctx, resourcePath)}
	plan, err := omnisdk.Default().New(method.Path, args)
	if err != nil {
		return nil, err
	}
	rows, err := plan.Open(context.Background())
	if err != nil {
		return nil, err
	}
	return &rowStream{rows: rows, columns: schemaColumns(method.Schema)}, nil
}

// rowStream adapts an omnisdk cursor to sqldata.ISQLResultStream. Each Read
// pulls the next batch, so rows reach the output writer as they arrive. The
// final Read returns its batch together with io.EOF, which is the contract the
// output writers expect.
type rowStream struct {
	rows    omnisdk.Rows
	columns []column
	table   sqldata.ISQLTable
	typCfg  columnFactory
	done    bool
}

// columnFactory is the slice of typing.Config needed to render columns.
type columnFactory interface {
	GetPlaceholderColumn(table sqldata.ISQLTable, colName string, colOID oid.Oid) sqldata.ISQLColumn
}

func (rs *rowStream) Read() (sqldata.ISQLResult, error) {
	if rs.done {
		return rs.result(nil), io.EOF
	}
	batch := make([]omnisdk.Row, 0, streamBatchSize)
	for len(batch) < streamBatchSize && rs.rows.Next() {
		batch = append(batch, rs.rows.Row())
	}
	if err := rs.rows.Err(); err != nil {
		rs.done = true
		return rs.result(nil), err
	}
	// A short batch means the cursor is exhausted; emit it with io.EOF.
	if len(batch) < streamBatchSize {
		rs.done = true
		return rs.result(batch), io.EOF
	}
	return rs.result(batch), nil
}

// result renders a batch, fixing the column order from the first batch seen
// when the method schema did not declare one.
func (rs *rowStream) result(batch []omnisdk.Row) sqldata.ISQLResult {
	if len(rs.columns) == 0 && len(batch) > 0 {
		for _, name := range sortedKeys(batch[0]) {
			rs.columns = append(rs.columns, column{name: name})
		}
	}
	columns := make([]sqldata.ISQLColumn, 0, len(rs.columns))
	for _, col := range rs.columns {
		columns = append(columns, rs.typCfg.GetPlaceholderColumn(rs.table, col.name, col.oid()))
	}
	rows := make([]sqldata.ISQLRow, 0, len(batch))
	for _, row := range batch {
		values := make([]interface{}, 0, len(rs.columns))
		for _, col := range rs.columns {
			values = append(values, row[col.name])
		}
		rows = append(rows, sqldata.NewSQLRow(values))
	}
	return sqldata.NewSQLResult(columns, uint64(len(rows)), 0, rows)
}

func (rs *rowStream) Write(sqldata.ISQLResult) error {
	return fmt.Errorf("intrinsic: omnisdk result stream is read-only")
}

func (rs *rowStream) Close() error {
	return rs.rows.Close()
}

func sortedKeys(row omnisdk.Row) []string {
	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// selectFunc routes a SELECT over an omnisdk data relation to a streaming
// executor. Selects over the view-backed relations return false, so that they
// continue through the normal relational path.
func selectFunc(
	ctx queryContext,
	node *sqlparser.Select,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	if len(node.From) != 1 {
		return nil, false
	}
	aliased, ok := node.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return nil, false
	}
	tableName, ok := aliased.Expr.(sqlparser.TableName)
	if !ok {
		return nil, false
	}
	if !isAuditService(
		tableName.QualifierSecond.GetRawVal(), tableName.Qualifier.GetRawVal(), currentProvider) {
		return nil, false
	}
	resource, ok := lookupDataRelation(tableName.Name.GetRawVal())
	if !ok {
		return nil, false
	}
	params := equalityPredicates(node.Where)
	return func() internaldto.ExecutorOutput {
		stream, err := openStream(ctx, resource.Path, params)
		if err != nil {
			return internaldto.NewErroneousExecutorOutput(err)
		}
		stream.table = sqldata.NewSQLTable(0, relationName(resource.Path))
		stream.typCfg = ctx.GetTypingConfig()
		// Pull the first batch here rather than leaving it to the output
		// writer. A writer that hits an error on its first read returns before
		// emitting anything, and HandleResponse discards that error, so a
		// failure would otherwise surface as silence. Probing converts it into
		// a reported error; the remaining batches still stream.
		primed, readErr := newPrimedStream(stream)
		if readErr != nil {
			return internaldto.NewErroneousExecutorOutput(readErr)
		}
		return internaldto.NewExecutorOutput(primed, nil, nil, nil, nil)
	}, true
}

// equalityPredicates collects simple `column = 'literal'` conjuncts, which is
// how a caller supplies an omnisdk method's parameters.
func equalityPredicates(where *sqlparser.Where) map[string]string {
	params := map[string]string{}
	if where == nil {
		return params
	}
	var walk func(expr sqlparser.Expr)
	walk = func(expr sqlparser.Expr) {
		switch node := expr.(type) {
		case *sqlparser.AndExpr:
			walk(node.Left)
			walk(node.Right)
		case *sqlparser.ComparisonExpr:
			if node.Operator != sqlparser.EqualStr {
				return
			}
			col, isCol := node.Left.(*sqlparser.ColName)
			val, isVal := node.Right.(*sqlparser.SQLVal)
			if !isCol || !isVal {
				return
			}
			params[col.Name.GetRawVal()] = string(val.Val)
		}
	}
	walk(where.Expr)
	return params
}

// relationMethod is one method a relation exposes to SHOW METHODS.
type relationMethod struct {
	name           string
	description    string
	requiredParams []string
}

// methods reports the relation's callable methods. A view relation has only
// the synthetic "select"; a data relation reports the omnisdk methods behind
// it, carrying their required parameters so that a caller can see what must be
// supplied as WHERE predicates.
func (t table) methods() []relationMethod {
	viewMethod := []relationMethod{{
		name:        selectMethodName,
		description: "select-only intrinsic method",
	}}
	if !t.isData {
		return viewMethod
	}
	resource, ok := lookupDataRelation(t.name)
	if !ok {
		return viewMethod
	}
	sdkMethods, err := omnisdk.Default().Methods(resource.Path)
	if err != nil {
		return viewMethod
	}
	out := make([]relationMethod, 0, len(sdkMethods))
	for _, method := range sdkMethods {
		out = append(out, relationMethod{
			name:           lastSegment(method.Path),
			description:    method.Summary,
			requiredParams: requiredParamNames(method),
		})
	}
	return out
}

// methodColumns resolves the response shape of one of the relation's methods.
// For a data relation each omnisdk method carries its own schema, so the shape
// is the method's, not the resource's.
func (t table) methodColumns(methodName string) ([]column, bool) {
	if !t.isData {
		if strings.EqualFold(methodName, selectMethodName) {
			return t.columns, true
		}
		return nil, false
	}
	resource, ok := lookupDataRelation(t.name)
	if !ok {
		return nil, false
	}
	sdkMethods, err := omnisdk.Default().Methods(resource.Path)
	if err != nil {
		return nil, false
	}
	for _, method := range sdkMethods {
		if strings.EqualFold(lastSegment(method.Path), methodName) ||
			strings.EqualFold(method.Path, methodName) {
			return schemaColumns(method.Schema), true
		}
	}
	return nil, false
}

// cloudProviders maps the leading segment of an omnisdk resource path to the
// stackql provider whose auth context supplies its credentials. omnisdk uses
// both "google" and "gcp" prefixes for Google Cloud.
var cloudProviders = map[string]string{ //nolint:gochecknoglobals // fixed mapping
	"aws":    "aws",
	"google": "google",
	"gcp":    "google",
	"azure":  "azure",
}

// omnisdkAuth translates the stackql auth context for the cloud behind a
// resource into omnisdk's auth DTO. Values are resolved through the AuthCtx
// accessors, which already apply stackql's inline -> env var -> file
// precedence, so omnisdk receives literals and stackql's configuration stays
// authoritative. A nil return leaves omnisdk to fall back to its own canonical
// environment variables.
func omnisdkAuth(ctx queryContext, resourcePath string) *omnisdk.Auth {
	cloud, _, _ := strings.Cut(resourcePath, ".")
	providerName, ok := cloudProviders[cloud]
	if !ok {
		return nil
	}
	authCtx, err := ctx.GetAuthContext(providerName)
	if err != nil || authCtx == nil {
		return nil
	}
	auth := &omnisdk.Auth{
		Type:        authCtx.Type,
		ValuePrefix: authCtx.ValuePrefix,
		Name:        authCtx.Name,
		Scopes:      authCtx.Scopes,
		TokenURL:    authCtx.GetTokenURL(),
	}
	// The credential blob is the AWS secret key, the GCP service-account JSON,
	// or the bearer token, depending on the cloud.
	if credentials, credErr := authCtx.GetCredentialsBytes(); credErr == nil {
		auth.SecretAccessKey = string(credentials)
		auth.Credentials = string(credentials)
	}
	if keyID, keyErr := authCtx.GetKeyIDString(); keyErr == nil {
		auth.AccessKeyID = keyID
	}
	if sessionToken, tokErr := authCtx.GetAwsSessionTokenString(); tokErr == nil {
		auth.SessionToken = sessionToken
	}
	if clientID, idErr := authCtx.GetClientID(); idErr == nil {
		auth.ClientID = clientID
	}
	if clientSecret, secErr := authCtx.GetClientSecret(); secErr == nil {
		auth.ClientSecret = clientSecret
	}
	return auth
}

// primedStream is a rowStream whose first batch has already been read, so that
// an immediate failure is reported rather than swallowed. It replays that
// batch, then delegates.
type primedStream struct {
	inner    *rowStream
	first    sqldata.ISQLResult
	firstErr error
	replayed bool
}

// newPrimedStream reads the first batch. A genuine failure is returned as the
// error; exhaustion (io.EOF) is not a failure and is replayed to the caller.
func newPrimedStream(inner *rowStream) (sqldata.ISQLResultStream, error) {
	first, err := inner.Read()
	if err != nil && !errors.Is(err, io.EOF) {
		inner.Close() //nolint:errcheck // the read error is the one worth reporting
		return nil, err
	}
	return &primedStream{inner: inner, first: first, firstErr: err}, nil
}

func (ps *primedStream) Read() (sqldata.ISQLResult, error) {
	if !ps.replayed {
		ps.replayed = true
		return ps.first, ps.firstErr
	}
	return ps.inner.Read()
}

func (ps *primedStream) Write(sqldata.ISQLResult) error {
	return fmt.Errorf("intrinsic: omnisdk result stream is read-only")
}

func (ps *primedStream) Close() error {
	return ps.inner.Close()
}

// reportedType is the type DESCRIBE shows for the column.
func (c column) reportedType() string {
	if c.dataType == "" {
		return columnType
	}
	return c.dataType
}

// oid maps the column's declared JSON Schema type onto a postgres OID, which is
// what fixes the column's type on the wire and in typed output formats.
func (c column) oid() oid.Oid {
	switch c.dataType {
	case "boolean":
		return oid.T_bool
	case "integer":
		return oid.T_int8
	case "number":
		return oid.T_float8
	default:
		return oid.T_text
	}
}

// schemaTypeOf reads a JSON Schema property's type. A nullable column is
// declared as a ["<type>", "null"] union, so the non-null member is the type.
// omnisdk also carries numerics losslessly as strings tagged with a `format`
// naming the logical type; that format is the truth for such a column.
func schemaTypeOf(property map[string]any) string {
	if lossless, ok := property["x-omnisdk-lossless"].(bool); ok && lossless {
		switch format, _ := property["format"].(string); format {
		case "int32", "int64":
			return "integer"
		case "float", "double":
			return "number"
		}
	}
	switch declared := property["type"].(type) {
	case string:
		return declared
	case []any:
		for _, member := range declared {
			if name, isString := member.(string); isString && name != "null" {
				return name
			}
		}
	}
	return ""
}

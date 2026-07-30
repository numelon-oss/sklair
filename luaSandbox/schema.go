package luaSandbox

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sklair/schema"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const schemaValidatorType = "sklair.schema.validator"

type schemaValidator struct {
	scope     *Scope
	validator *schema.Validator
}

type schemaResourceLoader struct {
	roots []string
}

func (s *Scope) openSchema() {
	L := s.runtime.state
	metatable := L.NewTypeMetatable(schemaValidatorType)
	methods := L.NewTable()
	L.SetFuncs(methods, map[string]lua.LGFunction{
		"validate":       validateJSONWithSchema,
		"validate_value": validateLuaValueWithSchema,
	})
	metatable.RawSetString("__index", methods)
	metatable.RawSetString("__metatable", lua.LString("Sklair JSON Schema validator"))

	s.SetModule("schema", map[string]lua.LGFunction{
		"compile": s.compileSchema,
	})
}

func (s *Scope) compileSchema(L *lua.LState) int {
	if L.GetTop() < 1 || L.GetTop() > 2 {
		L.RaiseError("schema.compile expects a schema and an optional options table")
		return 0
	}

	source, err := schemaCompileSource(L)
	if err != nil {
		return pushLuaError(L, err)
	}

	var document any
	switch value := L.Get(1).(type) {
	case lua.LString:
		document, err = parseJSON(string(value))
	default:
		document, err = s.StructuredValue(value, "schema")
	}
	if err != nil {
		return pushLuaError(L, fmt.Errorf("could not parse JSON Schema: %w", err))
	}
	location, err := s.schemaLocation(source)
	if err != nil {
		return pushLuaError(L, err)
	}
	compiled, err := schema.Compile(document, location, &schemaResourceLoader{roots: s.schemaRoots()})
	if err != nil {
		return pushLuaError(L, friendlySchemaError(err, location, source))
	}

	validator := &schemaValidator{scope: s, validator: compiled}
	userdata := L.NewUserData()
	userdata.Value = validator
	L.SetMetatable(userdata, L.GetTypeMetatable(schemaValidatorType))
	L.Push(userdata)
	return 1
}

func schemaCompileSource(L *lua.LState) (string, error) {
	if L.GetTop() == 1 || L.Get(2) == lua.LNil {
		return "", nil
	}
	options, ok := L.Get(2).(*lua.LTable)
	if !ok {
		return "", fmt.Errorf("schema.compile options must be a table or nil")
	}
	unknown := make([]string, 0)
	nonString := false
	options.ForEach(func(key lua.LValue, _ lua.LValue) {
		name, ok := key.(lua.LString)
		if !ok {
			nonString = true
			return
		}
		if string(name) != "source" {
			unknown = append(unknown, string(name))
		}
	})
	if nonString {
		return "", fmt.Errorf("schema.compile option names must be strings")
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return "", fmt.Errorf("unknown schema.compile option %q", unknown[0])
	}

	value := options.RawGetString("source")
	if value == lua.LNil {
		return "", nil
	}
	source, ok := value.(lua.LString)
	if !ok {
		return "", fmt.Errorf("schema.compile option %q must be a string", "source")
	}
	if strings.TrimSpace(string(source)) == "" {
		return "", fmt.Errorf("schema.compile option %q cannot be empty", "source")
	}
	return string(source), nil
}

func validateJSONWithSchema(L *lua.LState) int {
	if L.GetTop() != 2 {
		L.RaiseError("JSON Schema validator validate expects exactly one JSON string")
		return 0
	}
	validator := checkSchemaValidator(L, 1)
	value, err := parseJSON(L.CheckString(2))
	if err != nil {
		return pushLuaError(L, fmt.Errorf("could not parse JSON value: %w", err))
	}
	return validator.pushResult(L, value)
}

func validateLuaValueWithSchema(L *lua.LState) int {
	if L.GetTop() != 2 {
		L.RaiseError("JSON Schema validator validate_value expects exactly one value")
		return 0
	}
	validator := checkSchemaValidator(L, 1)
	value, err := validator.scope.StructuredValue(L.Get(2), "value")
	if err != nil {
		return pushLuaError(L, err)
	}
	return validator.pushResult(L, value)
}

func (v *schemaValidator) pushResult(L *lua.LState, value any) int {
	issues, err := v.validator.Validate(value)
	if err != nil {
		return pushLuaError(L, err)
	}
	luaValue, err := v.scope.luaStructuredValue(value)
	if err != nil {
		return pushLuaError(L, err)
	}

	result := L.CreateTable(0, 3)
	v.scope.tableKinds[result] = objectTable
	result.RawSetString("valid", lua.LBool(len(issues) == 0))
	result.RawSetString("value", luaValue)
	errorsTable := L.CreateTable(len(issues), 0)
	v.scope.tableKinds[errorsTable] = arrayTable
	for _, issue := range issues {
		item := L.CreateTable(0, 3)
		v.scope.tableKinds[item] = objectTable
		item.RawSetString("instancePath", lua.LString(issue.InstancePath))
		item.RawSetString("schemaPath", lua.LString(issue.SchemaPath))
		item.RawSetString("message", lua.LString(issue.Message))
		errorsTable.Append(item)
	}
	result.RawSetString("errors", errorsTable)
	L.Push(result)
	return 1
}

func checkSchemaValidator(L *lua.LState, index int) *schemaValidator {
	userdata := L.CheckUserData(index)
	validator, ok := userdata.Value.(*schemaValidator)
	if !ok {
		L.ArgError(index, "expected a Sklair JSON Schema validator")
		return nil
	}
	return validator
}

func pushLuaError(L *lua.LState, err error) int {
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}

func (s *Scope) schemaLocation(source string) (string, error) {
	if source == "" {
		return "sklair://memory/schema.json", nil
	}
	path, err := resolvePath(&s.options, source, AccessModeRead)
	if err != nil {
		return "", fmt.Errorf("invalid schema source %q: %w", source, err)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not resolve schema source %q: %w", source, err)
	}
	return schemaFileURL(absolute), nil
}

func schemaFileURL(path string) string {
	slashPath := filepath.ToSlash(filepath.Clean(path))
	if !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}
	return (&url.URL{Scheme: "sklair", Host: "local", Path: slashPath}).String()
}

func (s *Scope) schemaRoots() []string {
	roots := []string{
		filepath.Dir(s.options.FSContext.ProjectDir),
		s.options.FSContext.CacheDir,
		s.options.FSContext.TempDir,
		s.options.FSContext.GeneratedDir,
	}
	if s.options.FSContext.Mode == HookModePost {
		roots = append(roots, s.options.FSContext.BuiltDir)
	}
	resolved := make([]string, 0, len(roots))
	for _, root := range roots {
		if root == "" {
			continue
		}
		absolute, err := filepath.Abs(root)
		if err == nil {
			resolved = append(resolved, filepath.Clean(absolute))
		}
	}
	return resolved
}

func (l *schemaResourceLoader) Load(location string) (any, error) {
	parsed, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "sklair" || parsed.Host != "local" {
		return nil, fmt.Errorf("schema reference %q is unavailable; only sandboxed relative references are supported", location)
	}
	path := filepath.FromSlash(parsed.Path)
	if runtime.GOOS == "windows" {
		path = strings.TrimPrefix(path, string(filepath.Separator))
	}
	path = filepath.Clean(path)
	if !withinSchemaRoots(path, l.roots) {
		return nil, fmt.Errorf("schema reference %q escapes the Lua filesystem sandbox", location)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	document, err := parseJSON(string(content))
	if err != nil {
		return nil, err
	}
	if err := schema.CheckDocument(document); err != nil {
		return nil, err
	}
	return document, nil
}

func withinSchemaRoots(path string, roots []string) bool {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	resolvedPath, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return false
	}
	for _, root := range roots {
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		relative, err := filepath.Rel(resolvedRoot, resolvedPath)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func friendlySchemaError(err error, location string, source string) error {
	if source == "" {
		source = "in-memory schema"
	}
	return errors.New(strings.ReplaceAll(err.Error(), location, source))
}

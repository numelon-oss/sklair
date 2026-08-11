package luaSandbox

import (
	"fmt"
	"sklair/markdown"
	"sort"

	lua "github.com/yuin/gopher-lua"
)

const markdownDocumentType = "sklair.markdown.document"

type markdownDocument struct {
	scope    *Scope
	document *markdown.Document
}

func (s *Scope) openMarkdown() {
	L := s.runtime.state
	metatable := L.NewTypeMetatable(markdownDocumentType)
	methods := L.NewTable()
	L.SetFuncs(methods, map[string]lua.LGFunction{
		"headings": markdownHeadings,
		"sections": markdownSections,
		"to_html":  markdownToHTML,
	})
	metatable.RawSetString("__index", methods)
	metatable.RawSetString("__metatable", lua.LString("Sklair Markdown document"))

	s.SetModule("markdown", map[string]lua.LGFunction{
		"parse": s.parseMarkdown,
	})
}

func (s *Scope) parseMarkdown(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.RaiseError("markdown.parse expects exactly one string")
		return 0
	}
	document := &markdownDocument{scope: s, document: markdown.Parse(L.CheckString(1))}

	userdata := L.NewUserData()
	userdata.Value = document
	L.SetMetatable(userdata, L.GetTypeMetatable(markdownDocumentType))
	L.Push(userdata)
	return 1
}

func markdownToHTML(L *lua.LState) int {
	document := checkMarkdownDocument(L, 1)
	rawHTML, err := markdownRenderOptions(L)
	if err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	output, err := document.document.HTML(rawHTML)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(output))
	return 1
}

func markdownRenderOptions(L *lua.LState) (bool, error) {
	if L.GetTop() > 2 {
		return false, fmt.Errorf("Markdown document to_html expects no arguments or one options table")
	}
	if L.GetTop() == 1 || L.Get(2) == lua.LNil {
		return false, nil
	}
	options, ok := L.Get(2).(*lua.LTable)
	if !ok {
		return false, fmt.Errorf("Markdown document to_html options must be a table or nil")
	}

	unknown := make([]string, 0)
	nonString := false
	options.ForEach(func(key lua.LValue, _ lua.LValue) {
		name, ok := key.(lua.LString)
		if !ok {
			nonString = true
			return
		}
		if string(name) != "rawHTML" {
			unknown = append(unknown, string(name))
		}
	})
	if nonString {
		return false, fmt.Errorf("Markdown document to_html option names must be strings")
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return false, fmt.Errorf("unknown Markdown document to_html option %q", unknown[0])
	}

	value := options.RawGetString("rawHTML")
	if value == lua.LNil {
		return false, nil
	}
	rawHTML, ok := value.(lua.LBool)
	if !ok {
		return false, fmt.Errorf("Markdown document to_html option %q must be a boolean", "rawHTML")
	}
	return bool(rawHTML), nil
}

func markdownHeadings(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.RaiseError("Markdown document headings expects no arguments")
		return 0
	}
	document := checkMarkdownDocument(L, 1)
	documentHeadings := document.document.Headings()
	headings := L.CreateTable(len(documentHeadings), 0)
	document.scope.tableKinds[headings] = arrayTable
	for _, heading := range documentHeadings {
		item := L.CreateTable(0, 3)
		document.scope.tableKinds[item] = objectTable
		item.RawSetString("level", lua.LNumber(heading.Level))
		item.RawSetString("title", lua.LString(heading.Title))
		item.RawSetString("id", lua.LString(heading.ID))
		headings.Append(item)
	}
	L.Push(headings)
	return 1
}

func markdownSections(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.RaiseError("Markdown document sections expects no arguments")
		return 0
	}
	document := checkMarkdownDocument(L, 1)
	documentSections := document.document.Sections()
	sections := L.CreateTable(len(documentSections), 0)
	document.scope.tableKinds[sections] = arrayTable
	for _, section := range documentSections {
		item := L.CreateTable(0, 4)
		document.scope.tableKinds[item] = objectTable
		item.RawSetString("level", lua.LNumber(section.Level))
		item.RawSetString("title", lua.LString(section.Title))
		item.RawSetString("id", lua.LString(section.ID))
		item.RawSetString("text", lua.LString(section.Text))
		sections.Append(item)
	}
	L.Push(sections)
	return 1
}

func checkMarkdownDocument(L *lua.LState, index int) *markdownDocument {
	userdata := L.CheckUserData(index)
	document, ok := userdata.Value.(*markdownDocument)
	if !ok {
		L.ArgError(index, "expected a Sklair Markdown document")
		return nil
	}
	return document
}

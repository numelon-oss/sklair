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
		"links":    markdownLinks,
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
	options, err := markdownRenderOptions(L)
	if err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	output, err := document.document.HTMLWithLinks(options.rawHTML, options.links)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(output))
	return 1
}

type markdownOptions struct {
	rawHTML bool
	links   map[string]string
}

func markdownRenderOptions(L *lua.LState) (markdownOptions, error) {
	if L.GetTop() > 2 {
		return markdownOptions{}, fmt.Errorf("Markdown document to_html expects no arguments or one options table")
	}
	if L.GetTop() == 1 || L.Get(2) == lua.LNil {
		return markdownOptions{}, nil
	}
	options, ok := L.Get(2).(*lua.LTable)
	if !ok {
		return markdownOptions{}, fmt.Errorf("Markdown document to_html options must be a table or nil")
	}

	unknown := make([]string, 0)
	nonString := false
	options.ForEach(func(key lua.LValue, _ lua.LValue) {
		name, ok := key.(lua.LString)
		if !ok {
			nonString = true
			return
		}
		if string(name) != "rawHTML" && string(name) != "links" {
			unknown = append(unknown, string(name))
		}
	})
	if nonString {
		return markdownOptions{}, fmt.Errorf("Markdown document to_html option names must be strings")
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return markdownOptions{}, fmt.Errorf("unknown Markdown document to_html option %q", unknown[0])
	}

	result := markdownOptions{}
	value := options.RawGetString("rawHTML")
	if value != lua.LNil {
		rawHTML, ok := value.(lua.LBool)
		if !ok {
			return markdownOptions{}, fmt.Errorf("Markdown document to_html option %q must be a boolean", "rawHTML")
		}
		result.rawHTML = bool(rawHTML)
	}

	links := options.RawGetString("links")
	if links != lua.LNil {
		table, ok := links.(*lua.LTable)
		if !ok {
			return markdownOptions{}, fmt.Errorf("Markdown document to_html option %q must be a table", "links")
		}
		result.links = make(map[string]string)
		var linkError error
		table.ForEach(func(key lua.LValue, value lua.LValue) {
			if linkError != nil {
				return
			}
			from, fromOK := key.(lua.LString)
			to, toOK := value.(lua.LString)
			if !fromOK || !toOK {
				linkError = fmt.Errorf("Markdown document to_html links must contain only string keys and values")
				return
			}
			result.links[string(from)] = string(to)
		})
		if linkError != nil {
			return markdownOptions{}, linkError
		}
	}
	return result, nil
}

func markdownLinks(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.RaiseError("Markdown document links expects no arguments")
		return 0
	}
	document := checkMarkdownDocument(L, 1)
	documentLinks := document.document.Links()
	links := L.CreateTable(len(documentLinks), 0)
	document.scope.tableKinds[links] = arrayTable
	for _, link := range documentLinks {
		links.Append(lua.LString(link))
	}
	L.Push(links)
	return 1
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

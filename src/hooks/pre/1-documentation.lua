local docs_root = "docs"
local canonical_root = "https://sklair.numelon.com"
local separator = "\n---\n\n"

local function fail(message)
	error("documentation generation failed: " .. message, 0)
end

local function read(path)
	local content, err = fs.read("project:" .. path)
	if content == nil then
		fail(path .. ": " .. err)
	end
	return content
end

local function compile_schema(path)
	local definition = read(path)
	local validator, err = schema.compile(definition, {
		source = "project:" .. path,
	})
	if validator == nil then
		fail(path .. ": " .. err)
	end
	return validator
end

local function validate_json(validator, content, source, subject)
	local result, err = validator:validate(content)
	if result == nil then
		fail(source .. ": could not parse " .. subject .. ": " .. err)
	end
	if not result.valid then
		local messages = {}
		for _, issue in ipairs(result.errors) do
			local location = issue.instancePath
			if location == "" then
				location = "/"
			end
			messages[#messages + 1] = location .. " " .. issue.message
		end
		fail(source .. ": invalid " .. subject .. ": " .. table.concat(messages, "; "))
	end
	return result.value
end

local frontmatter_validator = compile_schema("hooks/documentation/frontmatter.schema.json")
local navigation_validator = compile_schema("hooks/documentation/navigation.schema.json")
local configuration_validator = compile_schema("hooks/documentation/configuration.schema.json")
local navigation_source = docs_root .. "/navigation.json"
local navigation = validate_json(
	navigation_validator,
	read(navigation_source),
	navigation_source,
	"navigation override"
)
local configuration_source = docs_root .. "/config.json"
local configuration = validate_json(
	configuration_validator,
	read(configuration_source),
	configuration_source,
	"documentation configuration"
)

local function scan(path)
	local entries, err = fs.scandir("project:" .. path)
	if entries == nil then
		fail(path .. ": " .. err)
	end
	table.sort(entries, function(left, right)
		return left.name < right.name
	end)
	return entries
end

local function without_markdown_extension(name)
	return name:sub(1, -4)
end

local function source_to_output(source)
	if source:sub(-9) == "/index.md" then
		return source:sub(1, -4) .. ".html"
	end
	return source:sub(1, -4) .. "/index.html"
end

local function output_to_url(output)
	if output:sub(-10) == "index.html" then
		return "/" .. output:sub(1, -11)
	end
	fail("internal output mapping error for " .. output)
end

local function parse_page(source)
	local content = read(source):gsub("\r\n", "\n")
	local boundary = content:find(separator, 1, true)
	if boundary == nil then
		fail(source .. ": missing exact JSON front-matter separator")
	end

	local metadata_source = content:sub(1, boundary - 1)
	local markdown_source = content:sub(boundary + #separator)
	local metadata = validate_json(
		frontmatter_validator,
		metadata_source,
		source,
		"front matter"
	)

	local document = markdown.parse(markdown_source)
	local document_headings = document:headings()
	local document_sections = document:sections()
	local title = nil
	local title_count = 0
	local headings = json.array()
	for _, heading in ipairs(document_headings) do
		if heading.level == 1 then
			title = heading.title
			title_count = title_count + 1
		else
			headings[#headings + 1] = json.object({
				level = heading.level,
				title = heading.title,
				id = heading.id,
				url = "#" .. heading.id,
			})
		end
	end
	if title_count == 0 then
		fail(source .. ": Markdown document does not contain a level-one heading")
	end
	if title_count > 1 then
		fail(source .. ": Markdown document contains more than one level-one heading")
	end

	local output = source_to_output(source)
	return {
		source = source,
		output = output,
		url = output_to_url(output),
		edit_url = configuration.repository.editBaseURL .. source,
		raw_url = configuration.repository.rawBaseURL .. source,
		title = title,
		description = metadata.description,
		headings = headings,
		sections = document_sections,
		document = document,
		definition = metadata.definition,
	}
end

local function documentation_entries(path)
	local entries = {}
	local index_found = false
	for _, entry in ipairs(scan(path)) do
		if entry.isDir then
			entries[#entries + 1] = {
				kind = "directory",
				key = entry.name,
				name = entry.name,
			}
		elseif entry.name:sub(-3) == ".md" then
			if entry.name == "index.md" then
				index_found = true
			else
				entries[#entries + 1] = {
					kind = "page",
					key = without_markdown_extension(entry.name),
					name = entry.name,
				}
			end
		end
	end
	if not index_found then
		fail(path .. ": every documentation directory must contain index.md")
	end
	table.sort(entries, function(left, right)
		return left.key < right.key
	end)
	return entries
end

local used_navigation_directories = {}

local function navigation_directory(path)
	local key = "."
	if path ~= docs_root then
		key = path:sub(#docs_root + 2)
	end
	local configuration = navigation.directories[key]
	if configuration ~= nil then
		used_navigation_directories[key] = true
	end
	return configuration
end

local function apply_directory_order(entries, path)
	local configuration = navigation_directory(path)
	if configuration == nil then
		return entries, false
	end

	local by_key = {}
	for _, entry in ipairs(entries) do
		by_key[entry.key] = entry
	end

	local ordered = {}
	local promoted = {}
	for _, key in ipairs(configuration.order) do
		local entry = by_key[key]
		if entry == nil then
			fail(navigation_source .. ': ordered entry "' .. key .. '" is missing from ' .. path)
		end
		if promoted[key] then
			fail(navigation_source .. ': ordered entry "' .. key .. '" is duplicated')
		end
		promoted[key] = true
		ordered[#ordered + 1] = entry
	end
	for _, entry in ipairs(entries) do
		if not promoted[entry.key] then
			ordered[#ordered + 1] = entry
		end
	end
	return ordered, configuration.numbered == true
end

local pages = json.array()
local claimed_outputs = {}

local function add_page(page)
	local previous = claimed_outputs[page.output]
	if previous ~= nil then
		fail(page.source .. " and " .. previous .. " both claim " .. page.output)
	end
	claimed_outputs[page.output] = page.source
	pages[#pages + 1] = page
end

local function discover_directory(path)
	local index = parse_page(path .. "/index.md")
	add_page(index)

	local entries = documentation_entries(path)
	local numbered
	entries, numbered = apply_directory_order(entries, path)

	local children = json.array()
	for entry_index, entry in ipairs(entries) do
		local child
		if entry.kind == "directory" then
			child = discover_directory(path .. "/" .. entry.name)
		else
			local page = parse_page(path .. "/" .. entry.name)
			add_page(page)
			child = json.object({
				page = page,
				children = json.array(),
			})
		end
		if numbered then
			child.navigation_title = tostring(entry_index) .. ". " .. child.page.title
		end
		children[#children + 1] = child
	end

	return json.object({
		page = index,
		children = children,
	})
end

local root = discover_directory(docs_root)

for path in pairs(navigation.directories) do
	if not used_navigation_directories[path] then
		fail(navigation_source .. ': configured directory "' .. path .. '" does not exist')
	end
end

local pages_by_source = {}
for _, page in ipairs(pages) do
	pages_by_source[page.source] = page
end

local function source_directory(source)
	return source:match("^(.*)/[^/]+$") or ""
end

local function normalise_source(path)
	local segments = {}
	for segment in path:gmatch("[^/]+") do
		if segment == ".." then
			if #segments == 0 then
				return nil
			end
			table.remove(segments)
		elseif segment ~= "." and segment ~= "" then
			segments[#segments + 1] = segment
		end
	end
	return table.concat(segments, "/")
end

local function relative_link_replacements(page)
	local replacements = {}
	for _, destination in ipairs(page.document:links()) do
		if destination:sub(1, 1) ~= "/"
			and destination:sub(1, 1) ~= "#"
			and not destination:match("^[%a][%w+.-]*:") then
			local path, suffix = destination:match("^([^?#]+)(.*)$")
			if path ~= nil and path:sub(-3) == ".md" then
				local resolved = normalise_source(source_directory(page.source) .. "/" .. path)
				local target = resolved ~= nil and pages_by_source[resolved] or nil
				if target == nil then
					fail(page.source .. ': relative Markdown link "' .. destination .. '" does not resolve to a documentation page')
				end
				replacements[destination] = target.url .. suffix
			end
		end
	end
	return replacements
end

for _, page in ipairs(pages) do
	local body, render_error = page.document:to_html({
		rawHTML = true,
		links = relative_link_replacements(page),
	})
	if body == nil then
		fail(page.source .. ": could not render Markdown: " .. render_error)
	end
	page.body = body
end

local definitions = json.object()
for _, page in ipairs(pages) do
	if page.definition ~= nil then
		definitions[page.url] = json.object({
			title = page.title,
			definition = page.definition,
		})
	end
end
local definitions_source, definitions_encode_error = json.encode(definitions)
if definitions_source == nil then
	fail("could not encode documentation definitions: " .. definitions_encode_error)
end
local definitions_written, definitions_write_error = fs.write(
	"generated:docs/definitions.json",
	definitions_source
)
if not definitions_written then
	fail("could not write documentation definitions: " .. definitions_write_error)
end

local function navigation_node(node, current_url)
	local children = json.array()
	for _, child in ipairs(node.children) do
		children[#children + 1] = navigation_node(child, current_url)
	end
	return json.object({
		title = node.navigation_title or node.page.title,
		url = node.page.url,
		current = node.page.url == current_url,
		children = children,
	})
end

local function navigation_tree(current_url)
	local tree = json.array()
	for _, child in ipairs(root.children) do
		tree[#tree + 1] = navigation_node(child, current_url)
	end
	return tree
end

local function find_ancestry(node, target, ancestry)
	local next_ancestry = {}
	for _, page in ipairs(ancestry) do
		next_ancestry[#next_ancestry + 1] = page
	end
	next_ancestry[#next_ancestry + 1] = node.page
	if node.page == target then
		return next_ancestry
	end
	for _, child in ipairs(node.children) do
		local found = find_ancestry(child, target, next_ancestry)
		if found ~= nil then
			return found
		end
	end
	return nil
end

local function breadcrumbs_for(page)
	local ancestry = find_ancestry(root, page, {})
	local breadcrumbs = json.array()
	for _, ancestor in ipairs(ancestry) do
		breadcrumbs[#breadcrumbs + 1] = json.object({
			title = ancestor.title,
			url = ancestor.url,
			current = ancestor == page,
		})
	end
	return breadcrumbs
end

local function neighbour(page)
	if page == nil then
		return nil
	end
	return json.object({
		title = page.title,
		url = page.url,
	})
end

local search_pages = json.array()
local search_sections = json.array()
local term_postings = {}
local search_stop_words = {}
for _, word in ipairs({
	"a", "an", "and", "are", "as", "at", "be", "been", "but", "by",
	"can", "do", "does", "for", "from", "has", "have", "in", "into",
	"is", "it", "its", "of", "on", "or", "that", "the", "their",
	"then", "there", "these", "this", "through", "to", "was", "when",
	"where", "which", "while", "with", "without", "you", "your",
}) do
	search_stop_words[word] = true
end

local function tokens(value)
	local frequencies = {}
	local count = 0
	for token in value:lower():gmatch("[a-z0-9]+") do
		if not search_stop_words[token] then
			frequencies[token] = (frequencies[token] or 0) + 1
			count = count + 1
		end
	end
	return frequencies, count
end

local function add_search_terms(value, section_id, field)
	local frequencies, count = tokens(value)
	for token, frequency in pairs(frequencies) do
		local by_section = term_postings[token]
		if by_section == nil then
			by_section = {}
			term_postings[token] = by_section
		end
		local posting = by_section[section_id]
		if posting == nil then
			posting = { 0, 0, 0, 0 }
			by_section[section_id] = posting
		end
		posting[field] = frequency
	end
	return count
end

for page_index, page in ipairs(pages) do
	local page_id = page_index - 1
	search_pages[#search_pages + 1] = json.array({
		page.url,
		page.title,
		page.description,
	})

	local section_stack = {}
	for _, section in ipairs(page.sections) do
		local section_id = #search_sections
		local anchor = ""
		if section.level ~= 1 then
			anchor = section.id
		end

		local body_length = add_search_terms(section.text, section_id, 1)
		add_search_terms(section.title, section_id, 4)
		if section.level == 1 then
			add_search_terms(page.description, section_id, 2)
			add_search_terms(page.title, section_id, 3)
		end

		local parent_section_id = -1
		for level = section.level - 1, 1, -1 do
			if section_stack[level] ~= nil then
				parent_section_id = section_stack[level]
				break
			end
		end

		search_sections[#search_sections + 1] = json.array({
			page_id,
			anchor,
			section.title,
			body_length,
			parent_section_id,
		})
		section_stack[section.level] = section_id
		for level = section.level + 1, 6 do
			section_stack[level] = nil
		end
	end
end

local search_terms = json.object()
local term_names = {}
for term in pairs(term_postings) do
	term_names[#term_names + 1] = term
end
table.sort(term_names)

for _, term in ipairs(term_names) do
	local postings = json.array()
	local section_ids = {}
	for section_id in pairs(term_postings[term]) do
		section_ids[#section_ids + 1] = section_id
	end
	table.sort(section_ids)
	local previous_section_id = 0
	for _, section_id in ipairs(section_ids) do
		local frequencies = term_postings[term][section_id]
		local packed = math.min(frequencies[1], 255)
			+ math.min(frequencies[2], 63) * 256
			+ math.min(frequencies[3], 31) * 16384
			+ math.min(frequencies[4], 31) * 524288
		postings[#postings + 1] = section_id - previous_section_id
		postings[#postings + 1] = packed
		previous_section_id = section_id
	end
	search_terms[term] = postings
end

local search_index = json.object({
	version = 1,
	pages = search_pages,
	sections = search_sections,
	terms = search_terms,
})

local search_source, search_encode_error = json.encode(search_index)
if search_source == nil then
	fail("could not encode documentation search index: " .. search_encode_error)
end
local search_written, search_write_error = fs.write(
	"generated:docs/search-index.json",
	search_source
)
if not search_written then
	fail("could not write documentation search index: " .. search_write_error)
end

for index, page in ipairs(pages) do
	documents.render({
		layout = "documentation",
		output = page.output,
		source = "project:" .. page.source,
		body = page.body,
		props = json.object({
			title = page.title,
			browsertitle = page.title .. " | Sklair Documentation",
			description = page.description,
			canonical = canonical_root .. page.url,
			currentpath = page.url,
			editurl = page.edit_url,
			rawurl = page.raw_url,
			docshome = page.url == "/docs/",
			docschild = page.url ~= "/docs/",
			hasheadings = #page.headings > 0,
			breadcrumbs = breadcrumbs_for(page),
			navigation = navigation_tree(page.url),
			previous = neighbour(pages[index - 1]),
			next = neighbour(pages[index + 1]),
			headings = page.headings,
		}),
	})
end

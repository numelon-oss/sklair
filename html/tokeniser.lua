local function strip(str)
    return str:gsub("^%s*", ""):gsub("%s*$", "")
end

local function parseAttributes(raw)
    local props = {}
    local i = 1
    local len = #raw
    local key, val = "", ""
    local state = "key"
    local quoteChar = nil
    local escape = false

    while i <= len do
        local c = raw:sub(i, i)

        if state == "key" then
            if c == "=" then
                state = "value"
            elseif c:match("%s") then
                if #key > 0 then
                    props[key] = true
                    key = ""
                end
            else
                key = key .. c
            end
        elseif state == "value" then
            if quoteChar then
                if escape then
                    val = val .. c
                    escape = false
                elseif c == "\\" then
                    escape = true
                elseif c == quoteChar then
                    props[key] = val
                    key, val, quoteChar = "", "", nil
                    state = "key"
                else
                    val = val .. c
                end
            elseif c == "'" or c == '"' then
                quoteChar = c
            elseif c:match("%s") then
                props[key] = val
                key, val = "", ""
                state = "key"
            else
                val = val .. c
            end
        end

        i = i + 1
    end

    if #key > 0 then
        if #val > 0 then
            props[key] = val
        else
            props[key] = true
        end
    end

    return props
end

-- TODO: merge skipWhitespace and skipComment since both a re compeltely disregarded anyways
-- merge into skipWhitespace
local function skipWhitespace(source, i)
    while i <= #source and source:sub(i, i):match("%s") do
        i = i + 1
    end
    return i
end

local function skipComments(source, i)
    if source:sub(i, i + 3) == "<!--" then
        i = i + 4
        while i <= #source and source:sub(i, i + 2) ~= "-->" do
            i = i + 1
        end
        return i + 3
    end
    return i
end

-- TODO: rewrite to not use goto, not advised in luvit or lua generally
local function tokenise(source)
    local tokens = {}
    local i = 1
    while i <= #source do
        --i = skipWhitespace(source, i)
        i = skipComments(source, i)

        local c = source:sub(i, i)

        if c == "<" then
            local nextChar = source:sub(i + 1, i + 1)

            -- <!DOCTYPE>
            if nextChar == "!" then
                local j = i + 2
                while j <= #source and source:sub(j, j) ~= ">" do j = j + 1 end
                local inner = strip(source:sub(i + 2, j - 1))
                tokens[#tokens + 1] = { type = "directive", value = inner }
                i = j + 1
                goto continue
            end

            -- closing tag
            if nextChar == "/" then
                local j = i + 2
                while j <= #source and source:sub(j, j) ~= ">" do j = j + 1 end
                local tagName = strip(source:sub(i + 2, j - 1))
                tokens[#tokens + 1] = { type = "tag_close", name = tagName }
                i = j + 1
                goto continue
            end

            -- opening or self-closing
            local j = i + 1
            while j <= #source and source:sub(j, j) ~= ">" do j = j + 1 end
            local chunk = source:sub(i + 1, j - 1)

            -- detect self-closing
            -- TODO: "self closing" tags are actually just regular tags with a trailing slash
            -- they are actually completely ignored by browsers, maybe we can ignore too?
            local selfClose = chunk:sub(-1) == "/"
            if selfClose then chunk = chunk:sub(1, -2) end

            local parts = {}
            local current = ""
            local quoteChar = nil
            local escape = false

            for k = 1, #chunk do
                local ch = chunk:sub(k, k)
                if quoteChar then
                    if escape then
                        current = current .. ch
                        escape = false
                    elseif ch == "\\" then
                        escape = true
                    elseif ch == quoteChar then
                        quoteChar = nil
                        current = current .. ch
                    else
                        current = current .. ch
                    end
                elseif ch == "'" or ch == '"' then
                    quoteChar = ch
                    current = current .. ch
                elseif ch:match("%s") then
                    if #current > 0 then
                        parts[#parts + 1] = current
                        current = ""
                    end
                else
                    current = current .. ch
                end
            end
            if #current > 0 then parts[#parts + 1] = current end

            local tagName = parts[1]
            local attrStr = table.concat({ table.unpack(parts, 2) }, " ")
            local props = parseAttributes(attrStr)

            tokens[#tokens + 1] = {
                type = selfClose and "tag_self" or "tag_open",
                name = tagName,
                props = props
            }

            i = j + 1
        else
            -- text node
            local j = i
            while j <= #source and source:sub(j, j) ~= "<" do j = j + 1 end
            --local text = strip(source:sub(i, j - 1))
            local text = source:sub(i, j - 1)
            if #text > 0 then
                tokens[#tokens + 1] = { type = "text", value = text }
            end
            i = j
        end
        ::continue::
    end
    return tokens
end

return tokenise

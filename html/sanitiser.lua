local insert = table.insert

local inlineTags = require("./elements/inline")

--- TODO: use efficient arr / make separate list
local rawTextPreserve = {
    pre = true, code = true, textarea = true
}

local function isInline(node)
    return node and node.type == "element" and inlineTags[node.tag]
end

local function isPreserveWhitespace(node)
    return node and node.type == "element" and rawTextPreserve[node.tag]
end

local function sanitise(node)
    if not node.children or isPreserveWhitespace(node) then
        return
    end

    local newChildren = {}
    local i = 1

    while i <= #node.children do
        local child = node.children[i]

        -- recursively sanitise child elements
        sanitise(child)

        if child.type == "text" then
            local trimmed = child.value:gsub("%s+", " ")

            local prev = node.children[i - 1]
            local next = node.children[i + 1]

            local isOnlyWhitespace = trimmed:match("^%s*$")
            local isInlineContext = isInline(prev) and isInline(next)

            if isOnlyWhitespace then
                if isInlineContext then
                    -- collapse multiple spaces to one space
                    insert(newChildren, { type = "text", value = " " })
                end
                -- otherwise skip adding this whitespace node entirely
            else
                -- keep real content, but collapse inner spaces
                insert(newChildren, { type = "text", value = trimmed })
            end
        else
            insert(newChildren, child)
        end

        i = i + 1
    end

    node.children = newChildren
end

return function(ast)
    sanitise(ast)
    return ast
end

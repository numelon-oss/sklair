local voidElements = require("./elements/void")

local fmt = string.format

local insert = table.insert
local concat = table.concat

local function construct(node)
    if node.type == "text" then
        return node.value
    elseif node.type == "directive" then
        return "<!" .. node.value .. ">"
    elseif node.type == "element" then
        local props = {}
        for k, v in pairs(node.props or {}) do
            if v == true then -- explicit check for true instead of a truthy value here
                insert(props, k)
            else
                insert(props, fmt('%s="%s"', k, v))
            end
        end

        local open = "<" .. node.tag
        if #props > 0 then open = open .. " " .. concat(props, " ") end

        local _children = node.children or {}
        if voidElements[node.tag] and #_children == 0 then
            -- TODO: create another void list which doesnt require trailing slash to close it, e.g. br
            -- actually void elements dont even need the trailing slash for closing?!
            return open .. " />"
        end

        open = open .. ">"

        local children = {}
        for _, child in ipairs(_children) do
            insert(children, construct(child))
        end

        local close = "</" .. node.tag .. ">"
        return open .. concat(children) .. close
    end
end

local function serialise(tree)
    local out = {}

    for _, child in ipairs(tree.children or {}) do
        insert(out, construct(child))
    end

    return concat(out, "\n")
end

return serialise

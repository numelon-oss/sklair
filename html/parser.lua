local sanitise = require("./sanitiser")
local voidElements = require("./elements/void")

local insert = table.insert
local remove = table.remove

local function parse(tokens)
    local root = { type = "root", children = {} }
    local stack = { root }

    for _, token in ipairs(tokens) do
        if token.type == "tag_open" then
            local node = {
                type = "element",
                tag = token.name,
                props = token.props,
                children = {}
            }
            insert(stack[#stack].children, node)

            -- the fix for void elements was somehow actually easy
            if not voidElements[token.name] then
                insert(stack, node)
            end
        elseif token.type == "tag_self" then
            local node = {
                type = "element",
                tag = token.name,
                props = token.props,
                children = {}
            }
            insert(stack[#stack].children, node)
        elseif token.type == "text" or token.type == "directive" then
            -- insert(stack[#stack].children, token)
            insert(stack[#stack].children, {
                type = token.type,
                value = token.value
            })
        elseif token.type == "tag_close" then
            if stack[#stack].tag == token.name then
                remove(stack)
            else
                print("mismatched closing tag: " .. token.name)
                -- optional error handling probably later idk
                -- TODO: error handling
                -- probaby just a warning
            end
        end
    end

    -- TODO: sanitise properly
    sanitise(root)

    return root
end

return parse

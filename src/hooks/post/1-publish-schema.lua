local source = "schema/sklair.schema.json"
local contents, read_error = fs.read("built:" .. source)
if contents == nil then
	error("could not read built:" .. source .. ": " .. read_error)
end

local definition, decode_error = json.decode(contents)
if definition == nil then
	error("could not decode built:" .. source .. ": " .. decode_error)
end

local identifier = definition["$id"]
local filename = type(identifier) == "string" and identifier:match("/([^/]+)$") or nil
if filename == nil or not filename:match("^sklair%-.*%.schema%.json$") then
	error("generated Sklair schema has an invalid $id")
end

local written, write_error = fs.write("built:schema/" .. filename, contents)
if not written then
	error("could not publish generated Sklair schema: " .. write_error)
end

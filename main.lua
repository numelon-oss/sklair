local fs = require("fs")

local tokenise = require("./html/tokeniser")
local parse = require("./html/parser")

local src = fs.readFileSync("./test.html")

local t = tokenise(src)

for _, token in pairs(t) do
    print(token.type, token.name)

    if token.type == "text" then p(token.value) end

    for var, val in pairs(token.props or {}) do
        p(var, val)
    end
end


p()
p()
p()
p()

local jenc = require("json").encode

print(jenc(t))

print()
print()

local parsed = parse(t)
print(jenc(parsed))

print()
print()

local serialise = require("./html/serialiser")
fs.writeFileSync("./out.html", serialise(parsed)) -- obviously just a test
-- in the real tool, we wont EVER be writing to src, only reading

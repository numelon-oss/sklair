{
  "description": "Run definition-level Lua once per build and reuse its unresolved emitted markup across every instance."
}
---

# Static Lua

Static Lua belongs to a definition, not an instance. Sklair runs it once when that document, component, or layout definition is first needed, stores its unresolved emitted nodes, and clones those nodes for every instance.

```html
<script type="application/x-sklair-lua" static>
  build_label = os.date("%Y-%m-%d")
  sklair.html('<small class="build-date">' .. build_label .. '</small>')
</script>
```

Blocks in one definition execute in source order and share that definition's environment. A later static block can therefore read a global assigned by an earlier one. Head blocks run before body blocks because the definition itself is prepared in document order.

<!-- TODO: explain static better here, in programming terms a bit more. for example, caluculated only once and boom done! compared to dynamic. -->

Static Lua cannot read component props or bodies: those values belong to individual instances which do not exist yet. For the same reason, a static block cannot sit beneath `sklair-if`; its execution cannot truthfully depend on an instance-level condition.

An unused definition is never prepared, so its static Lua never runs. A used definition runs once per build, not once per source file that happens to mention it. The cached output remains unresolved until instantiation, which means emitted component invocations and bindings still receive the correct surrounding context later.

Use static Lua for definition-wide derived markup or setup whose answer is identical across every instance. If props can change the answer, it is dynamic Lua.

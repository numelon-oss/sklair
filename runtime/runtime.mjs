function valueFor(values, key, templateName) {
	if (!Object.hasOwn(values, key)) {
		throw new Error(`Sklair template ${templateName} requires a value for ${key}`)
	}

	return values[key]
}

function bindTemplate(fragment, values, templateName) {
	const walker = document.createTreeWalker(fragment, NodeFilter.SHOW_ELEMENT)
	const elements = []

	while (walker.nextNode()) {
		elements.push(walker.currentNode)
	}

	for (const element of elements) {
		if (!fragment.contains(element)) continue

		const ifKey = element.getAttribute("data-sklair-if")
		if (ifKey !== null) {
			element.removeAttribute("data-sklair-if")
			if (!values[ifKey]) {
				element.remove()
				continue
			}
		}

		const textKey = element.getAttribute("data-sklair-text")
		if (textKey !== null) {
			element.textContent = valueFor(values, textKey, templateName) ?? ""
			element.removeAttribute("data-sklair-text")
		}

		for (const attribute of Array.from(element.attributes)) {
			if (attribute.name.startsWith("data-sklair-attr-")) {
				const name = attribute.name.slice("data-sklair-attr-".length)
				const value = valueFor(values, attribute.value, templateName)
				if (value == null) element.removeAttribute(name)
				else element.setAttribute(name, String(value))
				element.removeAttribute(attribute.name)
			}

			if (attribute.name.startsWith("data-sklair-prop-")) {
				const name = attribute.name.slice("data-sklair-prop-".length)
				element[name] = valueFor(values, attribute.value, templateName)
				element.removeAttribute(attribute.name)
			}

			if (attribute.name.startsWith("data-sklair-class-")) {
				const name = attribute.name.slice("data-sklair-class-".length)
				element.classList.toggle(name, Boolean(valueFor(values, attribute.value, templateName)))
				element.removeAttribute(attribute.name)
			}
		}
	}
}

function normaliseBody(body, templateName) {
	if (body instanceof Node) return [body]

	if (typeof body === "string" || body == null || typeof body[Symbol.iterator] !== "function") {
		throw new TypeError(`Sklair template ${templateName} body must be a Node, DocumentFragment, or iterable of nodes`)
	}

	const nodes = Array.from(body)
	for (const node of nodes) {
		if (!(node instanceof Node)) {
			throw new TypeError(`Sklair template ${templateName} body iterable contains a value which is not a Node`)
		}
	}
	return nodes
}

function projectBody(fragment, body, templateName) {
	const marker = fragment.querySelector("template[data-sklair-body]")

	if (body == null) {
		if (marker !== null) marker.remove()
		return
	}
	if (marker === null) {
		throw new Error(`Sklair template ${templateName} does not expose a runtime body`)
	}

	for (const node of normaliseBody(body, templateName)) {
		marker.parentNode.insertBefore(node, marker)
	}
	marker.remove()
}

function renderTemplate(name, values = {}, body) {
	const templateId = `sklair-template-${String(name).trim().toLowerCase()}`
	const template = document.getElementById(templateId)

	if (!(template instanceof HTMLTemplateElement)) {
		throw new Error(`Sklair runtime template ${String(name)} is not registered in this document`)
	}

	const fragment = template.content.cloneNode(true)
	bindTemplate(fragment, values, String(name))
	projectBody(fragment, body, String(name))
	return fragment
}

export { renderTemplate }

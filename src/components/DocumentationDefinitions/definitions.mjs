const definitionsURL = "/_sklair/docs/definitions.json";
const tooltip = document.querySelector("[data-documentation-definition], #documentation-definition");
const tooltipTitle = tooltip.querySelector("strong");
const tooltipText = tooltip.querySelector("span");
let activeLink;

function hideDefinition() {
    if (!activeLink) return;
    activeLink.removeAttribute("aria-describedby");
    activeLink = undefined;
    tooltip.hidden = true;
}

function showDefinition(link, definition) {
    activeLink?.removeAttribute("aria-describedby");
    activeLink = link;
    tooltipTitle.textContent = definition.title;
    tooltipText.textContent = definition.definition;
    tooltip.hidden = false;
    link.setAttribute("aria-describedby", tooltip.id);

    const linkRect = link.getBoundingClientRect();
    const tooltipRect = tooltip.getBoundingClientRect();
    const inset = 12;
    const left = Math.min(
        Math.max(inset, linkRect.left),
        window.innerWidth - tooltipRect.width - inset,
    );
    let top = linkRect.bottom + 8;
    if (top + tooltipRect.height > window.innerHeight - inset) {
        top = linkRect.top - tooltipRect.height - 8;
    }
    tooltip.style.left = `${left}px`;
    tooltip.style.top = `${Math.max(inset, top)}px`;
}

fetch(definitionsURL)
    .then(response => {
        if (!response.ok) throw new Error(`Definition index returned ${response.status}`);
        return response.json();
    })
    .then(definitions => {
        for (const link of document.querySelectorAll(".documentation-article a[href]")) {
            const target = new URL(link.href, window.location.href);
            if (target.origin !== window.location.origin) continue;
            const definition = definitions[target.pathname];
            if (!definition) continue;

            link.classList.add("documentation-definition-link");
            link.addEventListener("pointerenter", event => {
                if (event.pointerType === "mouse") showDefinition(link, definition);
            });
            link.addEventListener("pointerleave", hideDefinition);
            link.addEventListener("focus", () => {
                if (link.matches(":focus-visible")) showDefinition(link, definition);
            });
            link.addEventListener("blur", hideDefinition);
        }
    })
    .catch(error => console.error("Could not load Sklair documentation definitions", error));

document.addEventListener("keydown", event => {
    if (event.key === "Escape") hideDefinition();
});
window.addEventListener("scroll", hideDefinition, { passive: true });
window.addEventListener("resize", hideDefinition);

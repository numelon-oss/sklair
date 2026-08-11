const variants = {
    full: {
        source: `<SiteHeader links="true">\n    <a href="/docs/">Documentation</a>\n    <a href="/download/">Download</a>\n</SiteHeader>`,
        output: `<header class="site-header">\n    <a class="brand" href="/">Sklair</a>\n    <nav>...</nav>\n    <button aria-label="Change colour theme">...</button>\n</header>`,
        links: true,
    },
    minimal: {
        source: `<SiteHeader links="false"></SiteHeader>`,
        output: `<header class="site-header">\n    <a class="brand" href="/">Sklair</a>\n    <button aria-label="Change colour theme">...</button>\n</header>`,
        links: false,
    },
};

const source = document.querySelector("[data-compiler-source]");
const output = document.querySelector("[data-compiler-output]");
const links = document.querySelector("[data-compiler-links]");
const buttons = document.querySelectorAll("[data-compiler-variant]");

function highlight(code) {
    if (!window.Prism) return;

    try {
        window.Prism.highlightElement(code);
    } catch (error) {
        console.error("Could not highlight homepage code", error);
    }
}

function selectVariant(name) {
    const variant = variants[name];
    if (!variant) return;

    source.textContent = variant.source;
    output.textContent = variant.output;
    links.hidden = !variant.links;

    for (const button of buttons) {
        button.setAttribute("aria-pressed", String(button.dataset.compilerVariant === name));
    }

    highlight(source);
    highlight(output);
}

for (const code of document.querySelectorAll(".sklair-highlight code[class*='language-']")) {
    highlight(code);
}

for (const button of buttons) {
    button.addEventListener("click", () => selectVariant(button.dataset.compilerVariant));
}

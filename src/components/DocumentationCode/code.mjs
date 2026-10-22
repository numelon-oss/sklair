if (document.readyState === "loading") {
    await new Promise(resolve => document.addEventListener("DOMContentLoaded", resolve, { once: true }));
}

function loadScript(source) {
    return new Promise((resolve, reject) => {
        const script = document.createElement("script");
        script.src = new URL(source, import.meta.url);
        script.addEventListener("load", resolve, { once: true });
        script.addEventListener("error", () => reject(new Error(`Could not load ${script.src}`)), { once: true });
        document.head.append(script);
    });
}

await loadScript("./prism-1.30.0.min.js");
await loadScript("./prism-line-numbers-1.30.0.min.js");
await loadScript("./prism-line-highlight-1.30.0.min.js");

const status = document.querySelector("[data-documentation-code-status]");

if (window.Prism?.languages.markup && window.Prism.languages.lua) {
    window.Prism.languages.insertBefore("markup", "script", {
        "sklair-lua": {
            pattern: /(<script\b(?=[^>]*\btype=(['"])application\/x-sklair-lua\2)[^>]*>)[\s\S]*?(?=<\/script>)/i,
            lookbehind: true,
            greedy: true,
            alias: "language-lua",
            inside: window.Prism.languages.lua,
        },
    });
}

async function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        return;
    }

    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.setAttribute("readonly", "");
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.append(textarea);
    textarea.select();
    const copied = document.execCommand("copy");
    textarea.remove();
    if (!copied) throw new Error("The browser rejected the copy operation");
}

for (const code of document.querySelectorAll(".documentation-article pre > code")) {
    const pre = code.parentElement;
    if (pre.closest(".documentation-code-block")) continue;

    const source = code.textContent;
    const wrapper = document.createElement("div");
    wrapper.className = "documentation-code-block";
    pre.before(wrapper);

    const toolbar = document.createElement("div");
    toolbar.className = "documentation-code-block__toolbar";
    if (pre.dataset.title) {
        const title = document.createElement("span");
        title.className = "documentation-code-block__title";
        title.textContent = pre.dataset.title;
        toolbar.append(title);
    }

    const button = document.createElement("button");
    button.className = "documentation-code-block__copy";
    button.type = "button";
    button.textContent = "Copy";
    button.setAttribute("aria-label", "Copy code");
    toolbar.append(button);
    wrapper.append(toolbar, pre);

    button.addEventListener("click", async () => {
        try {
            await copyText(source);
            button.textContent = "Copied";
            button.setAttribute("aria-label", "Code copied");
            status.textContent = "Code copied to clipboard.";
            window.setTimeout(() => {
                button.textContent = "Copy";
                button.setAttribute("aria-label", "Copy code");
            }, 1800);
        } catch (error) {
            status.textContent = "Could not copy code to the clipboard.";
            console.error("Could not copy documentation code", error);
        }
    });

    if (window.Prism) {
        try {
            window.Prism.highlightElement(code);
        } catch (error) {
            console.error("Could not highlight documentation code", error);
        }
    }
}

for (const code of document.querySelectorAll('.documentation-article :not(pre) > code[class*="language-"]')) {
    if (!window.Prism) break;
    try {
        window.Prism.highlightElement(code);
    } catch (error) {
        console.error("Could not highlight inline documentation code", error);
    }
}

for (const pre of document.querySelectorAll(".documentation-code-block pre.line-numbers[data-line]")) {
    const rows = pre.querySelectorAll(".line-numbers-rows > span");
    const firstLine = Number(pre.dataset.start || 1);
    for (const range of pre.dataset.line.split(",")) {
        const [first, last = first] = range.split("-").map(Number);
        for (let line = first; line <= last; line++) {
            rows[line - firstLine]?.classList.add("line-numbers-row--highlighted");
        }
    }
}

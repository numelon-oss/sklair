const status = document.querySelector("[data-documentation-code-status]");

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
    wrapper.append(pre);

    const button = document.createElement("button");
    button.className = "documentation-code-block__copy";
    button.type = "button";
    button.textContent = "Copy";
    button.setAttribute("aria-label", "Copy code");
    wrapper.append(button);

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

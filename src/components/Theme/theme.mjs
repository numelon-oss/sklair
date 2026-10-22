const storageKey = "sklair-theme";
const preferences = new Set(["system", "light", "dark"]);
const preferenceOrder = ["system", "light", "dark"];
const systemTheme = matchMedia("(prefers-color-scheme: dark)");
const root = document.documentElement;

function resolvedTheme(preference) {
    return preference === "system" ? (systemTheme.matches ? "dark" : "light") : preference;
}

function applyTheme(preference, persist = false) {
    if (!preferences.has(preference)) return;

    const theme = resolvedTheme(preference);
    root.dataset.themePreference = preference;
    root.dataset.theme = theme;
    root.style.colorScheme = theme;

    const themeColour = document.querySelector('meta[name="theme-color"]');
    if (themeColour) themeColour.content = theme === "dark" ? "#17110f" : "#fff9f4";

    document.querySelectorAll('[data-theme-select]').forEach(select => {
        select.value = preference;
    });

    document.querySelectorAll('[data-theme-controls] input[name="theme"]').forEach(input => {
        input.checked = input.value === preference;
    });

    document.querySelectorAll('[data-theme-toggle]').forEach(control => {
        const nextPreference = preferenceOrder[(preferenceOrder.indexOf(preference) + 1) % preferenceOrder.length];
        const label = control.querySelector('[data-theme-label]');
        const currentLabel = preference[0].toUpperCase() + preference.slice(1);
        const nextLabel = nextPreference[0].toUpperCase() + nextPreference.slice(1);

        control.setAttribute("aria-label", `Colour theme: ${currentLabel}. Switch to ${nextLabel}.`);
        control.title = `Colour theme: ${currentLabel}`;
        if (label) label.textContent = `Colour theme: ${currentLabel}`;
    });

    if (persist) {
        try {
            localStorage.setItem(storageKey, preference);
        } catch (_) { }
    }

    window.dispatchEvent(new CustomEvent("sklairthemechange", {
        detail: { preference, theme }
    }));
}

document.addEventListener("change", event => {
    const control = event.target.closest('[data-theme-select], [data-theme-controls] input[name="theme"]');
    if (control) applyTheme(control.value, true);
});

document.addEventListener("click", event => {
    const control = event.target.closest('[data-theme-toggle]');
    if (!control) return;

    const currentPreference = preferences.has(root.dataset.themePreference) ? root.dataset.themePreference : "system";
    const nextPreference = preferenceOrder[(preferenceOrder.indexOf(currentPreference) + 1) % preferenceOrder.length];
    applyTheme(nextPreference, true);
});

systemTheme.addEventListener("change", () => {
    if (root.dataset.themePreference === "system") applyTheme("system");
});

applyTheme(root.dataset.themePreference || "system");

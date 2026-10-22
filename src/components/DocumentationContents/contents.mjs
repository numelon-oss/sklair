const headings = Array.from(document.querySelectorAll(".documentation-article h2[id], .documentation-article h3[id], .documentation-article h4[id], .documentation-article h5[id], .documentation-article h6[id]"));
const links = Array.from(document.querySelectorAll('.documentation-heading[href^="#"]'));

if (headings.length > 0 && links.length > 0) {
    let scheduled = false;

    function updateCurrentHeading() {
        scheduled = false;
        const threshold = Math.min(160, window.innerHeight * 0.25);
        let current = headings[0];

        for (const heading of headings) {
            if (heading.getBoundingClientRect().top > threshold) break;
            current = heading;
        }

        if (window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 2) {
            current = headings[headings.length - 1];
        }

        const target = `#${current.id}`;
        for (const link of links) {
            const active = link.getAttribute("href") === target;
            link.classList.toggle("active", active);
            if (active) {
                link.setAttribute("aria-current", "location");
            } else {
                link.removeAttribute("aria-current");
            }
        }
    }

    function scheduleUpdate() {
        if (scheduled) return;
        scheduled = true;
        window.requestAnimationFrame(updateCurrentHeading);
    }

    window.addEventListener("scroll", scheduleUpdate, { passive: true });
    window.addEventListener("resize", scheduleUpdate);
    updateCurrentHeading();
}

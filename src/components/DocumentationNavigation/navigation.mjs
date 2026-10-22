const storageKey = "sklair-documentation-navigation-scroll";
const desktopNavigation = document.querySelector(".documentation-shell__navigation");
const mobileDisclosure = document.querySelector(".documentation-mobile-navigation");

function savedPosition() {
	try {
		const value = Number.parseFloat(sessionStorage.getItem(storageKey));
		return Number.isFinite(value) ? value : 0;
	} catch (_) {
		return 0;
	}
}

function ensureActiveIsVisible(container, active) {
	if (!container || !active) return;
	const containerRect = container.getBoundingClientRect();
	const activeRect = active.getBoundingClientRect();
	if (activeRect.top >= containerRect.top && activeRect.bottom <= containerRect.bottom) return;
	container.scrollTop += activeRect.top - containerRect.top
		- (container.clientHeight - active.clientHeight) / 2;
}

if (desktopNavigation) {
	desktopNavigation.scrollTop = savedPosition();
	const active = desktopNavigation.querySelector(".documentation-nav-item.active");
	requestAnimationFrame(() => ensureActiveIsVisible(desktopNavigation, active));
	desktopNavigation.addEventListener("scroll", () => sessionStorage.setItem(storageKey, String(desktopNavigation.scrollTop)), { passive: true });
}

if (mobileDisclosure) {
	const mobileNavigation = mobileDisclosure.querySelector("nav");
	const active = mobileNavigation?.querySelector(".documentation-nav-item.active");
	mobileDisclosure.addEventListener("toggle", () => {
		if (mobileDisclosure.open) ensureActiveIsVisible(mobileNavigation, active);
	});
}

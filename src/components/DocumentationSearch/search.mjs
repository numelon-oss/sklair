import { renderTemplate } from "/_sklair/runtime.mjs";

const indexURL = "/_sklair/docs/search-index.json";
const stopWords = new Set([
	"a", "an", "and", "are", "as", "at", "be", "been", "but", "by", "can", "do", "does",
	"for", "from", "has", "have", "in", "into", "is", "it", "its", "of", "on", "or", "that",
	"the", "their", "then", "there", "these", "this", "through", "to", "was", "when", "where",
	"which", "while", "with", "without", "you", "your",
]);
const pageCache = new Map();
let indexRequest;

function normalise(value) {
	return value.toLocaleLowerCase().normalize("NFKD");
}

function tokenise(value) {
	return (normalise(value).match(/[a-z0-9]+/g) || []).filter(token => !stopWords.has(token));
}

function prepareIndex(source) {
	const terms = new Map(Object.entries(source.terms).map(([term, encoded]) => {
		const postings = [];
		let sectionID = 0;
		for (let index = 0; index < encoded.length; index += 2) {
			sectionID += encoded[index];
			postings.push([sectionID, encoded[index + 1]]);
		}
		return [term, postings];
	}));
	const averageLength = source.sections.reduce((sum, section) => sum + section[3], 0)
		/ Math.max(source.sections.length, 1);
	return { ...source, terms, vocabulary: [...terms.keys()], averageLength };
}

function loadIndex() {
	if (!indexRequest) {
		indexRequest = fetch(indexURL)
			.then(response => {
				if (!response.ok) throw new Error(`Search index returned ${response.status}`);
				return response.json();
			})
			.then(prepareIndex);
	}
	return indexRequest;
}

function editDistanceAtMostOne(left, right) {
	if (Math.abs(left.length - right.length) > 1) return false;
	if (left === right) return true;

	let leftIndex = 0;
	let rightIndex = 0;
	let edits = 0;
	while (leftIndex < left.length && rightIndex < right.length) {
		if (left[leftIndex] === right[rightIndex]) {
			leftIndex++;
			rightIndex++;
			continue;
		}
		if (++edits > 1) return false;
		if (left.length === right.length
			&& left[leftIndex] === right[rightIndex + 1]
			&& left[leftIndex + 1] === right[rightIndex]) {
			leftIndex += 2;
			rightIndex += 2;
		} else if (left.length > right.length) leftIndex++;
		else if (right.length > left.length) rightIndex++;
		else {
			leftIndex++;
			rightIndex++;
		}
	}
	return edits + Number(leftIndex < left.length || rightIndex < right.length) <= 1;
}

function singular(value) {
	if (value.endsWith("ies") && value.length > 4) return `${value.slice(0, -3)}y`;
	if (value.endsWith("sses") && value.length > 5) return value.slice(0, -2);
	if (value.endsWith("s") && !value.endsWith("ss") && value.length > 3) return value.slice(0, -1);
	return value;
}

function expansions(index, queryTerm) {
	const matches = [];
	for (const term of index.vocabulary) {
		if (term === queryTerm) matches.push([term, 1]);
		else if (queryTerm.length >= 3 && term.startsWith(queryTerm)) {
			matches.push([term, 0.72]);
		} else if (singular(term) === singular(queryTerm)) {
			matches.push([term, 0.82]);
		} else if (queryTerm.length >= 4 && editDistanceAtMostOne(singular(term), singular(queryTerm))) {
			matches.push([term, 0.48]);
		}
	}
	return matches;
}

function unpack(value) {
	return {
		body: value & 255,
		description: (value >> 8) & 63,
		pageTitle: (value >> 14) & 31,
		heading: (value >> 19) & 31,
	};
}

function searchIndex(index, query) {
	const queryTerms = [...new Set(tokenise(query))];
	if (queryTerms.length === 0) return [];

	const scores = new Map();
	for (const [queryPosition, queryTerm] of queryTerms.entries()) {
		for (const [term, expansionWeight] of expansions(index, queryTerm)) {
			const postings = index.terms.get(term);
			const inverseFrequency = Math.log(1 + (index.sections.length - postings.length + 0.5) / (postings.length + 0.5));
			for (const [sectionID, packed] of postings) {
				const section = index.sections[sectionID];
				const frequency = unpack(packed);
				const weightedFrequency = frequency.body
					+ frequency.description * 1.75
					+ frequency.pageTitle * 7
					+ frequency.heading * 4.5;
				const lengthNormalisation = 1.2 * (0.25 + 0.75 * section[3] / index.averageLength);
				const contribution = expansionWeight * inverseFrequency
					* (weightedFrequency * 2.2) / (weightedFrequency + lengthNormalisation);
				const match = scores.get(sectionID) || { score: 0, queryTerms: new Set(), matchedTerms: new Set() };
				match.score += contribution;
				match.queryTerms.add(queryPosition);
				match.matchedTerms.add(term);
				scores.set(sectionID, match);
			}
		}
	}

	const normalisedQuery = normalise(query.trim());
	const ranked = [];
	for (const [sectionID, match] of scores) {
		if (match.queryTerms.size !== queryTerms.length) continue;
		const section = index.sections[sectionID];
		const page = index.pages[section[0]];
		const pageTitle = normalise(page[1]);
		const sectionTitle = normalise(section[2]);
		if (pageTitle === normalisedQuery) match.score += 24;
		else if (pageTitle.includes(normalisedQuery)) match.score += 12;
		if (sectionTitle === normalisedQuery) match.score += 18;
		else if (sectionTitle.includes(normalisedQuery)) match.score += 9;
		ranked.push({ sectionID, section, page, ...match });
	}

	ranked.sort((left, right) => right.score - left.score || left.sectionID - right.sectionID);
	return ranked;
}

function groupMatches(matches) {
	const groups = new Map();
	for (const match of matches) {
		const pageID = match.section[0];
		const group = groups.get(pageID);
		if (group) group.push(match);
		else groups.set(pageID, [match]);
	}
	return [...groups.values()].slice(0, 8);
}

function loadPage(url) {
	if (!pageCache.has(url)) {
		pageCache.set(url, fetch(url)
			.then(response => {
				if (!response.ok) throw new Error(`Documentation page returned ${response.status}`);
				return response.text();
			})
			.then(source => new DOMParser().parseFromString(source, "text/html")));
	}
	return pageCache.get(url);
}

function sectionText(page, anchor) {
	const article = page.querySelector(".documentation-article");
	if (!article) return "";
	const heading = anchor ? page.getElementById(anchor) : article.querySelector("h1");
	if (!heading) return article.textContent;
	const fragments = [];
	for (let node = heading.nextElementSibling; node; node = node.nextElementSibling) {
		if (/^H[1-6]$/.test(node.tagName)) break;
		fragments.push(node.textContent);
	}
	return fragments.join(" ");
}

function snippet(text, terms, fallback) {
	const content = text.replace(/\s+/g, " ").trim();
	if (!content) return fallback;
	const lower = normalise(content);
	let match = -1;
	for (const term of terms) {
		const position = lower.indexOf(term);
		if (position !== -1 && (match === -1 || position < match)) match = position;
	}
	if (match === -1) match = 0;

	const radius = 105;
	let start = Math.max(0, match - radius);
	let end = Math.min(content.length, match + radius);
	if (start > 0) start = content.indexOf(" ", start) + 1;
	if (end < content.length) {
		const boundary = content.lastIndexOf(" ", end);
		if (boundary > start) end = boundary;
	}
	return `${start > 0 ? "..." : ""}${content.slice(start, end)}${end < content.length ? "..." : ""}`;
}

function resultURL(page, section) {
	return section[1] ? `${page[0]}#${section[1]}` : page[0];
}

function sectionLocation(index, sectionID) {
	const trail = [];
	let current = sectionID;
	while (current >= 0) {
		const section = index.sections[current];
		if (section[1]) trail.unshift(section[2]);
		current = section[4];
	}
	return trail.length > 0 ? trail.join(" › ") : "Introduction";
}

function highlightedSnippet(text, terms) {
	const fragment = document.createDocumentFragment();
	const candidates = [...terms].sort((left, right) => right.length - left.length);
	if (candidates.length === 0) {
		fragment.append(document.createTextNode(text));
		return fragment;
	}

	const escaped = candidates.map(term => term.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"));
	const matcher = new RegExp(`(${escaped.join("|")})`, "gi");
	let cursor = 0;
	for (const match of text.matchAll(matcher)) {
		if (match.index > cursor) fragment.append(document.createTextNode(text.slice(cursor, match.index)));
		const mark = document.createElement("mark");
		mark.textContent = match[0];
		fragment.append(mark);
		cursor = match.index + match[0].length;
	}
	if (cursor < text.length) fragment.append(document.createTextNode(text.slice(cursor)));
	return fragment;
}

function pathLabel(url) {
	return url === "/docs/" ? "/docs/" : url.replace(/^\/docs\//, "");
}

const dialog = document.querySelector("[data-documentation-search]");
const input = dialog.querySelector("input[type='search']");
const results = dialog.querySelector(".documentation-search__results");
const empty = results.querySelector(".documentation-search__empty");
const list = results.querySelector("ol");
const status = dialog.querySelector(".documentation-search__status");
let revision = 0;
let returnFocus;
let activeResultIndex = -1;

function resultLinks() {
	return [...list.querySelectorAll(".documentation-search-result a")];
}

function selectResult(index) {
	const links = resultLinks();
	activeResultIndex = links.length === 0 ? -1 : (index + links.length) % links.length;
	for (const [linkIndex, link] of links.entries()) {
		if (linkIndex === activeResultIndex) link.dataset.active = "";
		else delete link.dataset.active;
	}
	links[activeResultIndex]?.scrollIntoView({ block: "nearest" });
}

function clearResultSelection() {
	activeResultIndex = -1;
	for (const link of resultLinks()) delete link.dataset.active;
}

async function updateResults() {
	const currentRevision = ++revision;
	const query = input.value.trim();
	list.replaceChildren();
	clearResultSelection();
	empty.hidden = true;

	if (tokenise(query).length === 0) {
		status.textContent = "Search titles, headings, and documentation content.";
		return;
	}

	status.textContent = "Searching documentation...";
	try {
		const index = await loadIndex();
		const groups = groupMatches(searchIndex(index, query));
		const enriched = await Promise.all(groups.map(async group => {
			const match = group[0];
			try {
				const page = await loadPage(match.page[0]);
				return {
					...match,
					more: group.length - 1,
					snippet: snippet(sectionText(page, match.section[1]), match.matchedTerms, match.page[2]),
				};
			} catch (_) {
				return { ...match, more: group.length - 1, snippet: match.page[2] };
			}
		}));
		if (revision !== currentRevision) return;

		for (const match of enriched) {
			const values = {
				page: match.page[1],
				location: sectionLocation(index, match.sectionID),
				hasmore: match.more > 0,
				more: match.more === 1 ? "1 more match on this page" : `${match.more} more matches on this page`,
				path: pathLabel(match.page[0]),
				url: resultURL(match.page, match.section),
			};
			list.append(renderTemplate(
				"DocumentationSearchResult",
				values,
				highlightedSnippet(match.snippet, match.matchedTerms),
			));
		}

		empty.hidden = enriched.length > 0;
		status.textContent = enriched.length === 1
			? "1 documentation result."
			: `${enriched.length} documentation results.`;
		if (enriched.length > 0) selectResult(0);
	} catch (error) {
		status.textContent = "Documentation search is unavailable.";
		console.error("Could not search Sklair documentation", error);
	}
}

function openSearch(trigger) {
	returnFocus = trigger;
	dialog.showModal();
	input.focus();
	loadIndex().catch(() => {
		status.textContent = "Documentation search is unavailable.";
	});
}

for (const trigger of document.querySelectorAll("[data-documentation-search-open]")) {
	trigger.addEventListener("click", () => openSearch(trigger));
}

dialog.querySelector("form").addEventListener("submit", event => event.preventDefault());
dialog.querySelector("[data-documentation-search-close]").addEventListener("click", () => dialog.close());
dialog.addEventListener("click", event => {
	if (event.target === dialog) dialog.close();
});
dialog.addEventListener("close", () => {
	returnFocus?.focus();
});
input.addEventListener("input", updateResults);
input.addEventListener("keydown", event => {
	if (event.key === "Escape") {
		event.preventDefault();
		dialog.close();
	} else if (event.key === "ArrowDown") {
		event.preventDefault();
		selectResult(activeResultIndex + 1);
	} else if (event.key === "ArrowUp") {
		event.preventDefault();
		selectResult(activeResultIndex - 1);
	} else if (event.key === "Enter") {
		const link = resultLinks()[activeResultIndex];
		if (!link) return;
		event.preventDefault();
		link.click();
	}
});

list.addEventListener("pointerover", event => {
	const link = event.target.closest(".documentation-search-result a");
	if (!link) return;
	selectResult(resultLinks().indexOf(link));
});

document.addEventListener("keydown", event => {
	const target = event.target;
	const typing = target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target.isContentEditable;
	const shortcut = (event.key === "/" && !typing && !event.metaKey && !event.ctrlKey)
		|| (event.key.toLocaleLowerCase() === "k" && (event.metaKey || event.ctrlKey));
	if (!shortcut || dialog.open) return;
	event.preventDefault();
	const trigger = [...document.querySelectorAll("[data-documentation-search-open]")]
		.find(candidate => candidate.offsetParent !== null);
	if (trigger) openSearch(trigger);
});

const applePlatform = /Macintosh|Mac OS X|iPhone|iPad|iPod/i.test(navigator.userAgent);
for (const shortcut of document.querySelectorAll("[data-documentation-search-shortcut]")) {
	shortcut.textContent = applePlatform ? "⌘ K" : "Ctrl K";
}

document.documentElement.dataset.documentationSearchReady = "";

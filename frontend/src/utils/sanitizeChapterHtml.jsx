import DOMPurify from "dompurify";

const ALLOWED_TAGS = [
  "p",
  "br",
  "strong",
  "em",
  "u",
  "s",
  "ul",
  "ol",
  "li",
  "blockquote",
  "hr",
  "a",
  "code",
  "pre",
  "h3",
  "h4",
];
const ALLOWED_ATTR = ["href", "title", "target", "rel"];

function normalizeHtml(html) {
  if (!html) return "";

  html = html.replace(/<\/?div(\s[^>]*)?>/gi, (m) =>
    m.startsWith("</") ? "</p>" : "<p>"
  );

  html = html.replace(/<span[^>]*>/gi, "").replace(/<\/span>/gi, "");

  html = html.replace(/(&nbsp;){2,}/gi, "&nbsp;");

  html = html.replace(/<p>(\s|&nbsp;)*<\/p>/gi, "");

  html = html.replace(/<p>\s*<p>/gi, "<p>").replace(/<\/p>\s*<\/p>/gi, "</p>");

  return html.trim();
}

export function sanitizeForBackend(html) {
  const normalized = normalizeHtml(html);
  const clean = DOMPurify.sanitize(normalized, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    FORBID_ATTR: [
      "style",
      "class",
      "id",
      "onclick",
      "onerror",
      "onload",
      "color",
      "bgcolor",
      "face",
      "size",
    ],
    FORBID_TAGS: ["img", "video", "audio", "iframe", "script"],
    ADD_ATTR: ["target", "rel"],
    ALLOW_UNKNOWN_PROTOCOLS: false,
    USE_PROFILES: { html: true },
  });

  const wrapper = document.createElement("div");
  wrapper.innerHTML = clean;
  wrapper.querySelectorAll("a[href]").forEach((a) => {
    try {
      const url = new URL(a.getAttribute("href"), window.location.origin);
      // hanya http/https/mailto
      if (!/^https?:|^mailto:/i.test(url.protocol)) {
        a.removeAttribute("href");
      }
    } catch {
      a.removeAttribute("href");
    }
    a.setAttribute("target", "_blank");
    a.setAttribute("rel", "noopener noreferrer nofollow");
  });

  const finalHtml = wrapper.innerHTML.trim() || "<p><br></p>";
  return finalHtml;
}

export function sanitizeForView(html) {
  const normalized = normalizeHtml(html);
  return DOMPurify.sanitize(normalized, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    FORBID_ATTR: ["style", "class", "id", "onclick", "onerror", "onload"],
    FORBID_TAGS: ["script", "iframe", "video", "audio"],
    ADD_ATTR: ["target", "rel"],
    USE_PROFILES: { html: true },
  });
}

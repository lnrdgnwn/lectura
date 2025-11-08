// Components/Workspace/ChapterEditor.jsx
import { useEffect, useMemo, useRef } from "react";
import { FaBold, FaItalic, FaUndo, FaRedo } from "react-icons/fa";
import { sanitizeForBackend } from "../../../utils/sanitizeChapterHtml";

export default function ChapterEditor({
  value,
  onChange,
  placeholder = "Start writing your story here...",
  minHeight = 340,
  className = "",
}) {
  const editorRef = useRef(null);

  // sinkronisasi nilai dari luar
  useEffect(() => {
    if (editorRef.current && editorRef.current.innerHTML !== (value || "")) {
      editorRef.current.innerHTML = value || "";
    }
  }, [value]);

  const emitSanitized = () => {
    const html = editorRef.current?.innerHTML || "";
    const clean = sanitizeForBackend(html);
    // bila sanitize mengubah, tulis balik supaya DOM bersih
    if (editorRef.current && editorRef.current.innerHTML !== clean) {
      editorRef.current.innerHTML = clean;
      placeCaretAtEnd(editorRef.current);
    }
    onChange?.(clean);
  };

  const onInput = () => {
    emitSanitized();
  };

  // Paste: paksa jadi plain text lalu bungkus <p> dan <br>
  const onPaste = (e) => {
    e.preventDefault();
    const text = e.clipboardData.getData("text/plain");
    if (!text) return;
    document.execCommand("insertText", false, text);
    emitSanitized();
  };

  const exec = (cmd) => {
    editorRef.current?.focus();
    document.execCommand(cmd, false, null);
    emitSanitized();
  };

  const wordCount = useMemo(() => {
    const el = document.createElement("div");
    el.innerHTML = value || "";
    const text = (el.textContent || "").trim();
    return text ? text.split(/\s+/).filter(Boolean).length : 0;
  }, [value]);

  return (
    <div className={`flex flex-col ${className}`}>
      {/* Toolbar */}
      <div className="flex items-center gap-2 mb-2 sticky top-0 bg-white z-10 border-b border-gray-100 pb-2">
        <ToolbarButton title="Bold" onClick={() => exec("bold")}>
          <FaBold />
        </ToolbarButton>
        <ToolbarButton title="Italic" onClick={() => exec("italic")}>
          <FaItalic />
        </ToolbarButton>

        <div className="ml-auto flex items-center gap-2 text-xs text-gray-500">
          <ToolbarButton title="Undo" onClick={() => exec("undo")}>
            <FaUndo />
          </ToolbarButton>
          <ToolbarButton title="Redo" onClick={() => exec("redo")}>
            <FaRedo />
          </ToolbarButton>
          <span className="px-2 py-1 rounded bg-gray-50 border text-gray-600">
            {wordCount} words
          </span>
        </div>
      </div>

      {/* Editor area */}
      <div
        ref={editorRef}
        contentEditable
        suppressContentEditableWarning
        onInput={onInput}
        onPaste={onPaste}
        data-placeholder={placeholder}
        className="border border-gray-200 rounded-lg p-3 md:p-4 text-gray-800 text-sm leading-6 outline-none overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400"
        style={{
          minHeight,
          maxHeight: "60vh",
          wordBreak: "break-word",
        }}
      />
    </div>
  );
}

function ToolbarButton({ title, onClick, children }) {
  return (
    <button
      type="button"
      title={title}
      onClick={onClick}
      className="w-8 h-8 flex items-center justify-center border rounded-md text-gray-700 hover:bg-gray-50 shadow-sm"
    >
      {children}
    </button>
  );
}

// taruh caret di akhir setelah normalisasi
function placeCaretAtEnd(el) {
  try {
    el.focus();
    const range = document.createRange();
    range.selectNodeContents(el);
    range.collapse(false);
    const sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
  } catch {}
}

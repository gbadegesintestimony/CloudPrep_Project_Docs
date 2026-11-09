"use client";

import React, { useState } from "react";
import { Check, Copy, Terminal } from "lucide-react";

interface CodeSnippetProps {
  code: string;
  language?: string;
  inline?: boolean;
}

export function CodeSnippet({ code, language = "bash", inline = false }: CodeSnippetProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
    }
  };

  if (inline) {
    return (
      <code className="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs font-bold text-sky-800 border border-slate-200 shadow-2xs">
        {code}
      </code>
    );
  }

  return (
    <div className="relative group my-3 overflow-hidden rounded-xl border border-slate-200 bg-slate-900 shadow-md">
      <div className="flex items-center justify-between border-b border-slate-800 bg-slate-950 px-3.5 py-1.5 text-[11px] text-slate-400 font-mono">
        <span className="flex items-center gap-1.5 text-slate-400 font-bold">
          <Terminal className="h-3.5 w-3.5 text-sky-400" />
          <span>{language}</span>
        </span>
        <button
          onClick={handleCopy}
          type="button"
          aria-label="Copy code to clipboard"
          className="flex items-center gap-1 text-[10px] font-semibold text-slate-400 hover:text-white transition-colors cursor-pointer"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3 text-emerald-400" />
              <span className="text-emerald-400">Copied</span>
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>

      <pre className="overflow-x-auto p-4 font-mono text-xs text-slate-100 leading-relaxed font-semibold">
        <code>{code}</code>
      </pre>
    </div>
  );
}

// Helper to clean clunky syllabus meta-references and format numbered requirements into clean lists
function cleanStem(raw: string): string {
  if (!raw) return "";
  let s = raw;
  s = s.replace(/Which concept in \d+\.?\s*[^:]+\s+is characterized by the following description:\s*/gi, "Which concept is characterized by the following description: ");
  s = s.replace(/You need to perform the following action in \d+\.?\s*[^:]+:\s*/gi, "You need to perform the following task: ");
  s = s.replace(/In the context of \d+\.?\s*[^,]+,\s*(which statement)/gi, "Which statement");
  s = s.replace(/Regarding (\*\*[^*]+\*\*) in \d+\.?\s*[^,]+,\s*what is its primary/gi, "What is the primary underlying mechanism of $1");
  s = s.replace(/When studying (\*\*[^*]+\*\*) \([^)]+\),\s*which of the following/gi, "For $1, which of the following");
  s = s.replace(/Examine the following command or syntax snippet from \d+\.?\s*[^:]+:\s*/gi, "Examine the following command snippet:\n\n");
  s = s.replace(/During a lab exercise on \d+\.?\s*[^,]+,\s*you need a command/gi, "You need a command");
  s = s.replace(/In \d+\.?\s*[^,]+,\s*how should an engineer decide/gi, "How should a cloud engineer decide");
  s = s.replace(/Which pair of concepts in \d+\.?\s*[^ ]+\s+is distinguished by/gi, "Which pair of concepts is distinguished by");
  s = s.replace(/An architecture review in \d+\.?\s*[^ ]+\s+evaluates/gi, "An architecture team evaluates");

  // Format clustered numbering (1) ... (2) ... into clean separated bullet lines
  s = s.replace(/\s*\(1\)\s*/g, "\n\n• 1. ");
  s = s.replace(/,?\s*(?:and\s*)?\(2\)\s*/g, "\n• 2. ");
  s = s.replace(/,?\s*(?:and\s*)?\(3\)\s*/g, "\n• 3. ");
  s = s.replace(/,?\s*(?:and\s*)?\(4\)\s*/g, "\n• 4. ");

  // Separate scenario requirements, engineering approaches, and the final prompt
  s = s.replace(/\.?\s*To satisfy \*"?([^"]+)"?\*,?\s*the engineering team selects:\s*/gi, ".\n\nGoal: $1\n\nImplemented Solution: ");
  s = s.replace(/\.?\s*What is the primary technical justification for this choice\?/gi, "\n\nQuestion: What is the primary technical justification for this choice?");
  s = s.replace(/\.?\s*(Which solution (?:best satisfies|meets|achieves)[^?]*\?)/gi, "\n\nQuestion: $1");

  return s.trim();
}

// Helper to render text with automatic `code` spans, markdown bold, or ``` code blocks
export function FormattedStem({ text }: { text: string }) {
  if (!text) return null;
  const cleanedText = cleanStem(text);

  // Split by markdown code blocks ```
  if (cleanedText.includes("```")) {
    const parts = cleanedText.split(/(```[\s\S]*?```)/g);
    return (
      <div className="space-y-3 text-sm sm:text-base leading-relaxed font-medium">
        {parts.map((part, index) => {
          if (part.startsWith("```") && part.endsWith("```")) {
            const lines = part.slice(3, -3).trim().split("\n");
            let lang = "bash";
            let codeContent = part.slice(3, -3).trim();
            if (lines.length > 0 && !lines[0].includes(" ") && lines[0].length < 15) {
              lang = lines[0] || "bash";
              codeContent = lines.slice(1).join("\n");
            }
            return <CodeSnippet key={index} code={codeContent} language={lang} />;
          }
          return <FormattedParagraphs key={index} text={part} />;
        })}
      </div>
    );
  }

  return (
    <div className="space-y-2.5 text-sm sm:text-base leading-relaxed font-medium">
      <FormattedParagraphs text={cleanedText} />
    </div>
  );
}

function FormattedParagraphs({ text }: { text: string }) {
  const paragraphs = text.split(/\n\n+/);
  return (
    <>
      {paragraphs.map((p, pi) => {
        const lines = p.split(/\n/);
        return (
          <div key={pi} className="space-y-1.5">
            {lines.map((line, li) => {
              const trimmed = line.trim();
              if (trimmed.startsWith("• ")) {
                return (
                  <div key={li} className="flex items-start gap-2 pl-3 py-1 bg-sky-50/60 dark:bg-slate-800/60 rounded-lg border border-sky-100 dark:border-slate-700 my-1">
                    <span className="text-sky-600 dark:text-emerald-400 font-bold">•</span>
                    <div className="flex-1">
                      <FormattedInlineText text={trimmed.slice(2)} />
                    </div>
                  </div>
                );
              }
              return (
                <p key={li} className="leading-relaxed">
                  <FormattedInlineText text={line} />
                </p>
              );
            })}
          </div>
        );
      })}
    </>
  );
}

function FormattedInlineText({ text }: { text: string }) {
  if (!text) return null;
  // Clean all stray markdown artifacts and asterisks
  const sanitized = text
    .replace(/\*\*/g, "")
    .replace(/\*\"/g, '"')
    .replace(/\"\*/g, '"')
    .replace(/\*([a-zA-Z0-9_\- /]+)\*/g, "$1");

  const parts = sanitized.split(/(`[^`]+`)/g);
  return (
    <span>
      {parts.map((p, i) => {
        if (p.startsWith("`") && p.endsWith("`") && p.length > 2) {
          return <CodeSnippet key={i} code={p.slice(1, -1)} inline />;
        }
        return <span key={i}>{p}</span>;
      })}
    </span>
  );
}

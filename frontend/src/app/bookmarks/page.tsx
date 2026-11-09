"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Bookmark,
  BookOpen,
  ArrowRight,
  Trash2,
  Filter,
  Layers,
  HelpCircle,
  Loader2,
  Sparkles,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import { FormattedStem } from "@/components/CodeSnippet";
import {
  fetchBookmarks,
  toggleBookmark,
  createPracticeSession,
  BookmarkDTO,
  getStoredAccessToken,
} from "@/lib/api";

// Cleans legacy syllabus or PDF note references from explanations
function cleanExplanation(raw: string): string {
  if (!raw) return "";
  let s = raw;
  s = s.replace(/^In Module \d+,\s*Week \d+\s*—\s*[^,]+(?:\([^)]+\))?,\s*/i, "");
  s = s.replace(/^According to Module \d+,\s*Week \d+\s*—\s*[^,]+(?:\([^)]+\))?,\s*/i, "");
  s = s.replace(/^According to [^,]+,\s*/i, "");
  s = s.replace(/^In [^,]+,\s*(?=\*\*)/i, "");
  s = s.replace(/^Per [^,]+,\s*/i, "");
  s = s.replace(/\s*the course notes \([^)]+\) highlight/i, " standard cloud best practice emphasizes");
  s = s.replace(/\s*\([A-Za-z0-9_ -]+\.pdf\)/gi, "");
  s = s.replace(/\s*\(Module \d+,\s*Week \d+[^)]*\)/gi, "");
  return s.trim();
}

export default function BookmarksPage() {
  const router = useRouter();
  const [bookmarks, setBookmarks] = useState<BookmarkDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterTopic, setFilterTopic] = useState("ALL");
  const [filterStyle, setFilterStyle] = useState("ALL");
  const [startingRevision, setStartingRevision] = useState(false);

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }

    fetchBookmarks()
      .then(setBookmarks)
      .finally(() => setLoading(false));
  }, [router]);

  const handleRemoveBookmark = async (questionId: string) => {
    try {
      await toggleBookmark(questionId, true);
      setBookmarks((prev) => prev.filter((b) => b.questionId !== questionId));
    } catch {
      alert("Failed to remove bookmark");
    }
  };

  const handleLaunchRevisionQuiz = async () => {
    if (bookmarks.length === 0) return;
    try {
      setStartingRevision(true);
      const res = await createPracticeSession({
        mode: "BOOKMARK_REVISION",
        questionCount: Math.min(bookmarks.length, 25),
      });
      router.push(`/practice?sessionId=${res.session.id}`);
    } catch (err: any) {
      alert(err?.message || "Failed to start revision quiz");
    } finally {
      setStartingRevision(false);
    }
  };

  const topicsList = Array.from(new Set(bookmarks.map((b) => b.topicTitle || b.topicId)));
  const filtered = bookmarks.filter((b) => {
    if (filterTopic !== "ALL" && (b.topicTitle || b.topicId) !== filterTopic) return false;
    if (filterStyle !== "ALL" && b.questionStyle !== filterStyle) return false;
    return true;
  });

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-sky-100 pb-6">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-amber-600">
                Spaced Repetition &amp; Review
              </span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3 font-display">
              <Bookmark className="h-7 w-7 text-amber-500" />
              Saved Question Bookmarks ({bookmarks.length})
            </h1>
            <p className="text-sm text-slate-600 leading-relaxed font-normal">
              Review flagged items, inspect distractor rationales, and launch focused revision quizzes on difficult concepts.
            </p>
          </div>

          {bookmarks.length > 0 && (
            <button
              onClick={handleLaunchRevisionQuiz}
              disabled={startingRevision}
              className="flex items-center gap-2 rounded-xl bg-amber-600 px-5 py-2.5 text-xs font-semibold text-white hover:bg-amber-500 shadow-md shadow-amber-600/25 transition-all shrink-0 disabled:opacity-50"
            >
              {startingRevision ? (
                <Loader2 className="h-4 w-4 animate-spin text-white" />
              ) : (
                <Sparkles className="h-4 w-4" />
              )}
              <span>Practice Bookmarked Items</span>
            </button>
          )}
        </div>

        {/* Filters */}
        {bookmarks.length > 0 && (
          <div className="flex flex-wrap items-center gap-4 bg-white/95 p-4 rounded-2xl border border-sky-100 text-xs shadow-sm">
            <span className="font-semibold text-slate-600 flex items-center gap-1.5">
              <Filter className="h-3.5 w-3.5 text-sky-600" /> Filter By:
            </span>

            <select
              value={filterTopic}
              onChange={(e) => setFilterTopic(e.target.value)}
              className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs text-slate-800 focus:border-sky-500 focus:bg-white focus:outline-none"
            >
              <option value="ALL">All Topics ({bookmarks.length})</option>
              {topicsList.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>

            <select
              value={filterStyle}
              onChange={(e) => setFilterStyle(e.target.value)}
              className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs text-slate-800 focus:border-sky-500 focus:bg-white focus:outline-none"
            >
              <option value="ALL">All Question Styles</option>
              <option value="DEFINITION">Definition</option>
              <option value="COMMAND_SYNTAX">Command Syntax</option>
              <option value="COMPARISON">Comparison</option>
              <option value="SCENARIO">Scenario</option>
            </select>
          </div>
        )}

        {/* Content */}
        {loading ? (
          <div className="flex flex-col items-center justify-center min-h-[40vh] gap-3">
            <Loader2 className="h-6 w-6 animate-spin text-sky-600" />
            <p className="text-xs text-slate-500 font-medium">Loading your saved bookmarks...</p>
          </div>
        ) : bookmarks.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-sky-200 bg-white/50 p-12 text-center space-y-4">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-amber-50 text-amber-500 border border-amber-200">
              <Bookmark className="h-6 w-6" />
            </div>
            <div className="space-y-1">
              <h3 className="text-base font-bold text-slate-900 font-display">No bookmarked questions yet</h3>
              <p className="text-xs text-slate-600 max-w-sm mx-auto font-normal">
                While practicing or taking mock exams, click the Bookmark icon on any tricky question to save it here for targeted review.
              </p>
            </div>
            <Link
              href="/practice"
              className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-2.5 text-xs font-semibold text-white hover:from-sky-600 hover:to-blue-700 transition-all shadow-md shadow-sky-500/25"
            >
              <BookOpen className="h-4 w-4" /> Start Practicing Now
            </Link>
          </div>
        ) : (
          <div className="space-y-6">
            {filtered.map((b, idx) => (
              <div
                key={b.questionId}
                className="rounded-2xl border border-sky-100 bg-white/95 p-5 sm:p-6 space-y-4 shadow-lg shadow-sky-100/60 text-slate-900"
              >
                <div className="flex items-center justify-between border-b border-slate-100 pb-3 text-xs">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-slate-400 font-bold">#{idx + 1}</span>
                    <span className="rounded-full bg-sky-50 px-2.5 py-0.5 text-[10px] font-mono text-sky-700 font-semibold border border-sky-200">
                      {b.questionStyle}
                    </span>
                    <span className="text-slate-600 font-semibold">{b.topicTitle}</span>
                  </div>

                  <button
                    onClick={() => handleRemoveBookmark(b.questionId)}
                    title="Remove bookmark"
                    className="flex items-center gap-1 text-[11px] text-slate-400 hover:text-rose-600 transition-colors"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    <span>Remove</span>
                  </button>
                </div>

                {/* Stem */}
                <div className="py-1 text-slate-900">
                  <FormattedStem text={b.stem} />
                </div>

                {/* Options List */}
                <div className="space-y-2 pt-2">
                  {b.options.map((opt) => {
                    const isCorrect = opt.key === b.correctOptionKey;
                    return (
                      <div
                        key={opt.key}
                        className={`rounded-xl border p-3 text-xs flex items-start gap-3 shadow-sm ${
                          isCorrect
                            ? "border-emerald-300 bg-emerald-50 text-emerald-950 font-medium"
                            : "border-slate-200 bg-slate-50/70 text-slate-700"
                        }`}
                      >
                        <span
                          className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-lg font-bold text-[11px] ${
                            isCorrect ? "bg-emerald-600 text-white" : "bg-slate-200 text-slate-600"
                          }`}
                        >
                          {opt.key}
                        </span>
                        <div className="flex-1 pt-0.5">
                          <FormattedStem text={opt.text} />
                          {opt.distractorRationale && !isCorrect && (
                            <p className="mt-1 text-[11px] text-slate-500 italic">
                              Why incorrect: {opt.distractorRationale}
                            </p>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>

                {/* Pedagogical Explanation */}
                {b.explanation && (
                  <div className="mt-4 rounded-xl border border-sky-200 bg-sky-50/80 p-4 text-xs space-y-1 shadow-sm">
                    <span className="font-semibold text-sky-700 flex items-center gap-1.5">
                      <HelpCircle className="h-3.5 w-3.5 text-sky-600" /> Technical Explanation
                    </span>
                    <div className="text-slate-800 leading-relaxed font-normal">
                      <FormattedStem text={cleanExplanation(b.explanation)} />
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}


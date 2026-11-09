"use client";

import React, { useState } from "react";
import Link from "next/link";
import {
  Bookmark,
  Flag,
  CheckCircle2,
  XCircle,
  HelpCircle,
  BookOpen,
  Sparkles,
} from "lucide-react";
import { FormattedStem } from "./CodeSnippet";
import {
  ActiveSessionQuestionDTO,
  ReviewedQuestionDTO,
  toggleBookmark,
} from "@/lib/api";

interface QuestionCardProps {
  question: ActiveSessionQuestionDTO | ReviewedQuestionDTO;
  currentIndex: number;
  totalQuestions: number;
  selectedOptionKey?: string;
  isImmediateMode?: boolean;
  isReviewMode?: boolean;
  onSelectOption?: (optionKey: string) => void;
  onToggleFlag?: () => void;
}

// Cleans any legacy syllabus or PDF note references from explanations
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

export function QuestionCard({
  question,
  currentIndex,
  totalQuestions,
  selectedOptionKey,
  isImmediateMode = false,
  isReviewMode = false,
  onSelectOption,
  onToggleFlag,
}: QuestionCardProps) {
  const [isBookmarked, setIsBookmarked] = useState(question.isBookmarked);
  const [bookmarking, setBookmarking] = useState(false);

  const handleBookmarkToggle = async () => {
    try {
      setBookmarking(true);
      const newState = await toggleBookmark(question.questionId, isBookmarked);
      setIsBookmarked(newState);
    } catch {
      // rollback
    } finally {
      setBookmarking(false);
    }
  };

  const currentSelectedKey = selectedOptionKey || question.selectedOptionKey;
  const isAnswered = !!currentSelectedKey;
  const hasCorrectKey = !!question.correctOptionKey;
  // In immediate mode, show green/red feedback once the server has verified correctness
  const showFeedback = (isImmediateMode && isAnswered && hasCorrectKey) || isReviewMode;

  const styleBadgeStyles: Record<string, string> = {
    DEFINITION: "bg-sky-50 text-sky-700 border-sky-200",
    COMMAND_SYNTAX: "bg-indigo-50 text-indigo-700 border-indigo-200 font-mono",
    COMPARISON: "bg-purple-50 text-purple-700 border-purple-200",
    SCENARIO: "bg-emerald-50 text-emerald-700 border-emerald-200",
  };

  return (
    <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 sm:p-7 shadow-lg shadow-sky-100/60 space-y-6 text-slate-900">
      {/* Question Header & Meta */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 pb-4">
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs font-bold text-slate-500 font-mono">
            Q{currentIndex + 1} of {totalQuestions}
          </span>
          <span
            className={`rounded-full px-2.5 py-0.5 text-[10px] font-semibold border ${
              styleBadgeStyles[question.questionStyle] || "bg-slate-100 text-slate-700 border-slate-200"
            }`}
          >
            {question.questionStyle}
          </span>
          <span className="rounded-full bg-slate-100 px-2.5 py-0.5 text-[10px] font-medium text-slate-600 border border-slate-200">
            {question.difficulty}
          </span>
          {question.sourceReference && (
            <span className="rounded-full bg-slate-50 px-2.5 py-0.5 text-[10px] font-mono text-slate-500 border border-slate-200 truncate max-w-[260px]">
              {question.sourceReference}
            </span>
          )}
        </div>

        {/* Action Buttons: Flag & Bookmark */}
        <div className="flex items-center gap-1.5">
          {onToggleFlag && (
            <button
              onClick={onToggleFlag}
              type="button"
              title="Flag question for review"
              className={`flex items-center gap-1 rounded-xl px-2.5 py-1 text-xs font-medium border transition-colors shadow-sm ${
                question.isFlagged
                  ? "bg-amber-50 text-amber-800 border-amber-300"
                  : "bg-white text-slate-600 border-slate-200 hover:text-slate-900 hover:bg-slate-50"
              }`}
            >
              <Flag className={`h-3.5 w-3.5 ${question.isFlagged ? "fill-amber-500 text-amber-500" : ""}`} />
              <span className="hidden sm:inline">Flag</span>
            </button>
          )}

          <button
            onClick={handleBookmarkToggle}
            disabled={bookmarking}
            type="button"
            title="Save to bookmarks"
            className={`flex items-center gap-1 rounded-xl px-2.5 py-1 text-xs font-medium border transition-colors shadow-sm ${
              isBookmarked
                ? "bg-sky-50 text-sky-700 border-sky-300"
                : "bg-white text-slate-600 border-slate-200 hover:text-slate-900 hover:bg-slate-50"
            }`}
          >
            <Bookmark className={`h-3.5 w-3.5 ${isBookmarked ? "fill-sky-500 text-sky-500" : ""}`} />
            <span className="hidden sm:inline">{isBookmarked ? "Saved" : "Bookmark"}</span>
          </button>
        </div>
      </div>

      {/* Question Stem */}
      <div className="py-1 text-slate-900">
        <FormattedStem text={question.stem} />
      </div>

      {/* Options List */}
      <div className="space-y-3 pt-2">
        {question.options.map((option) => {
          const isSelected = (selectedOptionKey || question.selectedOptionKey) === option.key;
          const isCorrect = showFeedback && question.correctOptionKey === option.key;
          const isIncorrectChoice = showFeedback && isSelected && !isCorrect;

          let optionStyle =
            "border-slate-200 bg-white text-slate-900 hover:border-sky-300 hover:bg-sky-50/50 shadow-xs font-medium";
          let badgeStyle = "bg-slate-100 text-slate-800 border-slate-300 font-bold";

          if (showFeedback) {
            if (isCorrect) {
              optionStyle = "border-emerald-500 bg-emerald-50 text-emerald-950 font-medium shadow-sm";
              badgeStyle = "bg-emerald-500 text-white font-bold border-emerald-400";
            } else if (isIncorrectChoice) {
              optionStyle = "border-rose-400 bg-rose-50 text-rose-950 shadow-sm";
              badgeStyle = "bg-rose-500 text-white font-bold border-rose-400";
            } else {
              optionStyle = "border-slate-200 bg-slate-50/40 text-slate-400 opacity-60";
            }
          } else if (isSelected) {
            optionStyle = "border-sky-500 bg-sky-50 text-sky-950 shadow-md shadow-sky-100 font-medium";
            badgeStyle = "bg-sky-600 text-white font-bold border-sky-500";
          }

          return (
            <div key={option.key} className="space-y-1.5">
              <button
                type="button"
                onClick={() => !showFeedback && onSelectOption?.(option.key)}
                disabled={showFeedback && !isReviewMode}
                className={`w-full flex items-start gap-3.5 rounded-xl border p-3.5 sm:p-4 text-left transition-all ${optionStyle}`}
              >
                <div
                  className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border text-xs font-bold transition-all ${badgeStyle}`}
                >
                  {showFeedback && isCorrect ? (
                    <CheckCircle2 className="h-4 w-4" />
                  ) : showFeedback && isIncorrectChoice ? (
                    <XCircle className="h-4 w-4" />
                  ) : (
                    option.key
                  )}
                </div>

                <div className="flex-1 text-xs sm:text-sm font-semibold text-slate-900 leading-relaxed pt-0.5">
                  <FormattedStem text={option.text} />
                </div>
              </button>

              {/* Distractor Rationale (Shown on incorrect option in Feedback mode) */}
              {showFeedback && option.distractorRationale && (
                <div className="ml-10 text-[11px] text-slate-600 bg-slate-100 border-l-2 border-slate-300 px-3 py-1.5 rounded-r-lg">
                  <span className="font-semibold text-slate-800">Why {option.key} is incorrect: </span>
                  {option.distractorRationale}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Detailed Pedagogical Explanation (Pure Technical Reason) */}
      {showFeedback && question.explanation && (
        <div className="mt-6 rounded-xl border border-sky-200 bg-sky-50/80 p-4 sm:p-5 text-xs sm:text-sm space-y-2 shadow-sm">
          <div className="flex items-center justify-between text-sky-700 font-semibold">
            <span className="flex items-center gap-1.5 text-xs uppercase tracking-wider font-bold">
              <HelpCircle className="h-4 w-4 text-sky-600" /> Explanation
            </span>
            {question.correctOptionKey && (
              <span className="rounded-full bg-emerald-100 text-emerald-800 border border-emerald-200 px-2.5 py-0.5 text-xs font-bold">
                Correct Answer: {question.correctOptionKey}
              </span>
            )}
          </div>
          <div className="text-slate-800 leading-relaxed pt-1 font-normal text-sm">
            <FormattedStem text={cleanExplanation(question.explanation)} />
          </div>
        </div>
      )}
    </div>
  );
}


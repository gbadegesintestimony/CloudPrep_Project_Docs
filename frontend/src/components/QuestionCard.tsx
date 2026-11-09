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
  Terminal,
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

  const isCommandQuestion =
    question.questionStyle === "COMMAND_SYNTAX" ||
    question.questionStyle === "COMMAND" ||
    question.options.some((o) =>
      /^(sudo |chmod |chown |docker |git |aws |curl |kubectl |systemctl |df |kill |grep |awk |sed |cat |echo |export |ssh |find |tar |if \[)/.test(
        o.text.trim()
      )
    );

  const styleBadgeStyles: Record<string, string> = {
    DEFINITION: "bg-sky-50 text-sky-700 border-sky-200",
    COMMAND_SYNTAX: "bg-slate-900 text-emerald-400 border-slate-700 font-mono shadow-xs",
    COMPARISON: "bg-purple-50 text-purple-700 border-purple-200",
    SCENARIO: "bg-emerald-50 text-emerald-700 border-emerald-200",
  };

  return (
    <div
      className={`rounded-2xl border p-5 sm:p-7 shadow-xl space-y-6 transition-all ${
        isCommandQuestion
          ? "border-slate-800 bg-[#0b1120] text-slate-100 shadow-2xl shadow-slate-950/70"
          : "border-sky-100 bg-white/95 text-slate-900 shadow-sky-100/60"
      }`}
    >
      {/* Question Header & Meta */}
      <div
        className={`flex flex-wrap items-center justify-between gap-3 border-b pb-4 ${
          isCommandQuestion ? "border-slate-800/80" : "border-slate-100"
        }`}
      >
        <div className="flex flex-wrap items-center gap-2">
          {isCommandQuestion && (
            <div className="flex items-center gap-1.5 mr-1.5 select-none">
              <span className="h-3 w-3 rounded-full bg-rose-500/90 shadow-sm" />
              <span className="h-3 w-3 rounded-full bg-amber-400/90 shadow-sm" />
              <span className="h-3 w-3 rounded-full bg-emerald-500/90 shadow-sm" />
            </div>
          )}
          <span
            className={`text-xs font-bold font-mono ${
              isCommandQuestion ? "text-slate-400" : "text-slate-500"
            }`}
          >
            Q{currentIndex + 1} of {totalQuestions}
          </span>
          <span
            className={`rounded-full px-2.5 py-0.5 text-[10px] font-semibold border ${
              isCommandQuestion
                ? "bg-emerald-950/80 text-emerald-400 border-emerald-800/60 font-mono flex items-center gap-1"
                : styleBadgeStyles[question.questionStyle] ||
                  "bg-slate-100 text-slate-700 border-slate-200"
            }`}
          >
            {isCommandQuestion ? (
              <>
                <Terminal className="h-3 w-3 text-emerald-400" />
                <span>bash — cloudprep@terminal:~</span>
              </>
            ) : (
              question.questionStyle
            )}
          </span>
          <span
            className={`rounded-full px-2.5 py-0.5 text-[10px] font-medium border ${
              isCommandQuestion
                ? "bg-slate-800/90 text-slate-300 border-slate-700 font-mono"
                : "bg-slate-100 text-slate-600 border-slate-200"
            }`}
          >
            {question.difficulty}
          </span>
          {question.sourceReference && (
            <span
              className={`rounded-full px-2.5 py-0.5 text-[10px] font-mono border truncate max-w-[260px] ${
                isCommandQuestion
                  ? "bg-slate-800/60 text-slate-400 border-slate-700/80"
                  : "bg-slate-50 text-slate-500 border-slate-200"
              }`}
            >
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
                  ? "bg-amber-500/20 text-amber-300 border-amber-500/50"
                  : isCommandQuestion
                  ? "bg-slate-800 text-slate-300 border-slate-700 hover:text-white hover:bg-slate-700"
                  : "bg-white text-slate-600 border-slate-200 hover:text-slate-900 hover:bg-slate-50"
              }`}
            >
              <Flag
                className={`h-3.5 w-3.5 ${
                  question.isFlagged ? "fill-amber-500 text-amber-500" : ""
                }`}
              />
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
                ? "bg-sky-500/20 text-sky-300 border-sky-500/50"
                : isCommandQuestion
                ? "bg-slate-800 text-slate-300 border-slate-700 hover:text-white hover:bg-slate-700"
                : "bg-white text-slate-600 border-slate-200 hover:text-slate-900 hover:bg-slate-50"
            }`}
          >
            <Bookmark
              className={`h-3.5 w-3.5 ${
                isBookmarked ? "fill-sky-500 text-sky-500" : ""
              }`}
            />
            <span className="hidden sm:inline">{isBookmarked ? "Saved" : "Bookmark"}</span>
          </button>
        </div>
      </div>

      {/* Question Stem */}
      {isCommandQuestion ? (
        <div className="overflow-hidden rounded-xl border border-slate-800 bg-[#080d1a] p-4 sm:p-5 font-mono text-xs sm:text-sm leading-relaxed space-y-3">
          <div className="flex items-start gap-2.5">
            <span className="select-none font-bold text-emerald-400 shrink-0">$</span>
            <div className="flex-1 text-slate-100 font-medium">
              <FormattedStem text={question.stem} />
            </div>
          </div>
          <div className="flex items-center gap-2 text-[11px] text-slate-500 pt-2 border-t border-slate-800/70 select-none">
            <span className="text-emerald-400 font-bold">#</span>
            <span>Select the correct command syntax from the options below:</span>
          </div>
        </div>
      ) : (
        <div className="py-1 text-slate-900">
          <FormattedStem text={question.stem} />
        </div>
      )}

      {/* Options List */}
      <div className="space-y-3 pt-2">
        {question.options.map((option) => {
          const isSelected =
            (selectedOptionKey || question.selectedOptionKey) === option.key;
          const isCorrect = showFeedback && question.correctOptionKey === option.key;
          const isIncorrectChoice = showFeedback && isSelected && !isCorrect;

          let optionStyle =
            "border-slate-200 bg-white text-slate-900 hover:border-sky-300 hover:bg-sky-50/50 shadow-xs font-medium";
          let badgeStyle = "bg-slate-100 text-slate-800 border-slate-300 font-bold";

          if (isCommandQuestion) {
            if (showFeedback) {
              if (isCorrect) {
                optionStyle =
                  "border-emerald-500 bg-emerald-950/80 text-emerald-100 font-mono shadow-md ring-1 ring-emerald-500/50";
                badgeStyle =
                  "bg-emerald-500 text-white font-bold border-emerald-400 font-mono";
              } else if (isIncorrectChoice) {
                optionStyle =
                  "border-rose-500 bg-rose-950/80 text-rose-100 font-mono shadow-md ring-1 ring-rose-500/50";
                badgeStyle =
                  "bg-rose-500 text-white font-bold border-rose-400 font-mono";
              } else {
                optionStyle =
                  "border-slate-800 bg-[#080d1a] text-slate-500 opacity-50 font-mono";
                badgeStyle =
                  "bg-slate-900 text-slate-600 border-slate-800 font-mono";
              }
            } else if (isSelected) {
              optionStyle =
                "border-emerald-400 bg-[#092622] text-emerald-100 shadow-xl shadow-emerald-950/40 font-mono ring-2 ring-emerald-400/50";
              badgeStyle =
                "bg-emerald-500 text-slate-950 font-bold border-emerald-400 font-mono";
            } else {
              optionStyle =
                "border-slate-800 bg-[#0f172a] text-slate-200 hover:border-emerald-500 hover:bg-[#162238] shadow-sm font-mono";
              badgeStyle =
                "bg-slate-800 text-slate-300 border-slate-700 font-bold font-mono";
            }
          } else {
            if (showFeedback) {
              if (isCorrect) {
                optionStyle =
                  "border-emerald-500 bg-emerald-50 text-emerald-950 font-medium shadow-sm";
                badgeStyle =
                  "bg-emerald-500 text-white font-bold border-emerald-400";
              } else if (isIncorrectChoice) {
                optionStyle =
                  "border-rose-400 bg-rose-50 text-rose-950 shadow-sm";
                badgeStyle =
                  "bg-rose-500 text-white font-bold border-rose-400";
              } else {
                optionStyle =
                  "border-slate-200 bg-slate-50/40 text-slate-400 opacity-60";
              }
            } else if (isSelected) {
              optionStyle =
                "border-sky-500 bg-sky-50 text-sky-950 shadow-md shadow-sky-100 font-medium";
              badgeStyle = "bg-sky-600 text-white font-bold border-sky-500";
            }
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

                <div
                  className={`flex-1 text-xs sm:text-sm ${
                    isCommandQuestion
                      ? "font-mono font-medium text-slate-100 flex items-center"
                      : "font-semibold text-slate-900"
                  } leading-relaxed pt-0.5`}
                >
                  {isCommandQuestion && (
                    <span className="text-emerald-400 font-bold mr-2 select-none shrink-0">
                      $
                    </span>
                  )}
                  <div className="flex-1">
                    <FormattedStem text={option.text} />
                  </div>
                </div>
              </button>

              {/* Distractor Rationale (Shown on incorrect option in Feedback mode) */}
              {showFeedback && option.distractorRationale && (
                <div
                  className={`ml-10 text-[11px] px-3.5 py-2 rounded-r-lg border-l-2 ${
                    isCommandQuestion
                      ? "bg-[#080d1a] border-slate-700 text-slate-300 font-mono"
                      : "bg-slate-100 border-slate-300 text-slate-600"
                  }`}
                >
                  <span
                    className={`font-semibold ${
                      isCommandQuestion ? "text-slate-300" : "text-slate-800"
                    }`}
                  >
                    Why {option.key} is incorrect:{" "}
                  </span>
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


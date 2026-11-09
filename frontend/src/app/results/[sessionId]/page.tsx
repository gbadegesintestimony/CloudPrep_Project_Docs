"use client";

export const dynamic = "force-dynamic";

import React, { useEffect, useState, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Award,
  CheckCircle2,
  XCircle,
  Clock,
  ArrowRight,
  TrendingUp,
  Bookmark,
  BookOpen,
  Filter,
  Layers,
  ChevronDown,
  Loader2,
  AlertTriangle,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import { QuestionCard } from "@/components/QuestionCard";
import {
  getSessionReview,
  PracticeSessionDTO,
  ReviewedQuestionDTO,
  getStoredAccessToken,
} from "@/lib/api";

export default function ResultsPage({ params }: { params: Promise<{ sessionId: string }> }) {
  const router = useRouter();
  const resolvedParams = use(params);
  const sessionId = resolvedParams.sessionId;

  const [session, setSession] = useState<PracticeSessionDTO | null>(null);
  const [results, setResults] = useState<ReviewedQuestionDTO[]>([]);
  const [filterMode, setFilterMode] = useState<"ALL" | "INCORRECT" | "FLAGGED" | "BOOKMARKED">("ALL");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }

    async function loadResults() {
      try {
        const data = await getSessionReview(sessionId);
        setSession(data.session);
        const list = data.results || (data as any).questions || [];
        setResults(list);
      } catch (err: any) {
        alert(err?.message || "Failed to load exam results.");
      } finally {
        setLoading(false);
      }
    }

    loadResults();
  }, [sessionId, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />
        <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-sky-600" />
          <p className="text-sm text-slate-600">Loading your comprehensive score report...</p>
        </div>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />
        <div className="mx-auto max-w-md text-center py-20 space-y-4">
          <h2 className="font-display text-xl font-bold text-slate-900">Exam Session Not Found</h2>
          <Link href="/dashboard" className="text-xs font-semibold text-sky-600 underline">
            Return to Dashboard
          </Link>
        </div>
      </div>
    );
  }

  const finalScore =
    session.scorePercentage ??
    (session as any).scorePercent ??
    (session.totalQuestions > 0
      ? Math.round(((session.correctCount ?? 0) / session.totalQuestions) * 100)
      : 0);
  const isPassed = finalScore >= 75;
  const questionList = Array.isArray(results) ? results : [];
  const filteredQuestions = questionList.filter((q) => {
    if (filterMode === "INCORRECT") return q.isCorrect === false;
    if (filterMode === "FLAGGED") return q.isFlagged;
    if (filterMode === "BOOKMARKED") return q.isBookmarked;
    return true;
  });

  // Calculate Topic Breakdown
  const topicStats: Record<string, { title: string; total: number; correct: number }> = {};
  questionList.forEach((q) => {
    if (!topicStats[q.topicId]) {
      topicStats[q.topicId] = { title: q.topicTitle || q.topicId, total: 0, correct: 0 };
    }
    topicStats[q.topicId].total++;
    if (q.isCorrect) topicStats[q.topicId].correct++;
  });

  // Calculate Question Style Breakdown
  const styleStats: Record<string, { total: number; correct: number }> = {};
  questionList.forEach((q) => {
    if (!styleStats[q.questionStyle]) {
      styleStats[q.questionStyle] = { total: 0, correct: 0 };
    }
    styleStats[q.questionStyle].total++;
    if (q.isCorrect) styleStats[q.questionStyle].correct++;
  });

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-6xl mx-auto">
        {/* Score Report Hero Card */}
        <div
          className={`relative overflow-hidden rounded-2xl border p-6 sm:p-8 backdrop-blur-xl ${
            isPassed
              ? "border-emerald-200 bg-gradient-to-br from-emerald-50/90 via-white to-sky-50/70 shadow-xl shadow-emerald-100/50"
              : "border-amber-200 bg-gradient-to-br from-amber-50/90 via-white to-sky-50/70 shadow-xl shadow-amber-100/50"
          }`}
        >
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-bold border ${
                    isPassed
                      ? "border-emerald-300 bg-emerald-100 text-emerald-800"
                      : "border-amber-300 bg-amber-100 text-amber-800"
                  }`}
                >
                  {isPassed ? "PASSED BENCHMARK" : "STUDY RECOMMENDED"}
                </span>
                <span className="text-xs text-slate-500 font-mono">{session.title}</span>
              </div>
              <h1 className="font-display text-3xl sm:text-4xl font-extrabold tracking-tight text-slate-900">
                Exam Score: {finalScore}%
              </h1>
              <p className="text-sm text-slate-600">
                You correctly answered <strong className="text-slate-900">{session.correctCount}</strong> out of{" "}
                <strong className="text-slate-900">{session.totalQuestions}</strong> questions.
              </p>
            </div>

            <div className="flex items-center gap-3 shrink-0">
              <Link
                href="/exams"
                className="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-slate-900 transition-all shadow-sm"
              >
                Retake Exam
              </Link>
              <Link
                href="/practice"
                className="rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-4 py-2.5 text-xs font-semibold text-white hover:from-sky-600 hover:to-indigo-700 transition-all shadow-md shadow-sky-500/25"
              >
                Target Weak Topics
              </Link>
            </div>
          </div>
        </div>

        {/* Diagnostic Breakdowns (Topic & Question Style) */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Topic Performance Matrix */}
          <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-4 shadow-lg shadow-sky-100/50 backdrop-blur-xl">
            <h3 className="font-display text-xs font-bold uppercase tracking-wider text-slate-700 flex items-center gap-2">
              <Layers className="h-4 w-4 text-sky-600" /> Topic Performance Breakdown
            </h3>
            <div className="space-y-3">
              {Object.entries(topicStats).map(([topicId, st]) => {
                const pct = Math.round((st.correct / st.total) * 100);
                const isGood = pct >= 75;
                return (
                  <div key={topicId} className="space-y-1">
                    <div className="flex justify-between text-xs font-medium">
                      <span className="text-slate-700 truncate max-w-[280px]">{st.title}</span>
                      <span className={isGood ? "text-emerald-600 font-mono font-bold" : "text-amber-600 font-mono font-bold"}>
                        {st.correct}/{st.total} ({pct}%)
                      </span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-slate-100 overflow-hidden border border-slate-200/50">
                      <div
                        className={`h-full rounded-full ${isGood ? "bg-emerald-500" : "bg-amber-500"}`}
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Question Style Breakdown */}
          <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-4 shadow-lg shadow-sky-100/50 backdrop-blur-xl">
            <h3 className="font-display text-xs font-bold uppercase tracking-wider text-slate-700 flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-indigo-600" /> Question Style Accuracy
            </h3>
            <div className="space-y-3">
              {Object.entries(styleStats).map(([style, st]) => {
                const pct = Math.round((st.correct / st.total) * 100);
                return (
                  <div key={style} className="space-y-1">
                    <div className="flex justify-between text-xs font-medium">
                      <span className="text-slate-700 font-mono text-[11px] font-semibold">{style}</span>
                      <span className="font-mono text-slate-700 font-bold">
                        {st.correct}/{st.total} ({pct}%)
                      </span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-slate-100 overflow-hidden border border-slate-200/50">
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-sky-500 to-indigo-600"
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Detailed Question Review List */}
        <div className="space-y-6 pt-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-sky-100 pb-4">
            <div>
              <h2 className="font-display text-lg font-bold text-slate-900 flex items-center gap-2">
                <BookOpen className="h-5 w-5 text-sky-600" />
                Question-by-Question Review &amp; Distractor Rationales
              </h2>
              <p className="text-xs text-slate-600">
                Inspect correct answers, distractor rationales, and lecture notes citations for every question.
              </p>
            </div>

            {/* Filter Toggle Buttons */}
            <div className="flex items-center gap-1.5 rounded-xl border border-slate-200 bg-slate-50 p-1">
              {(
                [
                  { id: "ALL", label: `All (${results.length})` },
                  { id: "INCORRECT", label: `Incorrect (${results.filter((q) => q.isCorrect === false).length})` },
                  { id: "FLAGGED", label: `Flagged (${results.filter((q) => q.isFlagged).length})` },
                ] as const
              ).map((f) => (
                <button
                  key={f.id}
                  type="button"
                  onClick={() => setFilterMode(f.id)}
                  className={`rounded-lg px-2.5 py-1 text-xs font-semibold transition-all ${
                    filterMode === f.id
                      ? "bg-sky-600 text-white shadow-sm"
                      : "text-slate-600 hover:text-slate-900"
                  }`}
                >
                  {f.label}
                </button>
              ))}
            </div>
          </div>

          {/* Filtered Question Cards */}
          <div className="space-y-6">
            {filteredQuestions.map((q, idx) => (
              <QuestionCard
                key={q.questionId}
                question={q}
                currentIndex={idx}
                totalQuestions={filteredQuestions.length}
                selectedOptionKey={q.selectedOptionKey}
                isImmediateMode={false}
                isReviewMode={true}
              />
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}

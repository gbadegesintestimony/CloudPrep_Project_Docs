"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  TrendingUp,
  CheckCircle2,
  Bookmark,
  Award,
  Layers,
  AlertTriangle,
  ArrowRight,
  BookOpen,
  Loader2,
  BarChart3,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import {
  fetchProgressSummary,
  fetchCurriculum,
  ProgressSummaryDTO,
  CurriculumResponse,
  getStoredAccessToken,
} from "@/lib/api";

export default function ProgressPage() {
  const router = useRouter();
  const [progress, setProgress] = useState<ProgressSummaryDTO | null>(null);
  const [curriculum, setCurriculum] = useState<CurriculumResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }

    Promise.all([
      fetchProgressSummary().catch(() => null),
      fetchCurriculum().catch(() => null),
    ])
      .then(([p, c]) => {
        if (p) setProgress(p);
        if (c) setCurriculum(c);
      })
      .finally(() => setLoading(false));
  }, [router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />
        <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-sky-600" />
          <p className="text-sm text-slate-600">Loading progress analytics...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-6xl mx-auto">
        {/* Header */}
        <div className="space-y-1 border-b border-sky-100 pb-6">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold uppercase tracking-wider text-sky-600">
              Personalized Learning Analytics
            </span>
          </div>
          <h1 className="font-display text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3">
            <TrendingUp className="h-7 w-7 text-sky-600" />
            Curriculum Progress &amp; Mastery Overview
          </h1>
          <p className="text-sm text-slate-600 leading-relaxed">
            Track your performance metrics across all 10 curriculum topics, identify knowledge gaps, and optimize your exam readiness score.
          </p>
        </div>

        {/* High Level Metrics */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600 flex items-center gap-1.5">
              <CheckCircle2 className="h-4 w-4 text-emerald-600" /> Solved Questions
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight">
              {progress?.totalAnswered || (progress as any)?.totalQuestionsAnswered || 0}
            </div>
            <p className="text-[11px] text-slate-500">{progress?.totalCorrect ?? 0} Correct</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600 flex items-center gap-1.5">
              <TrendingUp className="h-4 w-4 text-sky-600" /> Overall Accuracy
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight">
              {progress && ((progress.totalAnswered ?? 0) > 0 || ((progress as any)?.totalQuestionsAnswered ?? 0) > 0)
                ? `${progress.overallAccuracy ?? (progress as any)?.overallAccuracyPercent ?? 0}%`
                : "—"}
            </div>
            <p className="text-[11px] text-slate-500">Benchmark goal: 80%</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600 flex items-center gap-1.5">
              <Award className="h-4 w-4 text-indigo-600" /> Completed Exams
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight">
              {progress?.totalCompletedSessions ?? 0}
            </div>
            <p className="text-[11px] text-slate-500">Mock presets & practice</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600 flex items-center gap-1.5">
              <Bookmark className="h-4 w-4 text-amber-500" /> Bookmarks
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight">
              {progress?.totalBookmarks ?? 0}
            </div>
            <Link href="/bookmarks" className="text-[11px] font-semibold text-sky-600 hover:underline">
              Review flagged items →
            </Link>
          </div>
        </div>

        {/* Strengths & Weaknesses Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <div className="rounded-2xl border border-emerald-200 bg-emerald-50/80 p-5 space-y-3 shadow-sm">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-emerald-800 flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-emerald-600" /> Demonstrated Strengths
            </h3>
            {progress?.strongestTopics && progress.strongestTopics.length > 0 ? (
              <div className="space-y-2">
                {progress.strongestTopics.map((topic, i) => (
                  <div key={i} className="flex items-center gap-2 text-xs font-medium text-emerald-950">
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                    <span>{topic}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-slate-500">Complete more questions to detect your top domains.</p>
            )}
          </div>

          <div className="rounded-2xl border border-rose-200 bg-rose-50/80 p-5 space-y-3 shadow-sm">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-rose-800 flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-rose-600" /> Priority Focus Domains
            </h3>
            {progress?.weakestTopics && progress.weakestTopics.length > 0 ? (
              <div className="space-y-2">
                {progress.weakestTopics.map((topic, i) => (
                  <div key={i} className="flex items-center justify-between gap-2 text-xs font-medium text-rose-950">
                    <span className="flex items-center gap-2">
                      <span className="h-1.5 w-1.5 rounded-full bg-rose-500" />
                      <span>{topic}</span>
                    </span>
                    <Link
                      href={`/practice?topicId=${topic}`}
                      className="text-[11px] font-bold text-rose-700 underline hover:text-rose-900"
                    >
                      Practice
                    </Link>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-slate-500">No weak topics identified yet. Keep it up!</p>
            )}
          </div>
        </div>

        {/* Topic Mastery Matrix */}
        <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-5 shadow-lg shadow-sky-100/50 backdrop-blur-xl">
          <div className="flex items-center justify-between border-b border-slate-100 pb-4">
            <h2 className="font-display text-base font-bold text-slate-900 flex items-center gap-2">
              <Layers className="h-5 w-5 text-sky-600" /> 10 Topics Mastery Matrix
            </h2>
            <span className="text-xs text-slate-500">Module 1 (Weeks 1–5) &amp; Module 2 (Weeks 6–10)</span>
          </div>

          <div className="space-y-4">
            {(progress?.topicBreakdown || []).map((t) => {
              const accuracy = t.accuracy || 0;
              const isGood = accuracy >= 75;
              return (
                <div key={t.key} className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-slate-800">{t.label || t.key}</span>
                    <span className="font-mono text-slate-600 font-medium">
                      {t.correct}/{t.total || (t as any).attempted || 0} ({accuracy}%)
                    </span>
                  </div>
                  <div className="h-2.5 w-full rounded-full bg-slate-100 overflow-hidden border border-slate-200/50">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        isGood ? "bg-emerald-500" : "bg-sky-500"
                      }`}
                      style={{ width: `${Math.min(100, accuracy)}%` }}
                    />
                  </div>
                </div>
              );
            })}

            {(!progress?.topicBreakdown || progress.topicBreakdown.length === 0) && (
              <p className="text-xs text-slate-500 text-center py-6">
                Start practicing topics to populate your curriculum mastery breakdown.
              </p>
            )}
          </div>
        </div>

        {/* Question Style Breakdown */}
        <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-5 shadow-lg shadow-sky-100/50 backdrop-blur-xl">
          <div className="flex items-center justify-between border-b border-slate-100 pb-4">
            <h2 className="font-display text-base font-bold text-slate-900 flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-indigo-600" /> Question Style Performance
            </h2>
            <span className="text-xs text-slate-500">Recall vs CLI Syntax vs Comparison vs Scenario</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {(progress?.styleBreakdown || []).map((s) => {
              const accuracy = s.accuracy || 0;
              return (
                <div key={s.key} className="rounded-xl border border-slate-200 bg-slate-50/70 p-4 space-y-2">
                  <div className="flex justify-between items-center text-xs">
                    <span className="font-mono font-bold text-slate-800">{s.label || s.key}</span>
                    <span className="font-mono text-sky-600 font-bold">{accuracy}%</span>
                  </div>
                  <div className="h-2 w-full rounded-full bg-slate-200 overflow-hidden">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-sky-500 to-indigo-600 transition-all duration-500"
                      style={{ width: `${Math.min(100, accuracy)}%` }}
                    />
                  </div>
                  <p className="text-[11px] text-slate-500">
                    {s.correct} correct out of {s.total} questions answered
                  </p>
                </div>
              );
            })}
          </div>
        </div>
      </main>
    </div>
  );
}

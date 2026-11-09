"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  BookOpen,
  GraduationCap,
  Bookmark,
  TrendingUp,
  Sparkles,
  ArrowRight,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Layers,
  Terminal,
  Server,
  Cloud,
  ChevronRight,
  ShieldAlert,
  Loader2,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import {
  fetchCurrentUser,
  fetchProgressSummary,
  fetchCurriculum,
  User,
  ProgressSummaryDTO,
  CurriculumResponse,
  getStoredAccessToken,
} from "@/lib/api";

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [progress, setProgress] = useState<ProgressSummaryDTO | null>(null);
  const [curriculum, setCurriculum] = useState<CurriculumResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }

    async function loadData() {
      try {
        const [u, p, c] = await Promise.all([
          fetchCurrentUser().catch(() => null),
          fetchProgressSummary().catch(() => null),
          fetchCurriculum().catch(() => null),
        ]);
        if (u) setUser(u);
        if (p) setProgress(p);
        if (c) setCurriculum(c);
      } finally {
        setLoading(false);
      }
    }

    loadData();
  }, [router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />
        <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3 pt-20">
          <Loader2 className="h-8 w-8 animate-spin text-sky-600" />
          <p className="text-sm text-slate-600">Loading your learner dashboard...</p>
        </div>
      </div>
    );
  }

  const module1Topics = curriculum?.topics.filter((t) => t.moduleNumber === 1) || [];
  const module2Topics = curriculum?.topics.filter((t) => t.moduleNumber === 2) || [];

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-7xl mx-auto">
        {/* Welcome Header (Directly on page, no enclosing card) */}
        <div className="space-y-1">
          <span className="text-xs font-bold uppercase tracking-wider text-sky-600">
            Cloud Engineering Exam Readiness
          </span>
          <h1 className="font-display text-3xl sm:text-4xl font-extrabold tracking-tight text-slate-900 leading-tight">
            Welcome back,<br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-sky-600 via-blue-600 to-indigo-600">
              {user?.name || "Testimony Gbadegesin"}
            </span> 👋
          </h1>

          {/* Guest Account Retention Alert */}
          {user?.isGuest && (
            <div className="mt-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 rounded-xl border border-amber-200 bg-amber-50/80 px-4 py-3 text-xs text-amber-900 shadow-xs">
              <div className="flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-amber-500 shrink-0" />
                <span>
                  Your progress is saved in this browser session. Create a permanent account
                  anytime to keep your exam history.
                </span>
              </div>
              <Link
                href="/register"
                className="font-bold text-amber-800 underline hover:text-amber-950 shrink-0"
              >
                Save My Progress →
              </Link>
            </div>
          )}
        </div>

        {/* Quick Stats Grid */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 shadow-sm shadow-sky-50 space-y-1">
            <span className="text-xs font-semibold text-slate-500 flex items-center gap-1.5">
              <CheckCircle2 className="h-4 w-4 text-emerald-500" /> Questions Solved
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight font-display">
              {progress?.totalAnswered || (progress as any)?.totalQuestionsAnswered || 0}
            </div>
            <p className="text-[11px] text-slate-500">
              {progress?.totalCorrect ?? 0} correct answers
            </p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 shadow-sm shadow-sky-50 space-y-1">
            <span className="text-xs font-semibold text-slate-500 flex items-center gap-1.5">
              <TrendingUp className="h-4 w-4 text-sky-500" /> Overall Accuracy
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight font-display">
              {progress && ((progress.totalAnswered ?? 0) > 0 || ((progress as any)?.totalQuestionsAnswered ?? 0) > 0)
                ? `${progress.overallAccuracy ?? (progress as any)?.overallAccuracyPercent ?? 0}%`
                : "—"}
            </div>
            <p className="text-[11px] text-slate-500">Target benchmark: 80%</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 shadow-sm shadow-sky-50 space-y-1">
            <span className="text-xs font-semibold text-slate-500 flex items-center gap-1.5">
              <Bookmark className="h-4 w-4 text-amber-500" /> Saved Bookmarks
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight font-display">
              {progress?.totalBookmarks ?? 0}
            </div>
            <Link
              href="/bookmarks"
              className="text-[11px] text-sky-600 hover:text-sky-700 font-semibold transition-colors inline-block"
            >
              Review flagged items →
            </Link>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 shadow-sm shadow-sky-50 space-y-1">
            <span className="text-xs font-semibold text-slate-500 flex items-center gap-1.5">
              <Layers className="h-4 w-4 text-indigo-500" /> Question Bank Size
            </span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight font-display">
              {curriculum?.totalBankQuestions ?? 1040}+
            </div>
            <p className="text-[11px] text-slate-500">10 Topics & Dual AI auto-growth</p>
          </div>
        </div>

        {/* Onboarding Welcome Banner for brand new learners */}
        {progress && (progress.totalAnswered ?? 0) === 0 && (
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-xl border border-sky-200 bg-sky-50/80 p-4 text-sm text-sky-900 shadow-xs">
            <div className="space-y-1">
              <span className="font-bold text-sky-950 block flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-sky-600" /> Welcome to CloudPrep!
              </span>
              <p className="text-xs text-sky-700 leading-relaxed">
                Take your first practice session to test your cloud engineering knowledge and unlock personalized topic recommendations.
              </p>
            </div>
            <Link
              href="/practice"
              className="rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-xs font-bold text-white hover:from-sky-600 hover:to-blue-700 shrink-0 transition-all shadow-sm"
            >
              Start First Practice →
            </Link>
          </div>
        )}

        {/* Weak Topic Alert Recommendation (Only for active learners with attempted questions) */}
        {progress && (progress.totalAnswered ?? 0) > 0 && progress.weakestTopics && progress.weakestTopics.length > 0 && (
          <div className="flex items-start gap-4 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800 shadow-sm">
            <AlertTriangle className="h-5 w-5 text-rose-500 shrink-0 mt-0.5" />
            <div className="space-y-1 flex-1">
              <span className="font-semibold text-rose-900 block">
                Targeted Recommendation: Strengthen Weak Topics
              </span>
              <p className="text-xs text-rose-700 leading-relaxed">
                Your performance shows lower accuracy in:{" "}
                <strong className="text-rose-950 font-bold">
                  {progress.weakestTopics.join(", ")}
                </strong>
                . Practice these topics to push your score past the 80% pass threshold.
              </p>
            </div>
            <Link
              href="/practice"
              className="rounded-xl bg-rose-100 border border-rose-300 px-3.5 py-1.5 text-xs font-bold text-rose-800 hover:bg-rose-200 shrink-0 transition-all shadow-sm"
            >
              Fix Mistakes
            </Link>
          </div>
        )}

        {/* Timed Exam Presets Grid */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h2 className="text-lg font-bold text-slate-900 flex items-center gap-2 font-display">
                <GraduationCap className="h-5 w-5 text-sky-600" />
                Authoritative Timed Mock Exams
              </h2>
              <p className="text-xs text-slate-500">
                Simulate real certification conditions with server-enforced timers and deferred answer review.
              </p>
            </div>
            <Link
              href="/exams"
              className="text-xs font-semibold text-sky-600 hover:text-sky-700 transition-colors"
            >
              View all exams →
            </Link>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            {/* Module 1 Exam */}
            <div className="group rounded-2xl border border-sky-100 bg-white/95 p-6 flex flex-col justify-between hover:border-sky-300 hover:shadow-lg hover:shadow-sky-100/60 transition-all shadow-sm">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="rounded-full bg-sky-50 px-2.5 py-0.5 text-[11px] font-semibold text-sky-700 border border-sky-200">
                    MODULE 1
                  </span>
                  <div className="flex items-center gap-1 text-xs text-slate-500 font-medium">
                    <Clock className="h-3.5 w-3.5" /> 40 Mins
                  </div>
                </div>
                <h3 className="text-base font-bold text-slate-900 group-hover:text-sky-600 transition-colors font-display">
                  Module 1 Midterm Exam
                </h3>
                <p className="text-xs text-slate-600 leading-relaxed font-normal">
                  Covers Linux internals, Bash scripting, Git workflows, Virtualization, and Network Protocols (Weeks 1–5).
                </p>
              </div>

              <div className="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between text-xs">
                <span className="text-slate-500 font-medium">30 Questions</span>
                <Link
                  href="/exams?preset=MODULE_1"
                  className="inline-flex items-center gap-1 font-bold text-sky-600 hover:text-sky-700 group-hover:translate-x-0.5 transition-transform"
                >
                  Start Exam <ChevronRight className="h-3.5 w-3.5" />
                </Link>
              </div>
            </div>

            {/* Module 2 Exam */}
            <div className="group rounded-2xl border border-indigo-100 bg-white/95 p-6 flex flex-col justify-between hover:border-indigo-300 hover:shadow-lg hover:shadow-indigo-100/60 transition-all shadow-sm">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="rounded-full bg-indigo-50 px-2.5 py-0.5 text-[11px] font-semibold text-indigo-700 border border-indigo-200">
                    MODULE 2
                  </span>
                  <div className="flex items-center gap-1 text-xs text-slate-500 font-medium">
                    <Clock className="h-3.5 w-3.5" /> 40 Mins
                  </div>
                </div>
                <h3 className="text-base font-bold text-slate-900 group-hover:text-indigo-600 transition-colors font-display">
                  Module 2 Cloud & Security Exam
                </h3>
                <p className="text-xs text-slate-600 leading-relaxed font-normal">
                  Covers Cloud Architecture, AWS IAM & Security, VPC Networking, Compute/Lambda, and Storage/Databases (Weeks 6–10).
                </p>
              </div>

              <div className="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between text-xs">
                <span className="text-slate-500 font-medium">30 Questions</span>
                <Link
                  href="/exams?preset=MODULE_2"
                  className="inline-flex items-center gap-1 font-bold text-indigo-600 hover:text-indigo-700 group-hover:translate-x-0.5 transition-transform"
                >
                  Start Exam <ChevronRight className="h-3.5 w-3.5" />
                </Link>
              </div>
            </div>

            {/* Comprehensive Final */}
            <div className="group rounded-2xl border border-purple-200 bg-gradient-to-b from-white to-purple-50/40 p-6 flex flex-col justify-between hover:border-purple-300 hover:shadow-lg hover:shadow-purple-100/60 transition-all shadow-sm">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="rounded-full bg-purple-100 px-2.5 py-0.5 text-[11px] font-semibold text-purple-800 border border-purple-200">
                    FULL CURRICULUM
                  </span>
                  <div className="flex items-center gap-1 text-xs text-purple-700 font-semibold">
                    <Clock className="h-3.5 w-3.5" /> 80 Mins
                  </div>
                </div>
                <h3 className="text-base font-bold text-slate-900 group-hover:text-purple-700 transition-colors font-display">
                  Comprehensive Final Mock Exam
                </h3>
                <p className="text-xs text-slate-600 leading-relaxed font-normal">
                  The ultimate benchmark. 60 balanced questions across all 10 topics with strict exam conditions.
                </p>
              </div>

              <div className="mt-6 pt-4 border-t border-purple-100 flex items-center justify-between text-xs">
                <span className="text-slate-500 font-medium">60 Questions</span>
                <Link
                  href="/exams?preset=COMPREHENSIVE_FINAL"
                  className="inline-flex items-center gap-1 font-bold text-purple-700 hover:text-purple-800 group-hover:translate-x-0.5 transition-transform"
                >
                  Launch Final <ChevronRight className="h-3.5 w-3.5" />
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Curriculum Topics Mastery Matrix */}
        <div className="space-y-4">
          <div className="space-y-1">
            <h2 className="text-lg font-bold text-slate-900 flex items-center gap-2 font-display">
              <Layers className="h-5 w-5 text-sky-600" />
              Curriculum Mastery Matrix (10 Core Topics)
            </h2>
            <p className="text-xs text-slate-500">
              Direct topic practice with instant feedback, full explanations, and distractor rationales.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Module 1 List */}
            <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-4 shadow-sm shadow-sky-50">
              <div className="flex items-center justify-between pb-3 border-b border-slate-100">
                <div className="flex items-center gap-2">
                  <Terminal className="h-4 w-4 text-sky-600" />
                  <span className="font-bold text-sm text-slate-900 font-display">Module 1: Infrastructure Foundations</span>
                </div>
                <span className="text-xs text-slate-500 font-mono">Weeks 1–5</span>
              </div>

              <div className="divide-y divide-slate-100">
                {module1Topics.map((t) => (
                  <div key={t.id} className="py-3 flex items-center justify-between gap-4 group">
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-mono font-semibold text-sky-600 bg-sky-50 px-2 py-0.5 rounded border border-sky-200/80">{t.weekCode}</span>
                        <span className="text-sm font-semibold text-slate-800 group-hover:text-sky-600 transition-colors">
                          {t.title}
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-500 font-normal">
                        {t.questionCount || 104} approved questions • 4 styles
                      </p>
                    </div>

                    <Link
                      href={`/practice?topicId=${t.id}`}
                      className="rounded-xl border border-sky-200 bg-sky-50 px-3.5 py-1.5 text-xs font-bold text-sky-700 hover:bg-sky-100 transition-all shrink-0 shadow-sm"
                    >
                      Practice
                    </Link>
                  </div>
                ))}
              </div>
            </div>

            {/* Module 2 List */}
            <div className="rounded-2xl border border-indigo-100 bg-white/95 p-6 space-y-4 shadow-sm shadow-indigo-50">
              <div className="flex items-center justify-between pb-3 border-b border-slate-100">
                <div className="flex items-center gap-2">
                  <Cloud className="h-4 w-4 text-indigo-600" />
                  <span className="font-bold text-sm text-slate-900 font-display">Module 2: AWS Cloud & Security Architecture</span>
                </div>
                <span className="text-xs text-slate-500 font-mono">Weeks 6–10</span>
              </div>

              <div className="divide-y divide-slate-100">
                {module2Topics.map((t) => (
                  <div key={t.id} className="py-3 flex items-center justify-between gap-4 group">
                    <div className="space-y-0.5">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-mono font-semibold text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded border border-indigo-200/80">{t.weekCode}</span>
                        <span className="text-sm font-semibold text-slate-800 group-hover:text-indigo-600 transition-colors">
                          {t.title}
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-500 font-normal">
                        {t.questionCount || 104} approved questions • 4 styles
                      </p>
                    </div>

                    <Link
                      href={`/practice?topicId=${t.id}`}
                      className="rounded-xl border border-indigo-200 bg-indigo-50 px-3.5 py-1.5 text-xs font-bold text-indigo-700 hover:bg-indigo-100 transition-all shrink-0 shadow-sm"
                    >
                      Practice
                    </Link>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}


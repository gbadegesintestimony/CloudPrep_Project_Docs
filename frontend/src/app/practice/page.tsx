"use client";

import React, { useEffect, useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import {
  BookOpen,
  ArrowRight,
  ArrowLeft,
  RotateCcw,
  Sparkles,
  CheckCircle2,
  XCircle,
  Bookmark,
  Award,
  ChevronRight,
  Loader2,
  SlidersHorizontal,
  Play,
  X,
  AlertTriangle,
  LogOut,
  Clock,
  Trash2,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import { QuestionCard } from "@/components/QuestionCard";
import {
  fetchCurriculum,
  createPracticeSession,
  getPracticeSession,
  fetchActivePracticeSessions,
  cancelPracticeSession,
  cancelAllActiveSessions,
  submitQuestionAnswer,
  submitExam,
  CurriculumResponse,
  PracticeSessionDTO,
  ActiveSessionQuestionDTO,
  getStoredAccessToken,
} from "@/lib/api";

export default function PracticePage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
          <LearnerNav />
          <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
            <Loader2 className="h-8 w-8 animate-spin text-sky-600" />
            <p className="text-sm text-slate-600 font-medium">Loading practice session...</p>
          </div>
        </div>
      }
    >
      <PracticeContent />
    </Suspense>
  );
}

function PracticeContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const preselectedTopic = searchParams.get("topicId") || "";

  const [curriculum, setCurriculum] = useState<CurriculumResponse | null>(null);
  const [session, setSession] = useState<PracticeSessionDTO | null>(null);
  const [questions, setQuestions] = useState<ActiveSessionQuestionDTO[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);

  // Configuration State
  const [selectedTopicId, setSelectedTopicId] = useState(preselectedTopic);
  const [selectedStyle, setSelectedStyle] = useState("ALL");
  const [selectedDifficulty, setSelectedDifficulty] = useState("ALL");
  const [questionCount, setQuestionCount] = useState(10);
  const [starting, setStarting] = useState(false);
  const [answering, setAnswering] = useState(false);
  const [sessionCompleted, setSessionCompleted] = useState(false);

  // Active Sessions & Continuity Management
  const [activeSessions, setActiveSessions] = useState<PracticeSessionDTO[]>([]);
  const [loadingActiveSessions, setLoadingActiveSessions] = useState(false);
  const [conflictModalOpen, setConflictModalOpen] = useState(false);
  const [exitModalOpen, setExitModalOpen] = useState(false);
  const [cancellingSessionId, setCancellingSessionId] = useState<string | null>(null);

  const loadActiveSessions = async () => {
    try {
      setLoadingActiveSessions(true);
      const list = await fetchActivePracticeSessions();
      setActiveSessions(list);
    } catch {
      setActiveSessions([]);
    } finally {
      setLoadingActiveSessions(false);
    }
  };

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }
    fetchCurriculum().then(setCurriculum).catch(() => null);
    loadActiveSessions();
  }, [router]);

  useEffect(() => {
    if (preselectedTopic) {
      setSelectedTopicId(preselectedTopic);
    }
  }, [preselectedTopic]);

  const handleResumeSession = async (sessionId: string) => {
    try {
      setStarting(true);
      const res = await getPracticeSession(sessionId);
      setSession(res.session);
      setQuestions(res.questions);
      // Automatically jump to the first question that hasn't been answered yet
      const firstUnanswered = res.questions.findIndex((q) => !q.selectedOptionKey);
      setCurrentIndex(firstUnanswered >= 0 ? firstUnanswered : 0);
      setSessionCompleted(res.questions.length > 0 && res.questions.every((q) => !!q.selectedOptionKey));
      setConflictModalOpen(false);
    } catch (err: any) {
      alert(err?.message || "Failed to resume session.");
    } finally {
      setStarting(false);
    }
  };

  const handleCancelSession = async (sessionId: string) => {
    try {
      setCancellingSessionId(sessionId);
      await cancelPracticeSession(sessionId);
      await loadActiveSessions();
      if (session?.id === sessionId) {
        setSession(null);
        setQuestions([]);
      }
    } catch (err: any) {
      alert(err?.message || "Failed to cancel session.");
    } finally {
      setCancellingSessionId(null);
    }
  };

  const handleCancelAllActive = async () => {
    if (!confirm("Cancel all unfinished active sessions so you can start clean?")) return;
    try {
      setStarting(true);
      await cancelAllActiveSessions();
      await loadActiveSessions();
      setConflictModalOpen(false);
    } catch (err: any) {
      alert(err?.message || "Failed to cancel active sessions.");
    } finally {
      setStarting(false);
    }
  };

  const handleStartPractice = async (cancelPrevious = false) => {
    try {
      setStarting(true);
      const res = await createPracticeSession({
        mode: selectedTopicId ? "TOPIC_PRACTICE" : "MIXED_PRACTICE",
        topicId: selectedTopicId || undefined,
        questionStyle: selectedStyle !== "ALL" ? selectedStyle : undefined,
        difficulty: selectedDifficulty !== "ALL" ? selectedDifficulty : undefined,
        feedbackMode: "IMMEDIATE",
        questionCount,
        cancelPrevious,
      });
      setSession(res.session);
      setQuestions(res.questions);
      setCurrentIndex(0);
      setSessionCompleted(false);
      setConflictModalOpen(false);
      await loadActiveSessions();
    } catch (err: any) {
      if (
        err?.error === "MAX_ACTIVE_SESSIONS_EXCEEDED" ||
        (err?.message && (err.message.includes("active sessions") || err.message.includes("finish an existing session")))
      ) {
        setConflictModalOpen(true);
        loadActiveSessions();
      } else {
        alert(err?.message || "Failed to start practice session");
      }
    } finally {
      setStarting(false);
    }
  };

  const [submittingSession, setSubmittingSession] = useState(false);

  const handleSelectOption = async (optionKey: string) => {
    if (!session || answering) return;
    const currentQ = questions[currentIndex];
    if (!currentQ || currentQ.selectedOptionKey) return;

    // 1. INSTANT LOCAL EVALUATION (0ms feedback! No lag waiting for remote DB)
    const isCorrect = currentQ.correctOptionKey ? optionKey === currentQ.correctOptionKey : undefined;

    const instant = [...questions];
    instant[currentIndex] = {
      ...currentQ,
      selectedOptionKey: optionKey,
      isCorrect: isCorrect,
    };
    setQuestions(instant);

    // 2. BACKGROUND PERSISTENCE: Save answer to DB asynchronously
    try {
      setAnswering(true);
      const res = await submitQuestionAnswer(session.id, currentQ.questionId, optionKey);

      // Merge verified backend result if needed
      const updated = [...instant];
      updated[currentIndex] = {
        ...updated[currentIndex],
        isCorrect: res.isCorrect !== undefined ? res.isCorrect : isCorrect,
        correctOptionKey: res.correctOptionKey || currentQ.correctOptionKey,
        explanation: res.explanation || currentQ.explanation,
      };
      setQuestions(updated);

      // Check if all questions are answered
      const allAnswered = updated.every((q) => !!q.selectedOptionKey);
      if (allAnswered && currentIndex === questions.length - 1) {
        setSessionCompleted(true);
      }
    } catch (err: any) {
      console.warn("Background answer save:", err);
    } finally {
      setAnswering(false);
    }
  };

  const handleCompleteSession = async () => {
    if (!session || submittingSession) return;
    try {
      setSubmittingSession(true);
      await submitExam(session.id);
      router.push(`/results/${session.id}`);
    } catch (err: any) {
      console.warn("Failed to submit exam session:", err);
      // Still show completed state or route
      setSessionCompleted(true);
    } finally {
      setSubmittingSession(false);
    }
  };

  const handleToggleFlag = () => {
    const updated = [...questions];
    updated[currentIndex] = {
      ...updated[currentIndex],
      isFlagged: !updated[currentIndex].isFlagged,
    };
    setQuestions(updated);
  };

  // If in active practice session
  if (session && questions.length > 0) {
    const currentQ = questions[currentIndex];
    const answeredCount = questions.filter((q) => !!q.selectedOptionKey).length;
    const correctCount = questions.filter((q) => q.isCorrect === true).length;
    const progressPercent = Math.round(((currentIndex + 1) / questions.length) * 100);

    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />

        <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-6 max-w-5xl mx-auto">
          {/* Practice Session Top Bar */}
          <div className="flex items-center justify-between border-b border-sky-100 pb-4">
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold uppercase tracking-wider text-sky-600">
                  Immediate Feedback Practice
                </span>
                <span className="rounded-full bg-sky-50 px-2.5 py-0.5 text-[10px] font-mono font-bold text-sky-700 border border-sky-200">
                  {session.title}
                </span>
              </div>
              <p className="text-xs text-slate-500 font-medium">
                {answeredCount} of {questions.length} answered • {correctCount} correct
              </p>
            </div>

            <button
              onClick={() => setExitModalOpen(true)}
              className="rounded-xl border border-slate-200 bg-white px-3.5 py-1.5 text-xs font-semibold text-slate-600 hover:text-slate-900 hover:bg-slate-50 transition-all shadow-sm flex items-center gap-1.5 cursor-pointer"
            >
              <LogOut className="h-3.5 w-3.5" /> Exit / Cancel
            </button>
          </div>

          {/* Progress Bar */}
          <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100">
            <div
              className="h-full bg-gradient-to-r from-sky-500 to-blue-600 transition-all duration-300"
              style={{ width: `${progressPercent}%` }}
            />
          </div>

          {/* Active Question Card */}
          <QuestionCard
            question={currentQ}
            currentIndex={currentIndex}
            totalQuestions={questions.length}
            selectedOptionKey={currentQ.selectedOptionKey}
            isImmediateMode={true}
            onSelectOption={handleSelectOption}
            onToggleFlag={handleToggleFlag}
          />

          {/* Instant Answer Feedback Banner */}
          {currentQ.selectedOptionKey && currentQ.isCorrect !== undefined && (
            <div
              className={`rounded-xl border px-4 py-3 flex items-center justify-between text-xs font-semibold shadow-xs transition-all animate-in fade-in slide-in-from-bottom-2 ${
                currentQ.isCorrect
                  ? "border-emerald-300 bg-emerald-50 text-emerald-950"
                  : "border-rose-300 bg-rose-50 text-rose-950"
              }`}
            >
              <div className="flex items-center gap-2.5">
                {currentQ.isCorrect ? (
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 shrink-0" />
                ) : (
                  <XCircle className="h-5 w-5 text-rose-600 shrink-0" />
                )}
                <span>
                  {currentQ.isCorrect
                    ? "✓ Correct! Review the explanation below before moving to the next question."
                    : "✗ Incorrect choice. Check the explanation below to master this concept."}
                </span>
              </div>
              <span className="text-[11px] font-mono font-bold text-slate-500 hidden sm:inline">
                Click Next to continue →
              </span>
            </div>
          )}

          {/* Bottom Navigation Buttons */}
          <div className="flex items-center justify-between gap-2 pt-2">
            <button
              type="button"
              onClick={() => setCurrentIndex((prev) => Math.max(0, prev - 1))}
              disabled={currentIndex === 0}
              className="flex items-center gap-1 sm:gap-1.5 rounded-xl border border-slate-200 bg-white px-3 sm:px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-slate-900 disabled:opacity-40 transition-all shadow-sm shrink-0"
            >
              <ArrowLeft className="h-4 w-4" /> <span>Previous</span>
            </button>

            {currentIndex < questions.length - 1 ? (
              <button
                type="button"
                onClick={() => setCurrentIndex((prev) => Math.min(questions.length - 1, prev + 1))}
                className="flex items-center gap-1 sm:gap-1.5 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-4 sm:px-5 py-2 text-xs font-semibold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/25 transition-all cursor-pointer shrink-0"
              >
                <span>Next</span> <span className="hidden xs:inline">Question</span> <ArrowRight className="h-4 w-4" />
              </button>
            ) : (
              <button
                type="button"
                onClick={handleCompleteSession}
                disabled={submittingSession}
                className="flex items-center gap-1 sm:gap-1.5 rounded-xl bg-emerald-600 px-3.5 sm:px-5 py-2 text-xs font-semibold text-white hover:bg-emerald-500 shadow-md shadow-emerald-600/25 transition-all cursor-pointer disabled:opacity-60 shrink-0"
              >
                {submittingSession ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" /> Completing...
                  </>
                ) : (
                  <>
                    <span>Complete</span> <span className="hidden xs:inline">Session</span> <CheckCircle2 className="h-4 w-4" />
                  </>
                )}
              </button>
            )}
          </div>

          {/* Session Completed Overlay / Summary */}
          {sessionCompleted && (
            <div className="mt-8 rounded-2xl border border-emerald-200 bg-emerald-50/80 p-6 sm:p-8 text-center space-y-4 shadow-lg shadow-emerald-50">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 border border-emerald-200">
                <Award className="h-6 w-6" />
              </div>
              <h2 className="text-xl font-bold text-slate-900 font-display">Practice Session Finished!</h2>
              <p className="text-sm text-slate-600">
                You answered <strong className="text-slate-900">{correctCount}</strong> out of{" "}
                <strong className="text-slate-900">{questions.length}</strong> questions correctly (
                <strong className="text-emerald-600 font-mono font-bold">
                  {Math.round((correctCount / questions.length) * 100)}%
                </strong>
                ).
              </p>
              <div className="flex flex-wrap items-center justify-center gap-3 pt-2">
                <Link
                  href={`/results/${session.id}`}
                  className="rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 px-4 py-2 text-xs font-bold text-white hover:from-emerald-700 hover:to-teal-700 shadow-md shadow-emerald-600/20"
                >
                  Review Full Results →
                </Link>
                <button
                  onClick={() => handleStartPractice()}
                  className="rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-xs font-semibold text-white hover:from-sky-600 hover:to-blue-700 shadow-md shadow-sky-500/25 cursor-pointer"
                >
                  Practice Again
                </button>
                <Link
                  href="/dashboard"
                  className="rounded-xl border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-sm"
                >
                  Back to Dashboard
                </Link>
              </div>
            </div>
          )}
          {/* Exit / Cancel Modal */}
          {exitModalOpen && (
            <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4 animate-in fade-in">
              <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl space-y-4 border border-slate-100">
                <div className="flex items-center gap-3 text-slate-900">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-50 text-amber-600 border border-amber-200">
                    <AlertTriangle className="h-5 w-5" />
                  </div>
                  <div>
                    <h3 className="text-base font-bold">Leave Practice Session?</h3>
                    <p className="text-xs text-slate-500">Choose whether to keep your progress or discard this session.</p>
                  </div>
                </div>

                <p className="text-xs text-slate-600 leading-relaxed">
                  You have answered <strong>{answeredCount}</strong> of <strong>{questions.length}</strong> questions.
                </p>

                <div className="space-y-2 pt-2">
                  <button
                    type="button"
                    onClick={() => {
                      setSession(null);
                      setExitModalOpen(false);
                      loadActiveSessions();
                    }}
                    className="w-full flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2.5 text-xs font-bold text-white hover:from-sky-600 hover:to-blue-700 shadow-md shadow-sky-500/20 cursor-pointer"
                  >
                    <Clock className="h-4 w-4" /> Save & Continue Later
                  </button>
                  <button
                    type="button"
                    onClick={async () => {
                      if (session) {
                        await handleCancelSession(session.id);
                      }
                      setExitModalOpen(false);
                    }}
                    className="w-full flex items-center justify-center gap-2 rounded-xl border border-rose-200 bg-rose-50/50 px-4 py-2.5 text-xs font-bold text-rose-700 hover:bg-rose-100 cursor-pointer"
                  >
                    <Trash2 className="h-4 w-4" /> Cancel & Discard Session
                  </button>
                  <button
                    type="button"
                    onClick={() => setExitModalOpen(false)}
                    className="w-full rounded-xl border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-50 cursor-pointer"
                  >
                    Resume Answering
                  </button>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    );
  }

  // Configuration Screen (Topic, Style, Difficulty, Count selection)
  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-5xl mx-auto">
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold uppercase tracking-wider text-sky-600">
              Interactive Practice Runner
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3 font-display">
            <BookOpen className="h-7 w-7 text-sky-600" />
            Configure Your Practice Session
          </h1>
          <p className="text-sm text-slate-600 leading-relaxed font-normal">
            Target a specific topic or practice mixed questions across the entire curriculum.
            Enjoy instant feedback, distractor rationales, and full course notes citations on every question.
          </p>
        </div>

        {/* Active Session Recovery Banner */}
        {activeSessions.length > 0 && (
          <div className="rounded-2xl border border-sky-200 bg-gradient-to-r from-sky-50 via-blue-50/50 to-indigo-50/40 p-5 sm:p-6 shadow-sm space-y-4 animate-in fade-in">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="flex h-2.5 w-2.5 rounded-full bg-sky-500 animate-pulse" />
                  <span className="text-xs font-bold uppercase tracking-wider text-sky-700">
                    Active Session in Progress
                  </span>
                  {activeSessions.length > 1 && (
                    <span className="text-[11px] font-semibold px-2 py-0.5 rounded-full bg-sky-100 text-sky-800 border border-sky-200">
                      {activeSessions.length} active sessions
                    </span>
                  )}
                </div>
                <h3 className="text-base font-bold text-slate-900">
                  {activeSessions[0].title || "Cloud Engineering Practice Session"}
                </h3>
                <p className="text-xs text-slate-600">
                  Started {new Date(activeSessions[0].startedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} • Mode: {activeSessions[0].mode}
                </p>
              </div>

              <div className="flex flex-wrap items-center gap-2.5">
                <button
                  type="button"
                  onClick={() => handleResumeSession(activeSessions[0].id)}
                  disabled={starting}
                  className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-xs font-bold text-white hover:from-sky-600 hover:to-blue-700 shadow-md shadow-sky-500/20 transition-all cursor-pointer"
                >
                  {starting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Play className="h-3.5 w-3.5 fill-current" />}
                  <span>Continue Answering</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleCancelSession(activeSessions[0].id)}
                  disabled={cancellingSessionId === activeSessions[0].id}
                  className="flex items-center gap-1.5 rounded-xl border border-rose-200 bg-white px-3.5 py-2 text-xs font-semibold text-rose-600 hover:bg-rose-50 shadow-xs transition-all cursor-pointer"
                >
                  {cancellingSessionId === activeSessions[0].id ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <X className="h-3.5 w-3.5" />
                  )}
                  <span>Cancel / Discard</span>
                </button>
              </div>
            </div>

            {activeSessions.length > 1 && (
              <div className="pt-3 border-t border-sky-100 flex items-center justify-between text-xs text-slate-500">
                <span>You have multiple unfinished sessions ({activeSessions.length} total).</span>
                <button
                  type="button"
                  onClick={handleCancelAllActive}
                  className="text-xs font-semibold text-rose-600 hover:text-rose-700 hover:underline cursor-pointer"
                >
                  Cancel all {activeSessions.length} active sessions
                </button>
              </div>
            )}
          </div>
        )}

        <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 sm:p-8 space-y-6 shadow-xl shadow-sky-100/60">
          {/* Topic Selector */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700">
              Select Curriculum Topic
            </label>
            <select
              value={selectedTopicId}
              onChange={(e) => setSelectedTopicId(e.target.value)}
              className="block w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-900 focus:border-sky-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-sky-500/20"
            >
              <option value="">✨ Mixed Practice (All 10 Topics across Module 1 & 2)</option>
              {curriculum?.topics.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.weekCode} — {t.title} ({t.questionCount || 104} Qs)
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-5">
            {/* Question Style Filter */}
            <div className="space-y-2">
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700">
                Question Style Filter
              </label>
              <select
                value={selectedStyle}
                onChange={(e) => setSelectedStyle(e.target.value)}
                className="block w-full rounded-xl border border-slate-200 bg-slate-50 px-3.5 py-2.5 text-xs text-slate-900 focus:border-sky-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-sky-500/20"
              >
                <option value="ALL">All 4 Styles (Balanced)</option>
                <option value="DEFINITION">Definitions & Core Recall (~30%)</option>
                <option value="COMMAND_SYNTAX">CLI, Bash & Git Syntax (~25%)</option>
                <option value="COMPARISON">Concept Comparisons (~20%)</option>
                <option value="SCENARIO">Cloud Workplace Scenarios (~25%)</option>
              </select>
            </div>

            {/* Difficulty Filter */}
            <div className="space-y-2">
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700">
                Difficulty Level
              </label>
              <select
                value={selectedDifficulty}
                onChange={(e) => setSelectedDifficulty(e.target.value)}
                className="block w-full rounded-xl border border-slate-200 bg-slate-50 px-3.5 py-2.5 text-xs text-slate-900 focus:border-sky-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-sky-500/20"
              >
                <option value="ALL">All Difficulties</option>
                <option value="FOUNDATIONAL">Foundational (Recall & Syntax)</option>
                <option value="INTERMEDIATE">Intermediate (Application)</option>
                <option value="ADVANCED">Advanced (Scenario Debugging)</option>
              </select>
            </div>

            {/* Question Count */}
            <div className="space-y-2">
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700">
                Question Count
              </label>
              <div className="grid grid-cols-4 gap-2">
                {[5, 10, 20, 30].map((count) => (
                  <button
                    key={count}
                    type="button"
                    onClick={() => setQuestionCount(count)}
                    className={`rounded-xl border py-2 text-xs font-bold transition-all ${
                      questionCount === count
                        ? "border-sky-500 bg-sky-50 text-sky-700 shadow-sm"
                        : "border-slate-200 bg-slate-50 text-slate-600 hover:text-slate-900 hover:bg-slate-100"
                    }`}
                  >
                    {count}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="pt-4 border-t border-slate-100 flex items-center justify-between">
            <span className="text-xs text-slate-500">
              Immediate feedback mode: answers revealed instantly on click.
            </span>
            <button
              type="button"
              onClick={() => handleStartPractice()}
              disabled={starting}
              className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-6 py-2.5 text-sm font-semibold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/25 transition-all disabled:opacity-50"
            >
              {starting ? (
                <Loader2 className="h-4 w-4 animate-spin text-white" />
              ) : (
                <ArrowRight className="h-4 w-4" />
              )}
              <span>Launch Practice Quiz</span>
            </button>
          </div>
        </div>
        {/* Active Session Limit / Conflict Modal */}
        {conflictModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4 animate-in fade-in">
            <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl space-y-4 border border-slate-100">
              <div className="flex items-center gap-3 text-slate-900">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-sky-50 text-sky-600 border border-sky-200">
                  <BookOpen className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-base font-bold">Active Sessions in Progress</h3>
                  <p className="text-xs text-slate-500">You have active practice sessions open.</p>
                </div>
              </div>

              <p className="text-xs text-slate-600 leading-relaxed">
                You currently have {activeSessions.length || "existing"} active practice session{activeSessions.length === 1 ? "" : "s"}. Would you like to continue answering questions from your latest session, or cancel previous sessions to start fresh?
              </p>

              <div className="space-y-2 pt-2">
                {activeSessions.length > 0 && (
                  <button
                    type="button"
                    onClick={() => handleResumeSession(activeSessions[0].id)}
                    className="w-full flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2.5 text-xs font-bold text-white hover:from-sky-600 hover:to-blue-700 shadow-md shadow-sky-500/20 cursor-pointer"
                  >
                    <Play className="h-4 w-4 fill-current" /> Continue Answering Latest Session
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => handleStartPractice(true)}
                  className="w-full flex items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-xs font-bold text-slate-700 hover:bg-slate-50 cursor-pointer"
                >
                  <RotateCcw className="h-4 w-4" /> Cancel Old Sessions & Start Fresh
                </button>
                <button
                  type="button"
                  onClick={() => setConflictModalOpen(false)}
                  className="w-full rounded-xl py-1.5 text-xs font-semibold text-slate-500 hover:text-slate-700 cursor-pointer"
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

"use client";

import React, { useEffect, useState, useRef, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import {
  GraduationCap,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Flag,
  ArrowRight,
  ArrowLeft,
  ChevronRight,
  Loader2,
  ShieldAlert,
  Send,
  HelpCircle,
  Play,
  X,
  Trash2,
  RotateCcw,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import { QuestionCard } from "@/components/QuestionCard";
import {
  createPracticeSession,
  getPracticeSession,
  fetchActivePracticeSessions,
  cancelPracticeSession,
  submitQuestionAnswer,
  submitExam,
  PracticeSessionDTO,
  ActiveSessionQuestionDTO,
  getStoredAccessToken,
} from "@/lib/api";

export default function ExamsPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
          <LearnerNav />
          <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
            <Loader2 className="h-8 w-8 animate-spin text-sky-600" />
            <p className="text-sm text-slate-500 font-medium">Loading mock exam simulation...</p>
          </div>
        </div>
      }
    >
      <ExamsContent />
    </Suspense>
  );
}

function ExamsContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const presetQuery = searchParams.get("preset");

  const [session, setSession] = useState<PracticeSessionDTO | null>(null);
  const [questions, setQuestions] = useState<ActiveSessionQuestionDTO[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [starting, setStarting] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [showSubmitModal, setShowSubmitModal] = useState(false);
  const [activeMocks, setActiveMocks] = useState<PracticeSessionDTO[]>([]);
  const [conflictModalOpen, setConflictModalOpen] = useState(false);
  const [cancellingSessionId, setCancellingSessionId] = useState<string | null>(null);

  // Authoritative countdown timer state
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [timeNotification, setTimeNotification] = useState<{
    type: "10min" | "5min";
    title: string;
    message: string;
  } | null>(null);
  const warned10mRef = useRef(false);
  const warned5mRef = useRef(false);

  const loadActiveMocks = async () => {
    try {
      const list = await fetchActivePracticeSessions();
      setActiveMocks(list.filter((s) => s.mode === "MOCK_EXAM"));
    } catch {
      setActiveMocks([]);
    }
  };

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }
    loadActiveMocks();

    // Auto-launch preset if preset query parameter is provided
    if (presetQuery && !session) {
      handleLaunchExam(presetQuery);
    }
  }, [presetQuery, router]);

  // Countdown Timer effect with 10m & 5m notifications
  useEffect(() => {
    if (!session || !session.deadline) return;

    warned10mRef.current = false;
    warned5mRef.current = false;

    const interval = setInterval(() => {
      const now = new Date().getTime();
      const deadline = new Date(session.deadline!).getTime();
      const diffSeconds = Math.max(0, Math.floor((deadline - now) / 1000));
      setTimeLeft(diffSeconds);

      // 10-Minute Warning: triggers once when crossing 600 seconds (10m)
      if (diffSeconds <= 600 && diffSeconds > 300 && !warned10mRef.current) {
        warned10mRef.current = true;
        setTimeNotification({
          type: "10min",
          title: "⏰ 10 Minutes Remaining",
          message: "You have 10 minutes left to complete and review your exam. Keep going, you are doing great!",
        });
      }

      // 5-Minute Warning: triggers once when crossing 300 seconds (5m)
      if (diffSeconds <= 300 && diffSeconds > 0 && !warned5mRef.current) {
        warned5mRef.current = true;
        setTimeNotification({
          type: "5min",
          title: "⚠️ 5 Minutes Remaining",
          message: "Only 5 minutes left! Please review any unanswered questions before the exam auto-submits.",
        });
      }

      if (diffSeconds <= 0) {
        clearInterval(interval);
        handleFinalSubmit(); // Auto-submit on expiration
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [session]);

  const handleResumeMock = async (sessionId: string) => {
    try {
      setStarting(true);
      const res = await getPracticeSession(sessionId);
      setSession(res.session);
      setQuestions(res.questions);
      const firstUnanswered = res.questions.findIndex((q) => !q.selectedOptionKey);
      setCurrentIndex(firstUnanswered >= 0 ? firstUnanswered : 0);
      setConflictModalOpen(false);
    } catch (err: any) {
      alert(err?.message || "Failed to resume mock exam.");
    } finally {
      setStarting(false);
    }
  };

  const handleCancelMock = async (sessionId: string) => {
    try {
      setCancellingSessionId(sessionId);
      await cancelPracticeSession(sessionId);
      await loadActiveMocks();
      if (session?.id === sessionId) {
        setSession(null);
        setQuestions([]);
      }
    } catch (err: any) {
      alert(err?.message || "Failed to cancel mock exam.");
    } finally {
      setCancellingSessionId(null);
    }
  };

  const handleLaunchExam = async (presetId: string, cancelPrevious = false) => {
    try {
      setStarting(true);
      const res = await createPracticeSession({
        mode: "MOCK_EXAM",
        preset: presetId,
        cancelPrevious,
      });
      setSession(res.session);
      setQuestions(res.questions);
      setCurrentIndex(0);
      setConflictModalOpen(false);
      await loadActiveMocks();
    } catch (err: any) {
      if (
        err?.error === "ACTIVE_MOCK_EXAM_EXISTS" ||
        err?.error === "MAX_ACTIVE_SESSIONS_EXCEEDED" ||
        (err?.message && (err.message.includes("active sessions") || err.message.includes("active mock") || err.message.includes("finish an existing session")))
      ) {
        setConflictModalOpen(true);
        loadActiveMocks();
      } else {
        alert(err?.message || "Failed to start timed mock exam");
      }
    } finally {
      setStarting(false);
    }
  };

  const handleSelectOption = async (optionKey: string) => {
    if (!session) return;
    const currentQ = questions[currentIndex];
    if (!currentQ) return;

    // Save answer in memory immediately
    const updated = [...questions];
    updated[currentIndex] = {
      ...currentQ,
      selectedOptionKey: optionKey,
    };
    setQuestions(updated);

    // Persist answer to backend
    try {
      await submitQuestionAnswer(session.id, currentQ.questionId, optionKey);
    } catch {
      // rollback or retry
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

  const handleFinalSubmit = async () => {
    if (!session || submitting) return;
    try {
      setSubmitting(true);
      await submitExam(session.id);
      router.push(`/results/${session.id}`);
    } catch (err: any) {
      alert(err?.message || "Failed to submit exam");
    } finally {
      setSubmitting(false);
      setShowSubmitModal(false);
    }
  };

  const formatTimer = (seconds: number | null) => {
    if (seconds === null) return "--:--";
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  // Active Timed Exam Runner
  if (session && questions.length > 0) {
    const currentQ = questions[currentIndex];
    const answeredCount = questions.filter((q) => !!q.selectedOptionKey).length;
    const flaggedCount = questions.filter((q) => q.isFlagged).length;
    const isUrgent = timeLeft !== null && timeLeft < 300; // < 5 mins left

    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />

        <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-6 max-w-6xl mx-auto">
          {/* Exam Header with Countdown Timer & Actions */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-sky-100 pb-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <span className="rounded-full bg-indigo-50 px-2.5 py-0.5 text-[10px] font-mono font-bold text-indigo-700 border border-indigo-200">
                  TIMED EXAM ENGINE
                </span>
                <span className="text-xs text-slate-600 font-semibold">{session.title}</span>
              </div>
              <p className="text-xs text-slate-500 font-medium">
                Question {currentIndex + 1} of {questions.length} • {answeredCount} Answered •{" "}
                {flaggedCount} Flagged
              </p>
            </div>

            <div className="flex items-center gap-4">
              {/* Timer Pill */}
              <div
                className={`flex items-center gap-2 rounded-xl border px-3.5 py-1.5 font-mono text-sm font-bold shadow-sm ${
                  isUrgent
                    ? "border-rose-300 bg-rose-50 text-rose-700 animate-pulse"
                    : "border-slate-200 bg-white text-slate-800"
                }`}
              >
                <Clock className={`h-4 w-4 ${isUrgent ? "text-rose-500" : "text-sky-600"}`} />
                <span>{formatTimer(timeLeft)}</span>
              </div>

              {/* Submit Exam Button */}
              <button
                type="button"
                onClick={() => setShowSubmitModal(true)}
                className="flex items-center gap-1.5 rounded-xl bg-emerald-600 px-4 py-2 text-xs font-semibold text-white hover:bg-emerald-500 shadow-md shadow-emerald-600/25 transition-all cursor-pointer"
              >
                <Send className="h-3.5 w-3.5" />
                <span>Submit Exam</span>
              </button>
            </div>
          </div>

          {/* Timed Notification Warning Banner (10m and 5m alerts) */}
          {timeNotification && (
            <div
              className={`flex items-center justify-between gap-4 rounded-xl border p-4 shadow-md transition-all animate-in fade-in slide-in-from-top-2 ${
                timeNotification.type === "5min"
                  ? "border-rose-300 bg-rose-50/95 text-rose-900"
                  : "border-amber-300 bg-amber-50/95 text-amber-900"
              }`}
            >
              <div className="flex items-center gap-3">
                <span className="text-xl shrink-0">
                  {timeNotification.type === "5min" ? "⚠️" : "⏰"}
                </span>
                <div className="space-y-0.5">
                  <h4 className="font-bold text-xs tracking-tight">
                    {timeNotification.title}
                  </h4>
                  <p className="text-xs opacity-90 leading-relaxed">
                    {timeNotification.message}
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setTimeNotification(null)}
                className="rounded-lg px-2.5 py-1 text-xs font-bold hover:bg-black/5 transition-colors shrink-0"
                aria-label="Dismiss notification"
              >
                ✕ Dismiss
              </button>
            </div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 items-start">
            {/* Main Question Area (3 Cols) */}
            <div className="lg:col-span-3 space-y-6">
              <QuestionCard
                question={currentQ}
                currentIndex={currentIndex}
                totalQuestions={questions.length}
                selectedOptionKey={currentQ.selectedOptionKey}
                isImmediateMode={false} // Deferred Feedback
                onSelectOption={handleSelectOption}
                onToggleFlag={handleToggleFlag}
              />

              {/* Bottom Next/Prev Bar */}
              <div className="flex items-center justify-between pt-2">
                <button
                  type="button"
                  onClick={() => setCurrentIndex((prev) => Math.max(0, prev - 1))}
                  disabled={currentIndex === 0}
                  className="flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-slate-900 disabled:opacity-40 transition-all shadow-sm"
                >
                  <ArrowLeft className="h-4 w-4" /> Previous
                </button>

                <button
                  type="button"
                  onClick={handleToggleFlag}
                  className={`flex items-center gap-1.5 rounded-xl border px-3.5 py-2 text-xs font-semibold transition-all shadow-sm ${
                    currentQ.isFlagged
                      ? "border-amber-300 bg-amber-50 text-amber-800"
                      : "border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                  }`}
                >
                  <Flag className={`h-3.5 w-3.5 ${currentQ.isFlagged ? "fill-amber-500 text-amber-500" : ""}`} />
                  <span>{currentQ.isFlagged ? "Marked for Review" : "Flag for Review"}</span>
                </button>

                {currentIndex < questions.length - 1 ? (
                  <button
                    type="button"
                    onClick={() => setCurrentIndex((prev) => Math.min(questions.length - 1, prev + 1))}
                    className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-5 py-2 text-xs font-semibold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/25 transition-all"
                  >
                    Next <ArrowRight className="h-4 w-4" />
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => setShowSubmitModal(true)}
                    className="flex items-center gap-1.5 rounded-xl bg-emerald-600 px-5 py-2 text-xs font-semibold text-white hover:bg-emerald-500 shadow-md shadow-emerald-600/25 transition-all"
                  >
                    Review &amp; Submit <Send className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>

            {/* Question Matrix Jump Grid (1 Col) */}
            <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 shadow-lg shadow-sky-100/60 space-y-4">
              <span className="block text-xs font-bold uppercase tracking-wider text-slate-600">
                Question Navigator
              </span>

              <div className="grid grid-cols-5 gap-1.5">
                {questions.map((q, idx) => {
                  const isCurrent = idx === currentIndex;
                  const isAnswered = !!q.selectedOptionKey;
                  const isFlagged = q.isFlagged;

                  let style = "border-slate-200 bg-slate-50 text-slate-600 hover:bg-slate-100";
                  if (isCurrent) {
                    style = "border-sky-500 bg-sky-50 text-sky-800 font-bold ring-2 ring-sky-500/30";
                  } else if (isFlagged) {
                    style = "border-amber-300 bg-amber-50 text-amber-800 font-semibold";
                  } else if (isAnswered) {
                    style = "border-emerald-300 bg-emerald-50 text-emerald-800 font-semibold";
                  }

                  return (
                    <button
                      key={q.questionId}
                      type="button"
                      onClick={() => setCurrentIndex(idx)}
                      className={`relative flex h-8 w-8 items-center justify-center rounded-lg border text-xs transition-all ${style}`}
                    >
                      {idx + 1}
                      {isFlagged && (
                        <span className="absolute -top-1 -right-1 h-2 w-2 rounded-full bg-amber-500" />
                      )}
                    </button>
                  );
                })}
              </div>

              {/* Legend */}
              <div className="border-t border-slate-100 pt-3 space-y-1.5 text-[11px] text-slate-500">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded bg-emerald-100 border border-emerald-300" />
                  <span>Answered</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded bg-amber-100 border border-amber-300" />
                  <span>Flagged for Review</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded bg-slate-100 border border-slate-300" />
                  <span>Unanswered</span>
                </div>
              </div>
            </div>
          </div>

          {/* Submit Confirmation Dialog Modal */}
          {showSubmitModal && (
            <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm p-4">
              <div className="w-full max-w-md rounded-2xl border border-sky-100 bg-white p-6 shadow-2xl space-y-5 text-slate-900">
                <div className="space-y-1.5">
                  <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2 font-display">
                    <Send className="h-5 w-5 text-emerald-600" />
                    Submit Final Exam?
                  </h3>
                  <p className="text-xs text-slate-500 leading-relaxed font-normal">
                    Once submitted, your answers will be locked, scored, and full distractor rationales
                    will be unlocked for review.
                  </p>
                </div>

                <div className="rounded-xl border border-slate-200 bg-slate-50 p-4 space-y-2 text-xs">
                  <div className="flex justify-between">
                    <span className="text-slate-600">Total Questions:</span>
                    <strong className="text-slate-900">{questions.length}</strong>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-600">Answered:</span>
                    <strong className="text-emerald-700">{answeredCount}</strong>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-600">Unanswered:</span>
                    <strong className={questions.length - answeredCount > 0 ? "text-amber-800 font-bold" : "text-slate-700"}>
                      {questions.length - answeredCount}
                    </strong>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-600">Flagged for Review:</span>
                    <strong className="text-amber-700">{flaggedCount}</strong>
                  </div>
                </div>

                {questions.length - answeredCount > 0 && (
                  <div className="flex items-start gap-2.5 rounded-xl border border-amber-200 bg-amber-50 p-3 text-xs text-amber-900">
                    <AlertTriangle className="h-4 w-4 shrink-0 text-amber-600 mt-0.5" />
                    <span>
                      You have {questions.length - answeredCount} unanswered questions. Unanswered questions will be scored as 0.
                    </span>
                  </div>
                )}

                <div className="flex items-center justify-end gap-3 pt-2">
                  <button
                    type="button"
                    onClick={() => setShowSubmitModal(false)}
                    disabled={submitting}
                    className="rounded-xl border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-sm"
                  >
                    Keep Working
                  </button>
                  <button
                    type="button"
                    onClick={handleFinalSubmit}
                    disabled={submitting}
                    className="flex items-center gap-1.5 rounded-xl bg-emerald-600 px-5 py-2 text-xs font-semibold text-white hover:bg-emerald-500 shadow-md shadow-emerald-600/25 disabled:opacity-50"
                  >
                    {submitting && <Loader2 className="h-3.5 w-3.5 animate-spin text-white" />}
                    <span>Confirm &amp; Submit</span>
                  </button>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    );
  }

  // Preset Selection Screen
  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-6xl mx-auto">
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold uppercase tracking-wider text-sky-600">
              Exam Simulation Engine
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3 font-display">
            <GraduationCap className="h-7 w-7 text-sky-600" />
            Authoritative Timed Mock Exams
          </h1>
          <p className="text-sm text-slate-600 leading-relaxed font-normal">
            Test yourself under authentic examination conditions: strict server countdown timer, deferred feedback,
            and question-by-question review with distractor rationales upon submission.
          </p>
        </div>

        {/* Active Mock Exam Banner */}
        {activeMocks.length > 0 && (
          <div className="rounded-2xl border border-indigo-200 bg-gradient-to-r from-indigo-50 via-sky-50 to-purple-50/40 p-5 sm:p-6 shadow-sm space-y-4 animate-in fade-in">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="flex h-2.5 w-2.5 rounded-full bg-indigo-500 animate-pulse" />
                  <span className="text-xs font-bold uppercase tracking-wider text-indigo-700">
                    Timed Mock Exam in Progress
                  </span>
                </div>
                <h3 className="text-base font-bold text-slate-900">
                  {activeMocks[0].title || "Timed Mock Exam"}
                </h3>
                <p className="text-xs text-slate-600">
                  Started {new Date(activeMocks[0].startedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} • Preset: {activeMocks[0].examPreset || "MOCK_EXAM"}
                </p>
              </div>

              <div className="flex flex-wrap items-center gap-2.5">
                <button
                  type="button"
                  onClick={() => handleResumeMock(activeMocks[0].id)}
                  disabled={starting}
                  className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 px-4 py-2 text-xs font-bold text-white hover:from-indigo-700 hover:to-purple-700 shadow-md shadow-indigo-500/20 transition-all cursor-pointer"
                >
                  {starting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Play className="h-3.5 w-3.5 fill-current" />}
                  <span>Continue Mock Exam</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleCancelMock(activeMocks[0].id)}
                  disabled={cancellingSessionId === activeMocks[0].id}
                  className="flex items-center gap-1.5 rounded-xl border border-rose-200 bg-white px-3.5 py-2 text-xs font-semibold text-rose-600 hover:bg-rose-50 shadow-xs transition-all cursor-pointer"
                >
                  {cancellingSessionId === activeMocks[0].id ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <X className="h-3.5 w-3.5" />
                  )}
                  <span>Cancel / Discard Exam</span>
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Preset 1: Module 1 */}
          <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 flex flex-col justify-between hover:border-sky-300 hover:shadow-lg hover:shadow-sky-100/60 shadow-sm transition-all">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="rounded-full bg-sky-50 px-2.5 py-0.5 text-[11px] font-semibold text-sky-700 border border-sky-200">
                  PRESET 1
                </span>
                <span className="flex items-center gap-1 font-mono text-xs text-slate-500 font-semibold">
                  <Clock className="h-3.5 w-3.5 text-sky-600" /> 40 Mins
                </span>
              </div>
              <h3 className="text-lg font-bold text-slate-900 font-display">Module 1 Midterm Exam</h3>
              <p className="text-xs text-slate-600 leading-relaxed font-normal">
                Linux Fundamentals, Bash Scripting, Git Operations, Virtualization Technologies, and Network Protocols.
              </p>
              <div className="pt-2 text-xs font-mono text-slate-500">
                • 30 Multiple-Choice Questions<br />
                • Pass Threshold: 75%
              </div>
            </div>

            <button
              onClick={() => handleLaunchExam("MODULE_1")}
              disabled={starting}
              className="mt-6 flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 py-2.5 text-xs font-semibold text-white hover:from-sky-600 hover:to-indigo-700 transition-all shadow-md shadow-sky-500/25 disabled:opacity-50"
            >
              {starting ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <ChevronRight className="h-4 w-4" />}
              <span>Launch Module 1 Exam</span>
            </button>
          </div>

          {/* Preset 2: Module 2 */}
          <div className="rounded-2xl border border-indigo-100 bg-white/95 p-6 flex flex-col justify-between hover:border-indigo-300 hover:shadow-lg hover:shadow-indigo-100/60 shadow-sm transition-all">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="rounded-full bg-indigo-50 px-2.5 py-0.5 text-[11px] font-semibold text-indigo-700 border border-indigo-200">
                  PRESET 2
                </span>
                <span className="flex items-center gap-1 font-mono text-xs text-slate-500 font-semibold">
                  <Clock className="h-3.5 w-3.5 text-indigo-600" /> 40 Mins
                </span>
              </div>
              <h3 className="text-lg font-bold text-slate-900 font-display">Module 2 Cloud & Security Exam</h3>
              <p className="text-xs text-slate-600 leading-relaxed font-normal">
                Cloud Architecture, AWS IAM & Security Policies, VPC Networking, Compute/Lambda, and Storage/Databases.
              </p>
              <div className="pt-2 text-xs font-mono text-slate-500">
                • 30 Multiple-Choice Questions<br />
                • Pass Threshold: 75%
              </div>
            </div>

            <button
              onClick={() => handleLaunchExam("MODULE_2")}
              disabled={starting}
              className="mt-6 flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-indigo-500 via-indigo-600 to-blue-600 py-2.5 text-xs font-semibold text-white hover:from-indigo-600 hover:to-blue-700 transition-all shadow-md shadow-indigo-500/25 disabled:opacity-50"
            >
              {starting ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <ChevronRight className="h-4 w-4" />}
              <span>Launch Module 2 Exam</span>
            </button>
          </div>

          {/* Preset 3: Comprehensive Final */}
          <div className="rounded-2xl border border-purple-200 bg-gradient-to-b from-white to-purple-50/40 p-6 flex flex-col justify-between hover:border-purple-300 hover:shadow-lg hover:shadow-purple-100/60 shadow-sm transition-all">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="rounded-full bg-purple-100 px-2.5 py-0.5 text-[11px] font-semibold text-purple-800 border border-purple-200">
                  FINAL BENCHMARK
                </span>
                <span className="flex items-center gap-1 font-mono text-xs text-purple-700 font-semibold">
                  <Clock className="h-3.5 w-3.5 text-purple-600" /> 80 Mins
                </span>
              </div>
              <h3 className="text-lg font-bold text-slate-900 font-display">Comprehensive Final Mock Exam</h3>
              <p className="text-xs text-slate-600 leading-relaxed font-normal">
                Full 10-topic curriculum assessment with balanced distribution across definitions, CLI commands, comparisons, and scenarios.
              </p>
              <div className="pt-2 text-xs font-mono text-slate-500">
                • 60 Multiple-Choice Questions<br />
                • Pass Threshold: 80%
              </div>
            </div>

            <button
              onClick={() => handleLaunchExam("COMPREHENSIVE_FINAL")}
              disabled={starting}
              className="mt-6 flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-purple-600 to-indigo-600 py-2.5 text-xs font-semibold text-white hover:from-purple-700 hover:to-indigo-700 transition-all shadow-md shadow-purple-500/25 disabled:opacity-50"
            >
              {starting ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <ChevronRight className="h-4 w-4" />}
              <span>Launch Comprehensive Final</span>
            </button>
          </div>
        </div>
        {/* Mock Exam Conflict / Limit Modal */}
        {conflictModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4 animate-in fade-in">
            <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl space-y-4 border border-slate-100">
              <div className="flex items-center gap-3 text-slate-900">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 border border-indigo-200">
                  <GraduationCap className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-base font-bold">Active Mock Exam in Progress</h3>
                  <p className="text-xs text-slate-500">You already have an active exam open.</p>
                </div>
              </div>

              <p className="text-xs text-slate-600 leading-relaxed">
                You currently have an active timed mock exam in progress. Would you like to continue answering questions, or cancel it to start a new exam?
              </p>

              <div className="space-y-2 pt-2">
                {activeMocks.length > 0 && (
                  <button
                    type="button"
                    onClick={() => handleResumeMock(activeMocks[0].id)}
                    className="w-full flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 px-4 py-2.5 text-xs font-bold text-white hover:from-indigo-700 hover:to-purple-700 shadow-md shadow-indigo-500/20 cursor-pointer"
                  >
                    <Play className="h-4 w-4 fill-current" /> Continue Answering Current Exam
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => handleLaunchExam("MODULE_1", true)}
                  className="w-full flex items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-xs font-bold text-slate-700 hover:bg-slate-50 cursor-pointer"
                >
                  <RotateCcw className="h-4 w-4" /> Cancel Old Exam & Start Fresh
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

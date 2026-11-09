"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Shield,
  Search,
  Filter,
  Sparkles,
  Bot,
  RefreshCw,
  Layers,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
  ChevronRight,
  Database,
  ArrowRight,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";
import { FormattedStem } from "@/components/CodeSnippet";
import {
  fetchCurrentUser,
  fetchCurriculum,
  apiFetch,
  User,
  CurriculumResponse,
  getStoredAccessToken,
} from "@/lib/api";

interface AdminQuestion {
  id: string;
  topicId: string;
  objectiveId: string;
  stem: string;
  questionStyle: string;
  difficulty: string;
  sourceType: string;
  sourceReference: string;
  status: string;
  createdAt: string;
}

interface AdminBankStats {
  totalQuestions: number;
  sourceBreakdown: {
    seedBank: number;
    aiGemini: number;
    aiGroq: number;
  };
  generationStats: {
    totalRecords: number;
    failoverCount: number;
  };
}

interface GenerationRecord {
  id: string;
  provider: string;
  modelId: string;
  failoverUsed: boolean;
  topicId: string;
  questionStyle: string;
  requestedCount: number;
  acceptedCount: number;
  rejectedCount: number;
  durationMs: number;
  createdAt: string;
}

export default function AdminPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [curriculum, setCurriculum] = useState<CurriculumResponse | null>(null);
  const [loading, setLoading] = useState(true);

  // Stats & Questions
  const [stats, setStats] = useState<AdminBankStats | null>(null);
  const [questions, setQuestions] = useState<AdminQuestion[]>([]);
  const [records, setRecords] = useState<GenerationRecord[]>([]);
  const [totalQuestions, setTotalQuestions] = useState(0);

  // Filters
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTopic, setSelectedTopic] = useState("");
  const [selectedStyle, setSelectedStyle] = useState("");
  const [selectedSource, setSelectedSource] = useState("");
  const [selectedStatus, setSelectedStatus] = useState("ALL");
  const [pageOffset, setPageOffset] = useState(0);

  // Batch Generation Form
  const [genTopic, setGenTopic] = useState("topic-1-cloud-intro");
  const [genStyle, setGenStyle] = useState("SCENARIO");
  const [genDiff, setGenDiff] = useState("INTERMEDIATE");
  const [genCount, setGenCount] = useState(3);
  const [generating, setGenerating] = useState(false);
  const [genFeedback, setGenFeedback] = useState<string | null>(null);

  useEffect(() => {
    const token = getStoredAccessToken();
    if (!token) {
      router.push("/login");
      return;
    }

    fetchCurrentUser()
      .then((u) => {
        setUser(u);
        if (u.role !== "ADMIN") {
          setLoading(false);
          return;
        }
        return Promise.all([
          fetchCurriculum().then(setCurriculum),
          loadBankStats(),
          loadQuestions(0),
          loadGenerationRecords(),
        ]);
      })
      .finally(() => setLoading(false));
  }, [router]);

  const loadBankStats = async () => {
    try {
      const data = await apiFetch<AdminBankStats>("/api/admin/stats");
      setStats(data);
    } catch {
      // ignore
    }
  };

  const loadQuestions = async (offset = 0) => {
    try {
      const params = new URLSearchParams();
      if (selectedTopic) params.set("topicId", selectedTopic);
      if (selectedStyle && selectedStyle !== "ALL") params.set("style", selectedStyle);
      if (selectedSource && selectedSource !== "ALL") params.set("sourceType", selectedSource);
      if (selectedStatus && selectedStatus !== "ALL") params.set("status", selectedStatus);
      if (searchQuery) params.set("q", searchQuery);
      params.set("limit", "20");
      params.set("offset", offset.toString());

      const data = await apiFetch<{ total: number; questions: AdminQuestion[] }>(
        `/api/admin/questions?${params.toString()}`
      );
      setQuestions(data.questions || []);
      setTotalQuestions(data.total || 0);
      setPageOffset(offset);
    } catch {
      // ignore
    }
  };

  const loadGenerationRecords = async () => {
    try {
      const data = await apiFetch<{ records: GenerationRecord[] }>("/api/admin/ai/records?limit=10");
      setRecords(data.records || []);
    } catch {
      // ignore
    }
  };

  const handleToggleStatus = async (q: AdminQuestion) => {
    const nextStatus = q.status === "APPROVED" ? "RETIRED" : "APPROVED";
    try {
      await apiFetch(`/api/admin/questions/${q.id}`, {
        method: "PATCH",
        body: JSON.stringify({ status: nextStatus }),
      });
      setQuestions((prev) =>
        prev.map((item) => (item.id === q.id ? { ...item, status: nextStatus } : item))
      );
    } catch (err: any) {
      alert(err?.message || "Failed to update question status");
    }
  };

  const handleTriggerGeneration = async (e: React.FormEvent) => {
    e.preventDefault();
    setGenFeedback(null);
    try {
      setGenerating(true);
      const res = await apiFetch<{
        success: boolean;
        acceptedCount: number;
        record: GenerationRecord;
      }>("/api/admin/ai/generate", {
        method: "POST",
        body: JSON.stringify({
          topicId: genTopic,
          questionStyle: genStyle,
          difficulty: genDiff,
          count: genCount,
        }),
      });

      setGenFeedback(
        `Successfully generated ${res.acceptedCount} questions via ${res.record?.provider || "AI"} (Failover: ${
          res.record?.failoverUsed ? "Yes" : "No"
        }) in ${res.record?.durationMs}ms.`
      );
      await Promise.all([loadBankStats(), loadQuestions(0), loadGenerationRecords()]);
    } catch (err: any) {
      setGenFeedback(`Generation error: ${err?.message || "Failed"}`);
    } finally {
      setGenerating(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />
        <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-sky-600" />
          <p className="text-sm text-slate-600">Verifying administrator authorization...</p>
        </div>
      </div>
    );
  }

  if (user?.role !== "ADMIN") {
    return (
      <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900">
        <LearnerNav />
        <div className="mx-auto max-w-md text-center py-20 space-y-4">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-rose-50 text-rose-600 border border-rose-200">
            <Shield className="h-7 w-7" />
          </div>
          <h2 className="font-display text-xl font-bold text-slate-900">Administrator Access Required</h2>
          <p className="text-xs text-slate-600 leading-relaxed">
            This management console is restricted to CloudPrep administrators.
          </p>
          <Link
            href="/dashboard"
            className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-4 py-2.5 text-xs font-semibold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/25"
          >
            Return to Learner Dashboard
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-sky-100 pb-6">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="rounded bg-sky-100 px-2 py-0.5 text-[10px] font-mono font-semibold text-sky-700 border border-sky-200">
                ADMIN CONSOLE
              </span>
            </div>
            <h1 className="font-display text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3">
              <Shield className="h-7 w-7 text-sky-600" /> Question Bank &amp; Dual-AI Router
            </h1>
            <p className="text-sm text-slate-600">
              Manage the 1,040+ question bank, trigger on-demand Dual-AI generation (Gemini + Groq), and review failovers.
            </p>
          </div>
        </div>

        {/* Bank Growth Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600">Total Approved Bank</span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight">{stats?.totalQuestions ?? 0}</div>
            <p className="text-[11px] text-slate-500">10 Curriculum Topics</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600">Pre-Seeded Bank</span>
            <div className="text-2xl font-bold text-slate-900 tracking-tight">
              {stats?.sourceBreakdown.seedBank ?? 1040}
            </div>
            <p className="text-[11px] text-slate-500">Curriculum Grounded</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600">AI Generated (Gemini)</span>
            <div className="text-2xl font-bold text-sky-600 tracking-tight">
              {stats?.sourceBreakdown.aiGemini ?? 0}
            </div>
            <p className="text-[11px] text-slate-500">Google Generative Language</p>
          </div>

          <div className="rounded-2xl border border-sky-100 bg-white/95 p-5 space-y-1 shadow-md shadow-sky-100/50">
            <span className="text-xs font-medium text-slate-600">AI Failover (Groq)</span>
            <div className="text-2xl font-bold text-indigo-600 tracking-tight">
              {stats?.sourceBreakdown.aiGroq ?? 0}
            </div>
            <p className="text-[11px] text-slate-500">Groq Cloud LLM</p>
          </div>
        </div>

        {/* On-Demand Batch AI Generator */}
        <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-4 shadow-lg shadow-sky-100/50 backdrop-blur-xl">
          <div className="flex items-center justify-between border-b border-slate-100 pb-3">
            <h2 className="font-display text-sm font-bold text-slate-900 flex items-center gap-2">
              <Bot className="h-4 w-4 text-sky-600" /> Trigger On-Demand Dual-AI Generation
            </h2>
            <span className="text-xs text-slate-500">Gemini Flash with Groq Auto-Failover</span>
          </div>

          {genFeedback && (
            <div className="rounded-xl border border-sky-200 bg-sky-50 p-3 text-xs text-sky-800">
              {genFeedback}
            </div>
          )}

          <form onSubmit={handleTriggerGeneration} className="grid grid-cols-1 sm:grid-cols-5 gap-3 items-end">
            <div className="sm:col-span-2 space-y-1">
              <label className="text-[11px] font-semibold text-slate-700">Topic</label>
              <select
                value={genTopic}
                onChange={(e) => setGenTopic(e.target.value)}
                className="w-full rounded-xl border border-slate-200 bg-slate-50/80 px-3 py-2 text-xs text-slate-900 focus:border-sky-500 focus:bg-white focus:outline-none"
              >
                {curriculum?.topics.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.weekCode} — {t.title}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-1">
              <label className="text-[11px] font-semibold text-slate-700">Style</label>
              <select
                value={genStyle}
                onChange={(e) => setGenStyle(e.target.value)}
                className="w-full rounded-xl border border-slate-200 bg-slate-50/80 px-3 py-2 text-xs text-slate-900 focus:border-sky-500 focus:bg-white focus:outline-none"
              >
                <option value="SCENARIO">Scenario</option>
                <option value="COMMAND_SYNTAX">Command Syntax</option>
                <option value="COMPARISON">Comparison</option>
                <option value="DEFINITION">Definition</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="text-[11px] font-semibold text-slate-700">Count</label>
              <select
                value={genCount}
                onChange={(e) => setGenCount(Number(e.target.value))}
                className="w-full rounded-xl border border-slate-200 bg-slate-50/80 px-3 py-2 text-xs text-slate-900 focus:border-sky-500 focus:bg-white focus:outline-none"
              >
                {[1, 2, 3, 5, 10].map((c) => (
                  <option key={c} value={c}>
                    {c} Questions
                  </option>
                ))}
              </select>
            </div>

            <button
              type="submit"
              disabled={generating}
              className="flex items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-4 py-2 text-xs font-semibold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/25 disabled:opacity-50"
            >
              {generating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Sparkles className="h-3.5 w-3.5" />}
              <span>Generate Questions</span>
            </button>
          </form>
        </div>

        {/* Question Bank Search & Filter Table */}
        <div className="rounded-2xl border border-sky-100 bg-white/95 p-6 space-y-4 shadow-xl shadow-sky-100/50 backdrop-blur-xl">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 pb-4">
            <h2 className="font-display text-base font-bold text-slate-900 flex items-center gap-2">
              <Database className="h-4 w-4 text-sky-600" /> Browse Question Bank ({totalQuestions} Total)
            </h2>

            {/* Keyword Search */}
            <div className="flex items-center gap-2">
              <div className="relative">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && loadQuestions(0)}
                  placeholder="Search stem keywords..."
                  className="rounded-xl border border-slate-200 bg-slate-50/80 pl-8 pr-3 py-1.5 text-xs text-slate-900 placeholder-slate-400 focus:border-sky-500 focus:bg-white focus:outline-none"
                />
                <Search className="absolute left-2.5 top-2 h-3.5 w-3.5 text-slate-400" />
              </div>
              <button
                type="button"
                onClick={() => loadQuestions(0)}
                className="rounded-xl bg-slate-100 border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-200 transition-colors"
              >
                Search
              </button>
            </div>
          </div>

          {/* Filter Bar */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
            <select
              value={selectedTopic}
              onChange={(e) => {
                setSelectedTopic(e.target.value);
              }}
              className="rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1.5 text-xs text-slate-800 focus:border-sky-500 focus:outline-none"
            >
              <option value="">All Topics</option>
              {curriculum?.topics.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.weekCode} — {t.title}
                </option>
              ))}
            </select>

            <select
              value={selectedStyle}
              onChange={(e) => setSelectedStyle(e.target.value)}
              className="rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1.5 text-xs text-slate-800 focus:border-sky-500 focus:outline-none"
            >
              <option value="">All Styles</option>
              <option value="DEFINITION">Definition</option>
              <option value="COMMAND_SYNTAX">Command Syntax</option>
              <option value="COMPARISON">Comparison</option>
              <option value="SCENARIO">Scenario</option>
            </select>

            <select
              value={selectedSource}
              onChange={(e) => setSelectedSource(e.target.value)}
              className="rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1.5 text-xs text-slate-800 focus:border-sky-500 focus:outline-none"
            >
              <option value="">All Sources</option>
              <option value="SEED_BANK">Pre-Seeded Bank</option>
              <option value="AI_GEMINI">AI Gemini</option>
              <option value="AI_GROQ">AI Groq (Failover)</option>
            </select>

            <button
              onClick={() => loadQuestions(0)}
              className="rounded-lg bg-sky-50 border border-sky-200 px-3 py-1.5 text-xs font-semibold text-sky-700 hover:bg-sky-100 transition-colors"
            >
              Apply Filters
            </button>
          </div>

          {/* Questions Table */}
          <div className="overflow-x-auto divide-y divide-slate-100">
            {questions.map((q) => (
              <div key={q.id} className="py-3.5 flex flex-col sm:flex-row sm:items-start justify-between gap-4 group">
                <div className="space-y-1.5 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-mono text-[11px] text-slate-400">{q.id}</span>
                    <span className="rounded bg-sky-50 px-2 py-0.5 text-[10px] font-mono font-semibold text-sky-700 border border-sky-100">
                      {q.questionStyle}
                    </span>
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-medium text-slate-600">
                      {q.difficulty}
                    </span>
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-mono text-slate-500">
                      {q.sourceType}
                    </span>
                  </div>

                  <p className="text-xs text-slate-800 font-medium leading-relaxed">
                    {q.stem}
                  </p>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <span
                    className={`rounded-full px-2.5 py-0.5 text-[10px] font-bold border ${
                      q.status === "APPROVED"
                        ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                        : "border-slate-200 bg-slate-100 text-slate-500"
                    }`}
                  >
                    {q.status}
                  </span>

                  <button
                    onClick={() => handleToggleStatus(q)}
                    className="rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1 text-[11px] font-medium text-slate-700 hover:bg-slate-100 hover:text-slate-900 transition-colors"
                  >
                    {q.status === "APPROVED" ? "Retire" : "Approve"}
                  </button>
                </div>
              </div>
            ))}

            {questions.length === 0 && (
              <p className="text-center text-xs text-slate-500 py-8">
                No questions found matching your filter criteria.
              </p>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

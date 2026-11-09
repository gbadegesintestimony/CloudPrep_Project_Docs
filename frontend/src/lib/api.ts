// CloudPrep Frontend API Client & Type Definitions

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

// --- Domain Models & DTOs ---

export interface User {
  id: string;
  email: string;
  name: string;
  role: "LEARNER" | "GUEST" | "ADMIN" | string;
  isGuest: boolean;
  status: string;
  createdAt: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface HealthResponse {
  status: string;
  dbActive: boolean;
  dbDriver: string;
  topicCount: number;
  objectiveCount: number;
  bankCount: number;
  timestamp: string;
}

export interface LearningObjectiveDTO {
  id: string;
  topicId: string;
  code: string;
  title: string;
  summary: string;
  sourceNotes: string;
  curriculumExcerpt?: string;
}

export interface TopicDTO {
  id: string;
  slug: string;
  title: string;
  moduleTitle: string;
  moduleNumber: number;
  weekCode: string;
  sourcePdf: string;
  description: string;
  displayRank: number;
  questionCount: number;
  styleCounts: Record<string, number>;
  objectives: LearningObjectiveDTO[];
}

export interface ExamPresetDTO {
  id: string;
  title: string;
  description: string;
  topicIds: string[];
  questionCount: number;
  timeLimitMinutes: number;
}

export interface CurriculumResponse {
  topics: TopicDTO[];
  examPresets: ExamPresetDTO[];
  totalBankQuestions: number;
  supportedStyles: string[];
  supportedDifficulties: string[];
}

export interface ShuffledOptionDTO {
  key: string; // "A" | "B" | "C" | "D"
  text: string;
  distractorRationale?: string;
}

export interface ActiveSessionQuestionDTO {
  questionId: string;
  ordering: number;
  topicId: string;
  objectiveId: string;
  stem: string;
  questionStyle: "DEFINITION" | "COMMAND_SYNTAX" | "COMPARISON" | "SCENARIO" | string;
  difficulty: "FOUNDATIONAL" | "INTERMEDIATE" | "ADVANCED" | string;
  sourceReference: string;
  isFlagged: boolean;
  isBookmarked: boolean;
  selectedOptionKey?: string;
  options: ShuffledOptionDTO[];
  isCorrect?: boolean;
  correctOptionKey?: string;
  explanation?: string;
}

export interface ReviewedQuestionDTO extends ActiveSessionQuestionDTO {
  topicTitle: string;
  moduleNumber: number;
  weekCode: string;
  sourceType: string;
  isAnswered: boolean;
}

export interface PracticeSessionDTO {
  id: string;
  userId: string;
  mode: "TOPIC_PRACTICE" | "MIXED_PRACTICE" | "REVIEW_MISTAKES" | "BOOKMARK_REVISION" | "MOCK_EXAM" | string;
  examPreset: string;
  title: string;
  topicIds: string[];
  questionStyleFilter: string;
  difficultyFilter: string;
  feedbackMode: "IMMEDIATE" | "DEFERRED";
  totalQuestions: number;
  answeredCount: number;
  unansweredCount: number;
  correctCount: number;
  incorrectCount: number;
  scorePercentage: number;
  passed?: boolean;
  timeLimitSeconds: number;
  remainingSeconds: number;
  status: "ACTIVE" | "COMPLETED" | "EXPIRED" | "ABANDONED" | string;
  startedAt: string;
  submittedAt?: string;
  deadline?: string;
}

export interface CreateSessionRequest {
  mode: string;
  preset?: string;
  topicId?: string;
  topicIds?: string[];
  questionStyle?: string;
  difficulty?: string;
  feedbackMode?: "IMMEDIATE" | "DEFERRED";
  questionCount?: number;
  timeLimitSeconds?: number;
  cancelPrevious?: boolean;
}

export interface BookmarkDTO {
  id: string;
  userId: string;
  questionId: string;
  topicId: string;
  topicTitle: string;
  stem: string;
  questionStyle: string;
  difficulty: string;
  sourceReference: string;
  options: ShuffledOptionDTO[];
  correctOptionKey?: string;
  explanation?: string;
  createdAt: string;
}

export interface BreakdownStat {
  key: string;
  label: string;
  total: number;
  correct: number;
  accuracy: number;
  masteryStatus?: string;
  moduleNumber?: number;
  weekCode?: string;
}

export interface ProgressSummaryDTO {
  totalAnswered: number;
  totalCorrect: number;
  overallAccuracy: number;
  totalBookmarks: number;
  totalCompletedSessions: number;
  activeSessionsCount: number;
  totalBankQuestions: number;
  uniqueQuestionsSeen: number;
  topicBreakdown: BreakdownStat[];
  styleBreakdown: BreakdownStat[];
  weakestTopics: string[];
  strongestTopics: string[];
}

// --- Client Token Management ---

const ACCESS_KEY = "cloudprep_access_token";
const REFRESH_KEY = "cloudprep_refresh_token";
const USER_KEY = "cloudprep_user";

export function getStoredAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS_KEY);
}

export function getStoredRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(REFRESH_KEY);
}

export function getStoredUser(): User | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function storeAuthData(data: AuthTokens): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(ACCESS_KEY, data.accessToken);
  localStorage.setItem(REFRESH_KEY, data.refreshToken);
  localStorage.setItem(USER_KEY, JSON.stringify(data.user));
}

export function clearStoredAuth(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(USER_KEY);
}

// --- Authenticated Fetch Helper with Auto-Refresh ---

export async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;
  const token = getStoredAccessToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let res = await fetch(url, { ...options, headers });

  // Handle Token Expiration (401) with automatic refresh
  if (res.status === 401 && getStoredRefreshToken()) {
    const refreshed = await refreshAuthTokens();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${refreshed.accessToken}`;
      res = await fetch(url, { ...options, headers });
    } else {
      clearStoredAuth();
    }
  }

  if (!res.ok) {
    let errMessage = `HTTP error ${res.status}`;
    let errJson: any = {};
    try {
      errJson = await res.json();
      errMessage = errJson.message || errJson.error || errMessage;
    } catch {
      // ignore
    }
    const err: any = new Error(errMessage);
    err.status = res.status;
    err.error = errJson?.error;
    err.data = errJson;
    throw err;
  }

  return res.json();
}

async function refreshAuthTokens(): Promise<AuthTokens | null> {
  const refreshToken = getStoredRefreshToken();
  if (!refreshToken) return null;

  try {
    const res = await fetch(`${API_BASE_URL}/api/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    if (!res.ok) return null;
    const data = await res.json();
    const currentUser = getStoredUser();
    if (currentUser) {
      const updated: AuthTokens = {
        accessToken: data.accessToken,
        refreshToken: data.refreshToken || refreshToken,
        user: currentUser,
      };
      storeAuthData(updated);
      return updated;
    }
    return null;
  } catch {
    return null;
  }
}

// --- API Service Functions ---

// 1. Health & Server Cold-Start
export async function pingHealth(): Promise<HealthResponse | null> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/health`, {
      method: "GET",
      cache: "no-store",
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

// 2. Curriculum Catalog
export async function fetchCurriculum(): Promise<CurriculumResponse> {
  const res = await fetch(`${API_BASE_URL}/api/curriculum`, {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load curriculum from CloudPrep API");
  }
  return res.json();
}

// 3. Auth Endpoints
export async function registerUser(name: string, email: string, password: string): Promise<AuthTokens> {
  const res = await fetch(`${API_BASE_URL}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, email, password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || "Failed to register account");
  }
  const data: AuthTokens = await res.json();
  storeAuthData(data);
  return data;
}

export async function loginUser(email: string, password: string): Promise<AuthTokens> {
  const res = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || "Invalid credentials");
  }
  const data: AuthTokens = await res.json();
  storeAuthData(data);
  return data;
}

export async function loginAsGuest(): Promise<AuthTokens> {
  const res = await fetch(`${API_BASE_URL}/api/auth/guest`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || "Failed to initiate 1-Click Guest session");
  }
  const data: AuthTokens = await res.json();
  storeAuthData(data);
  return data;
}

export async function fetchCurrentUser(): Promise<User> {
  const data = await apiFetch<{ user: User }>("/api/me");
  if (data.user) {
    if (typeof window !== "undefined") {
      localStorage.setItem(USER_KEY, JSON.stringify(data.user));
    }
  }
  return data.user;
}

// 4. Practice & Mock Exam Sessions
export async function createPracticeSession(req: CreateSessionRequest): Promise<{
  session: PracticeSessionDTO;
  questions: ActiveSessionQuestionDTO[];
}> {
  return apiFetch("/api/practice-sessions", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function getPracticeSession(sessionId: string): Promise<{
  session: PracticeSessionDTO;
  questions: ActiveSessionQuestionDTO[];
}> {
  return apiFetch(`/api/practice-sessions/${sessionId}`);
}

export async function fetchActivePracticeSessions(): Promise<PracticeSessionDTO[]> {
  const data = await apiFetch<{ sessions: PracticeSessionDTO[] }>("/api/practice-sessions?status=ACTIVE");
  return data.sessions || [];
}

export async function cancelPracticeSession(sessionId: string): Promise<void> {
  await apiFetch(`/api/practice-sessions/${sessionId}/cancel`, {
    method: "POST",
  });
}

export async function cancelAllActiveSessions(): Promise<{ affectedCount: number }> {
  return apiFetch("/api/practice-sessions/cancel-active", {
    method: "POST",
  });
}

export async function submitQuestionAnswer(
  sessionId: string,
  questionId: string,
  selectedOptionKey: string
): Promise<{
  success: boolean;
  answeredCount: number;
  unansweredCount: number;
  isCorrect?: boolean;
  correctOptionKey?: string;
  explanation?: string;
}> {
  return apiFetch(`/api/practice-sessions/${sessionId}/responses/${questionId}`, {
    method: "PUT",
    body: JSON.stringify({ selectedOptionKey }),
  });
}

export async function submitExam(sessionId: string): Promise<{
  session: PracticeSessionDTO;
  results: ReviewedQuestionDTO[];
  summary: {
    totalQuestions: number;
    answeredCount: number;
    correctCount: number;
    scorePercentage: number;
    passed?: boolean;
    durationSeconds: number;
  };
}> {
  return apiFetch(`/api/practice-sessions/${sessionId}/submit`, {
    method: "POST",
  });
}

export async function getSessionReview(sessionId: string): Promise<{
  session: PracticeSessionDTO;
  results: ReviewedQuestionDTO[];
  questions: ReviewedQuestionDTO[];
}> {
  const data = await apiFetch<any>(`/api/practice-sessions/${sessionId}/review`);
  const list = data?.questions || data?.results || [];
  const rawSession = data?.session || {};
  const correctCount = rawSession.correctCount ?? list.filter((q: any) => q.isCorrect).length;
  const totalQuestions = rawSession.totalQuestions ?? list.length;
  const calculatedPercent = totalQuestions > 0 ? Math.round((correctCount / totalQuestions) * 100) : 0;
  const scorePercentage = rawSession.scorePercentage ?? rawSession.scorePercent ?? rawSession.score_percent ?? calculatedPercent;

  const session: PracticeSessionDTO = {
    ...rawSession,
    correctCount,
    totalQuestions,
    scorePercentage,
  };

  return {
    session,
    results: list,
    questions: list,
  };
}

// 5. Bookmarks & Progress Summary
export async function fetchBookmarks(): Promise<BookmarkDTO[]> {
  const res = await apiFetch<{ bookmarks: BookmarkDTO[] }>("/api/bookmarks");
  return res.bookmarks || [];
}

export async function toggleBookmark(questionId: string, currentlyBookmarked: boolean): Promise<boolean> {
  if (currentlyBookmarked) {
    await apiFetch(`/api/bookmarks/${questionId}`, { method: "DELETE" });
    return false;
  } else {
    await apiFetch(`/api/bookmarks/${questionId}`, { method: "POST" });
    return true;
  }
}

export async function fetchProgressSummary(): Promise<ProgressSummaryDTO> {
  const raw: any = await apiFetch("/api/progress/summary");

  const topicBreakdown: BreakdownStat[] = (raw?.topicBreakdown || []).map((t: any) => ({
    key: t.topicId || t.slug || t.key || "",
    label: t.title || t.label || t.key || "Topic",
    total: t.attempted ?? t.total ?? 0,
    correct: t.correct ?? 0,
    accuracy: Math.round(t.accuracyPercent ?? t.accuracy ?? 0),
    masteryStatus: t.masteryStatus,
    moduleNumber: t.moduleNumber,
    weekCode: t.weekCode,
  }));

  const styleBreakdown: BreakdownStat[] = (raw?.styleBreakdown || []).map((s: any) => ({
    key: s.questionStyle || s.key || "",
    label: s.questionStyle || s.label || "",
    total: s.attempted ?? s.total ?? 0,
    correct: s.correct ?? 0,
    accuracy: Math.round(s.accuracyPercent ?? s.accuracy ?? 0),
  }));

  const weakestTopics = (raw?.recommendedWeakTopics || []).map((w: any) => w.title || w.topicId || w);

  return {
    totalAnswered: raw?.totalQuestionsAnswered ?? raw?.totalAnswered ?? 0,
    totalCorrect: raw?.totalCorrect ?? 0,
    overallAccuracy: Math.round(raw?.overallAccuracyPercent ?? raw?.overallAccuracy ?? 0),
    totalBookmarks: raw?.bookmarkCount ?? raw?.totalBookmarks ?? 0,
    totalCompletedSessions: raw?.completedSessions ?? raw?.totalCompletedSessions ?? 0,
    activeSessionsCount: raw?.activeSessions ?? raw?.activeSessionsCount ?? 0,
    totalBankQuestions: raw?.totalBankQuestions ?? 1040,
    uniqueQuestionsSeen: raw?.uniqueQuestionsSeen ?? 0,
    topicBreakdown,
    styleBreakdown,
    weakestTopics,
    strongestTopics: raw?.strongestTopics || [],
  };
}

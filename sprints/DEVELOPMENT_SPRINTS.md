# CloudPrep — Development Sprint Plan & Execution Roadmap

**Status:** Ready for Execution  
**Target Stack:** Next.js + TypeScript (Vercel) + **Golang (`Go 1.22+`) & Gin Framework + GORM/`pgx/v5`** (Render) + Supabase PostgreSQL + Dual AI (Google Gemini & Groq)  
**Curriculum Baseline:** 10 Topics across Module 1 & Module 2 (`school notes/` PDFs) + 1,000+ Pre-Seeded Multi-Style Question Bank

---

## Sprint Overview

| Sprint | Title | Primary Goal / Deliverable | Maps to DoD Questions |
|---|---|---|---|
| **Sprint 1** | **Monorepo, Go + Gin Backend, DB Schema & Curriculum Extraction** | Scaffold `backend/` (**Golang + Gin + GORM/`pgx/v5`**) and `frontend/` (Next.js), configure Supabase PostgreSQL models, extract all 8 `school notes/` PDFs into structured curriculum modules/objectives, and expose `GET /api/health` (with Supabase DB heartbeat). | Q1, Q7 |
| **Sprint 2** | **1,000+ Pre-Seeded Multi-Style Question Bank** | Build and seed **1,000+ approved questions** (~100+ per topic across all 10 topics) covering Definitions (`~30%`), Commands & Syntax (`~25%`), Comparisons (`~20%`), and Scenarios (`~25%`) with full explanations, distractor rationales, and course module/week citations. | Q1, Q2, Q5 |
| **Sprint 3** | **Zero-Friction Auth, 1-Click Guest & Core Go Exam Engine** | Implement Name/Email/Password sign-up (no email verification gate), **1-Click Guest Mode**, unseen-first selection, **dynamic per-session option shuffling (`A–D`)**, mock-exam answer concealment, idempotent submission, bookmarks API, and progress analytics in Go + Gin. | Q2, Q4, Q5, Q6 |
| **Sprint 4** | **Dual-AI Engine in Go (Google Gemini + Groq Failover & Bank Growth)** | Wire **Google Gemini (`gemini-2.5-flash`)** and **Groq (`llama-3.3-70b-versatile`)** in Go with automatic failover, JSON schema & curriculum-grounding validation, fingerprint deduplication, automatic database bank persistence, native Go goroutine background replenishment, and `<200ms` zero-downtime fallback to the 1,000+ bank. | Q1, Q2, Q3 |
| **Sprint 5** | **Frontend Foundation, Silent Warm-Up, Auth & Dashboard** | Build Next.js App Router shell, design system (including monospace code typography), silent `GET /api/health` Render cold-start & Supabase anti-pause warm-up on `/` and `/login`, 1-Click Guest / Login / Sign-Up screens, and Learner Dashboard. | Q6, Q7 |
| **Sprint 6** | **Practice Mode, Mock Exam Presets, Code Blocks, Results & Bookmarks UI** | Build Practice UI (with Question Style filter), **Module 1 / Module 2 / Comprehensive Final Mock Exam presets**, monospace code/CLI rendering, authoritative server-deadline timer, Results with course module/week citations, Bookmarks (`/bookmarks`), History, Progress, and Admin UI. | Q1, Q2, Q4, Q5 |
| **Sprint 7** | **Verification against 7 Core Questions, Keep-Alive Cron & Deployment** | Verify all 7 Definition-of-Done questions with automated tests, add scheduled GitHub Actions `/api/health` keep-alive cron (so Supabase never pauses after 7 days), and finalize production deployment configuration for Supabase, Render, and Vercel. | Q1–Q7 |

---

## Detailed Sprint Breakdown

### Sprint 1 — Monorepo Setup, Go + Gin Backend, Database Schema & Curriculum Extraction
**Goal:** Establish the project structure with **Golang + Gin** on the backend and **Next.js** on the frontend, configure the Supabase PostgreSQL schema via GORM/`pgx/v5`, implement the DB-backed health check endpoint, and extract structured curriculum notes from the 8 PDFs in `school notes/`.

#### Tasks
- [x] **1.1 Workspace Scaffolding:**
  - Initialize `backend/` (**Golang `1.22+` + Gin Framework `github.com/gin-gonic/gin` + GORM + `pgx/v5` + `golang-jwt/jwt/v5`**).
  - Initialize `frontend/` (Next.js App Router + TypeScript + Tailwind CSS + shadcn/ui + TanStack Query + Lucide icons).
- [x] **1.2 Go GORM Models & Supabase PostgreSQL Schema Setup:**
  - Define `User` struct (supporting both registered users and 1-click guest accounts via `IsGuest`).
  - Define `CurriculumTopic` (with `ModuleNumber` 1 or 2, `WeekCode`, `Title`, `SourcePdf`) and `LearningObjective` (with extracted slide notes/excerpts).
  - Define `Question` (`Stem`, `QuestionStyle` enum: `DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`, `Difficulty`, `SourceType` enum: `SEED_BANK`, `AI_GEMINI`, `AI_GROQ`, `ADMIN`, `SourceReference`, `ContentFingerprint`, `Explanation`) and `QuestionOption` (`Key`, `Text`, `DistractorRationale`).
  - Define `PracticeSession` (`Mode`, `ExamPreset` enum: `MODULE_1`, `MODULE_2`, `COMPREHENSIVE_FINAL`, `CUSTOM`, `QuestionStyleFilter`, `Deadline`, `Status`, `Score`).
  - Define `SessionQuestion` (including `ShuffledOptionMapping` JSONB storing the per-session permutation of `A, B, C, D` -> canonical option keys, and `IsFlagged`).
  - Define `Response`, `QuestionExposure`, `QuestionBookmark`, `GenerationRecord`, and `AuditEvent`.
- [x] **1.3 Curriculum Extraction from `school notes/`:**
  - Extract key concepts, definitions, commands, comparisons, and scenarios from all 8 PDFs:
    - **Module 1 (Foundations):**
      1. Introduction to Computing (`M1_W1_Introduction to Cloud Computing.pdf`)
      2. Linux Fundamentals (`M1_W2_Linux Fundamentals.pdf`)
      3. Bash Scripting & Automation (`M1_ W4_Bash Scripting & Automation.pdf`)
      4. Git & Version Control (`M1_W3_Git & Version Control - Google Slides.pdf`)
    - **Module 2 (Cloud Architecture, Services & AI):**
      5. Core Compute & IAM (`M2_W1_Core Compute & IAM.pdf`)
      6. Cloud Networking (`M2_W2_Networking & Security.pdf`)
      7. Cloud Storage (`M2_W2_Networking & Security.pdf`)
      8. Databases (`M2_W3_ Databases, Serverless & Application Services.pdf`)
      9. Serverless & Application Services (`M2_W3_ Databases, Serverless & Application Services.pdf`)
      10. Cloud AI & MLOps (`M2_W4_ Cloud AI & MLOps Fundamentals.pdf`)
- [x] **1.4 Public Health, Cold-Start & Supabase Anti-Pause Endpoint (`GET /api/health`):**
  - Implement `GET /api/health` in Gin returning `{ status: "ok", dbActive: true, timestamp, bankCount }`.
  - Execute a lightweight PostgreSQL query (`SELECT 1` / cached question count check) against Supabase so every `/api/health` call registers real database activity and resets Supabase's 7-day inactivity pause timer.

**Sprint 1 Exit Criteria:** Go backend compiles cleanly; GORM auto-migrations create all PostgreSQL tables; all 10 curriculum topics and learning objectives are structured; `GET /api/health` wakes the API and touches Supabase PostgreSQL in `<50ms`.

---

### Sprint 2 — The 1,000+ Pre-Seeded Multi-Style Question Bank
**Goal:** Build and seed **1,000+ curriculum-grounded questions** (~100+ questions per topic across all 10 topics) strictly within the jurisdiction of `school notes/`.

#### Tasks
- [x] **2.1 Author & Structure 1,000+ Questions Across All 10 Topics (~100+ per topic):**
  - Topic 1: Introduction to Computing (`100+` questions)
  - Topic 2: Linux Fundamentals (`100+` questions)
  - Topic 3: Bash Scripting & Automation (`100+` questions)
  - Topic 4: Git & Version Control (`100+` questions)
  - Topic 5: Core Compute & IAM (`100+` questions)
  - Topic 6: Cloud Networking (`100+` questions)
  - Topic 7: Cloud Storage (`100+` questions)
  - Topic 8: Databases (`100+` questions)
  - Topic 9: Serverless & Application Services (`100+` questions)
  - Topic 10: Cloud AI & MLOps (`100+` questions)
- [x] **2.2 Enforce Multi-Style Distribution in Every Topic:**
  - **Definitions & Core Concepts (~30%):** Terminology, cloud deployment/service models, foundational facts.
  - **Commands, Syntax & Service Identification (~25%):** Linux commands, Bash scripts (`#!/bin/bash`, loops, conditionals, exit codes), Git CLI commands/flags, IAM policy syntax, and service identification—formatted with inline backticks and fenced code blocks.
  - **Concept Comparisons & Distinctions (~20%):** Contrasting pairs/groups (e.g., Security Groups vs. NACLs, IaaS vs. PaaS vs. SaaS, `git merge` vs. `git rebase`, SQL vs. NoSQL, Block vs. Object vs. File storage).
  - **Scenarios & Troubleshooting (~25%):** Practical startup/enterprise design, security, cost, and failure troubleshooting grounded in the notes.
- [x] **2.3 Explanations, Distractor Rationales & Course Citations:**
  - Ensure every question includes 4 options (`A–D`), 1 unambiguous correct option, a clear explanation, rationales explaining why each distractor is wrong, a normalized content fingerprint, and an exact `sourceReference` (e.g., `"Module 1, Week 2 — Linux Fundamentals"`).
- [x] **2.4 Idempotent Go Database Seeder (`cmd/seed/main.go`):**
  - Batch-insert all 10 topics, learning objectives, and 1,000+ questions into Supabase PostgreSQL with duplicate-fingerprint protection (`ON CONFLICT DO NOTHING`).

**Sprint 2 Exit Criteria:** Database query confirms `>= 1,000` approved questions evenly distributed across all 10 topics and all 4 question styles (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`).

---

### Sprint 3 — Zero-Friction Auth, 1-Click Guest Mode & Core Exam Engine
**Goal:** Build the backend authentication system, unseen-first question selector, dynamic option shuffler, practice/mock session lifecycle, bookmarking, and progress analytics.

#### Tasks
- [x] **3.1 Zero-Friction Auth & 1-Click Guest Mode (`/api/auth/*`):**
  - `POST /api/auth/register`: Name, Email, Password registration with immediate JWT/session token issuance (no email verification wall).
  - `POST /api/auth/login`: Email & password sign-in.
  - `POST /api/auth/guest`: 1-click Guest Learner creation/sign-in so evaluators and students can start immediately.
  - `GET /api/me` and role guards (`LEARNER`, `GUEST`, `ADMIN`).
- [x] **3.2 Question Selector, Anti-Repetition & Option Shuffling:**
  - Filter by selected topics or **Mock Exam Preset** (`MODULE_1` = Topics 1–4; `MODULE_2` = Topics 5–10; `COMPREHENSIVE_FINAL` = Topics 1–10).
  - Filter by `questionStyle` in Practice Mode (or enforce a balanced mix of all 4 styles in Mock Exams) and `difficulty`.
  - Prioritize unseen questions for the current user (`QuestionExposure` lookup) so returning learners get fresh questions.
  - **Dynamic Option Shuffling (`A–D`):** Generate a randomized option permutation for every `SessionQuestion` and store the mapping in the database so options stay consistent on browser refresh while preventing positional memorization across sessions.
- [x] **3.3 Session Lifecycle & Mock Exam Integrity (`/api/practice-sessions/*`):**
  - `POST /api/practice-sessions`: Create session in `<200ms` from the 1,000+ bank (with authoritative server `deadline` for timed mock exams).
  - `GET /api/practice-sessions/:sessionId`: Return session state and shuffled options. **Strictly strip `correctOptionKey`, `explanation`, and `distractorRationales` on active mock exams.**
  - `PUT /api/practice-sessions/:sessionId/responses/:questionId`: Upsert learner's answer (resolving shuffled option key) and support immediate feedback in Practice Mode when configured.
  - `PUT /api/practice-sessions/:sessionId/questions/:questionId/flag`: Toggle in-session review flag.
  - `POST /api/practice-sessions/:sessionId/submit`: Transactional, **idempotent submission** that scores against the session's shuffled option mapping, updates `QuestionExposure`, and returns the final result without creating duplicates on retry.
  - `GET /api/practice-sessions/:sessionId/review`: Full post-submission review with explanations, distractor rationales, and course module/week citations.
- [x] **3.4 Bookmarks & Progress Analytics (`/api/bookmarks/*`, `/api/progress/summary`):**
  - `GET /api/bookmarks`, `POST /api/bookmarks/:questionId`, `DELETE /api/bookmarks/:questionId` for saving and revising tricky questions.
  - `GET /api/progress/summary`: Accuracy and attempts broken down by curriculum topic, Module 1 vs. Module 2, and Question Style (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`), plus recommended weaker topics.
- [x] **3.5 Security Audit Penetration Testing & Hardening:**
  - Automated penetration audit identifying and remediating 8 attack vectors (C-1 to M-8).
  - Remediated answer leaks via bookmarks and active mock exam redaction (C-1, C-2).
  - Implemented atomic submission state transition preventing race condition exposure corruption (H-1).
  - Configured Render reverse proxy trusted network and dedicated `/api/auth/*` rate limiter (H-2, H-3).
  - Enforced active session quotas, answer locking after immediate feedback (H-4), and mock exam preset immutability (H-5).
  - Verified with 10 automated exploit probes in `sprint3_hardened_test.go` (100% passing).

**Sprint 3 Exit Criteria:** All auth, session, shuffling, mock-exam concealment, idempotent submit, bookmark, progress endpoints, and penetration hardening tests pass with 100% green test suite.

---

### Sprint 4 — Dual-AI Generation Engine (Google Gemini + Groq Failover & Bank Growth)
**Goal:** Wire both **Google Gemini (`gemini-3.8-flash`)** and **Groq (`openai/gpt-oss-120b`)** with automatic failover, strict validation, automatic bank persistence, and zero-downtime fallback to the 1,000+ Question Bank.

#### Tasks
- [x] **4.1 Provider Interface & Dual Adapters:**
  - Defined `QuestionGenerationProvider` interface in `internal/ai/types.go`.
  - Implemented `GeminiProvider` (`internal/ai/gemini_provider.go`) using Google Generative Language API with structured JSON output.
  - Implemented `GroqProvider` (`internal/ai/groq_provider.go`) using Groq Cloud API with JSON object response format.
- [x] **4.2 Failover Router & Prompt Grounding (`MultiProviderAiRouter`):**
  - Assembles curriculum-grounded prompts containing the exact `school notes/` excerpts for the target topic, requested `questionStyle` (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`), difficulty, and recent stems to avoid (`internal/ai/prompt.go`).
  - Routes to primary provider (`AI_PRIMARY_PROVIDER`, default `gemini`); on 429 rate-limit, timeout, or error, **automatically fails over** to secondary provider (`AI_FALLBACK_PROVIDER`, default `groq`).
- [x] **4.3 Automated Quality, Grounding & Similarity Validator:**
  - Verifies valid JSON schema, 4 distinct options (`A–D`), valid `correctOptionKey`, non-empty `explanation` and `distractorRationales`, valid `questionStyle`, and course `sourceReference` (`internal/ai/validator.go`).
  - Computes normalized SHA-256 `contentFingerprint` and checks near-duplicate similarity against existing questions in the bank.
- [x] **4.4 Automatic Bank Persistence, Background Replenishment & Zero-Downtime Fallback:**
  - Saves accepted AI questions directly into the `Question` table (`sourceType: AI_GEMINI` or `AI_GROQ`) and logs `GenerationRecord` so the bank grows beyond 1,000+ items (`internal/ai/router.go`).
  - Triggers non-blocking background replenishment when a learner's unseen pool for a topic runs low (`internal/ai/replenishment.go` wired to `SessionHandler.OnLowUnseen`).
  - If both Gemini and Groq are unconfigured, offline, or out of quota, immediately serves eligible questions from the **1,000+ Question Bank** in `<200ms` with zero error to the learner.
- [x] **4.5 Admin Question Bank & Batch AI Generator Endpoints (`/api/admin/*`):**
  - Search/filter the 1,000+ bank (`GET /api/admin/questions`), inspect AI generation records (`GET /api/admin/ai/records`), edit/approve/retire questions (`PATCH /api/admin/questions/:questionId`), and trigger batch AI generation for any topic/style (`POST /api/admin/ai/generate`).
  - Question bank growth analytics (`GET /api/admin/stats`).

**Sprint 4 Exit Criteria:** Unit & integration tests confirm Gemini generation, automatic failover to Groq when Gemini fails, automatic persistence into the bank, and `<200ms` fallback to the 1,000+ bank when both AI providers are disabled. All tests passing 100%.

---

### Sprint 5 — Frontend Foundation, Silent Warm-Up, Auth & Learner Dashboard
**Goal:** Build the Next.js App Router frontend foundation, design system, silent Render cold-start warm-up ping, zero-friction auth + 1-click Guest Mode UI, and Learner Dashboard.

#### Tasks
- [ ] **5.1 Design System & API Client Setup:**
  - Configure Tailwind CSS tokens, clean dark/light surface palette, and monospace code font (`JetBrains Mono` / system monospace).
  - Configure typed `apiClient` (`NEXT_PUBLIC_API_BASE_URL`) and TanStack Query provider.
- [ ] **5.2 Silent Render Cold-Start Warm-Up (`BackendWarmupPing`):**
  - Mount a non-blocking client component on `/`, `/login`, and `/register` that fires `GET /api/health` in the background as soon as a visitor arrives so Render wakes up early.
- [ ] **5.3 Public Landing Page (`/`):**
  - Hero section with **"Continue as Guest (1-Click)"**, **"Start Practicing"**, and **"Sign In"** actions.
  - Visual showcase of Module 1 & Module 2 (all 10 topics), the 4 Question Styles (Definitions, CLI Commands, Comparisons, Scenarios), the 1,000+ Question Bank + Dual-AI engine, and Timed Mock Exams.
- [ ] **5.4 Zero-Friction Auth Screens (`/login`, `/register`):**
  - Clean Sign-In and Sign-Up forms (Name, Email, Password — instant redirect to `/dashboard`) plus a prominent **1-Click "Continue as Guest"** button.
- [ ] **5.5 Learner Shell & Dashboard (`/dashboard`):**
  - Responsive sidebar (desktop) and bottom/drawer navigation (mobile) with links to Dashboard, Practice, Mock Exams, Saved Bookmarks, History, Progress, and Admin (when authorized).
  - Dashboard cards for quick-starting **Topic Practice**, **Module 1 Mock Exam**, **Module 2 Mock Exam**, and **Comprehensive Final Mock Exam**, plus recent sessions, topic/style accuracy overview, and weak-topic recommendations.

**Sprint 5 Exit Criteria:** Visitors can open `/`, trigger the silent `/api/health` ping, click "Continue as Guest" or register in seconds, and land on a responsive, data-driven `/dashboard`.

---

### Sprint 6 — Practice Mode, Mock Exam Presets, Code Blocks, Results & Bookmarks UI
**Goal:** Build all interactive learning, examination, code rendering, results review, bookmarking, analytics, and admin screens.

#### Tasks
- [ ] **6.1 Monospace Code & CLI Formatter (`FormattedCodeText`):**
  - Render inline `` `code` `` and fenced multi-line ```code blocks``` (Bash scripts, Linux commands, Git commands, JSON IAM policies) with high-contrast terminal styling and mobile scroll protection.
- [ ] **6.2 Practice Configuration (`/practice`) & Active Practice (`/practice/session/[sessionId]`):**
  - Topic selector grouped by **Module 1** and **Module 2**.
  - **Question Style Filter:** `All Styles (Mixed)`, `Definitions & Core Concepts`, `Commands & Syntax`, `Comparisons & Architecture`, `Scenario & Troubleshooting`.
  - Question count, difficulty selector, feedback timing toggle (`Immediate` vs. `End of Session`), and optional *"Generate Fresh AI Questions"* toggle.
  - Active question card displaying question style badge, formatted stem, shuffled options (`A–D`), Flag button, **Bookmark button**, question navigator grid, and immediate/deferred explanation panel.
- [ ] **6.3 Mock Exam Presets (`/exams`, `/exams/[examId]`) & Active Timed Mock Exam (`/exams/session/[sessionId]`):**
  - Dedicated preset cards for **Module 1 Mock Exam** (Topics 1–4), **Module 2 Mock Exam** (Topics 5–10), and **Comprehensive Final Mock Exam** (All 10 Topics).
  - Authoritative server-deadline countdown timer (`ExamTimer`) with low-time warnings and auto-submit on expiry.
  - Full state restoration across browser refreshes and double-submit protection.
- [ ] **6.4 Results & Learning Feedback (`/results/[sessionId]`):**
  - Score summary, completion time, **Topic Performance breakdown**, and **Question Style Performance breakdown**.
  - Filterable question-by-question review (`All`, `Incorrect`, `Flagged`, `Bookmarked`) showing selected option, correct option, full explanation, per-distractor rationales, **Course Module/Week Citation Badge** (e.g., `Module 1, Week 2 — Linux Fundamentals`), and **Bookmark toggle**.
- [ ] **6.5 Saved Bookmarks (`/bookmarks`), History (`/history`), Progress (`/progress`) & Admin (`/admin/*`):**
  - `/bookmarks`: Filter bookmarked questions by topic/style, expand explanations and course citations, or launch a bookmark revision quiz.
  - `/history`: Past practice and mock exam sessions with links to full review.
  - `/progress`: Visual charts and accessible tables for Module 1 & 2 topic mastery and question-style accuracy.
  - `/admin/questions`: Browse/filter the 1,000+ bank, inspect source (`SEED_BANK`, `AI_GEMINI`, `AI_GROQ`), edit/approve questions, and trigger batch Gemini/Groq generation.

**Sprint 6 Exit Criteria:** Complete end-to-end learner and admin journeys work smoothly in the browser across desktop and mobile viewports.

---

### Sprint 7 — Verification Against the 7 Core Questions, Keep-Alive Cron & Deployment
**Goal:** Prove that every Definition-of-Done question answers **"YES"** through automated tests, add the scheduled Supabase anti-pause keep-alive workflow, and prepare production deployment for Vercel, Render (Golang + Gin), and Supabase.

#### Tasks
- [ ] **7.1 Verification Suite for the 7 Core Questions:**
  1. **Curriculum & 1,000+ Multi-Style Bank Check:** Automated Go test verifying `>= 1,000` seeded questions across all 10 topics and all 4 styles (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`).
  2. **Anti-Memorization & Option Shuffling Check:** Test confirming returning sessions serve unseen questions and shuffle option order (`A–D`).
  3. **Dual-AI Failover & Zero-Downtime Fallback Check:** Test confirming Gemini -> Groq failover, bank persistence, and `<200ms` session creation when both AI providers are offline.
  4. **Mock Exam Integrity Check:** Test confirming active mock exam payloads omit answer keys/explanations, survive page refresh, and handle duplicate submit requests idempotently.
  5. **Actionable Feedback & Bookmarking Check:** Test confirming explanations, distractor rationales, course module/week citations, code blocks, and `/bookmarks` work end-to-end.
  6. **Zero-Friction Access Check:** Test confirming 1-click Guest Mode and instant Name/Email/Password sign-up.
  7. **Free-Tier Resilience, Supabase Anti-Pause & Responsive Check:** Verify `/api/health` warm-up ping executes a lightweight Supabase PostgreSQL query (`SELECT 1` / cached count), verify mobile/desktop responsiveness, and verify strict server-side secret isolation.
- [ ] **7.2 Scheduled Keep-Alive Cron & Deployment Configuration:**
  - Add `.github/workflows/keep-alive.yml` scheduled workflow to ping `GET /api/health` every 2–3 days so Supabase PostgreSQL never pauses due to 7 days of inactivity, even during weeks with zero human traffic.
  - Provide `.env.example` files and Render/Vercel config for `backend/` (Golang + Gin + Supabase + Gemini + Groq) and `frontend/` (Next.js on Vercel).
  - Verify production builds (`go build ./...`, `go test ./...`, and `npm run build`).

**Sprint 7 Exit Criteria:** All 7 Core Verification Questions pass automated and manual checks; both `backend/` (Golang + Gin) and `frontend/` (Next.js) build cleanly with zero errors.


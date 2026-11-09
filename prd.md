# Product Requirements Document (PRD) — CloudPrep

**Status:** Draft for review  
**Product:** CloudPrep  
**Audience:** Cloud Engineering learners preparing for multiple-choice examinations

## 1. Product vision

Build a seamless, useful exam-preparation environment where learners practice Cloud Engineering through a balanced variety of multiple-choice questions strictly grounded in their course materials (`school notes/`). Rather than limiting practice to a single style, the platform tests learners across **definitions and core concepts**, **commands and syntax** (Linux, Bash, Git), **concept comparisons**, and **realistic architecture/troubleshooting scenarios** (covering design, security, and cost).

The experience should not feel like repeatedly receiving the same static quiz or memorizing option letters. A learner should receive varied questions and dynamically shuffled answer options when they return to practice, backed by a **1,000+ pre-seeded question bank** combined with dual-provider AI generation (**Google Gemini** and **Groq**), while maintaining strict curriculum alignment, consistent difficulty, and trustworthy explanations.

## 2. Problem statement

Traditional practice tools often repeat a small fixed set of questions in the same option order or rely on only one narrow question format. This can reward positional memorization rather than genuine mastery and can make returning practice sessions feel stale. Conversely, purely live-AI quizzes can suffer from slow loading spinners, provider rate limits, or ungrounded off-syllabus questions.

CloudPrep addresses this by combining:
- A structured 10-topic curriculum mapped directly to the 8 course slide decks in `school notes/` (Module 1 & Module 2).
- Multiple question styles: **Definitions & Recall**, **Commands & Syntax**, **Comparisons & Distinctions**, and **Scenarios & Troubleshooting**.
- A **1,000+ pre-seeded Question Bank** (~100+ questions per topic) that serves both as a primary/hybrid pool and as a zero-downtime fallback when AI is unavailable.
- Dual-provider AI-assisted generation (**Google Gemini** + **Groq** with automatic failover) of new, curriculum-grounded questions that continually expand the question bank.
- Question selection that considers the learner's exposure history and dynamically **shuffles option order (`A–D`)** per session.
- Quality checks that reject ambiguous, incorrect, duplicated, or off-syllabus questions.
- Detailed feedback with course module/week citations, code-block formatting, question bookmarking, and progress tracking.

## 3. Goals and success measures

### Goals
- Provide topic-based practice (with optional **Question Style filtering**), mixed practice, mistake review, bookmarked question revision, and timed mock examinations (**Module 1**, **Module 2**, and **Comprehensive Final** presets).
- Cover all ten curriculum areas strictly within the jurisdiction of the provided course materials.
- Maintain a **1,000+ question bank** (~100+ questions per topic across all question styles) and combine it seamlessly with dual-AI generation (**Gemini** + **Groq**).
- Start sessions rapidly (<200ms from the unseen bank when available) while replenishing or mixing AI-generated questions without blocking the learner if AI providers fail.
- Shuffle answer options per session so learners cannot rely on memorizing option letters.
- Explain answers clearly (including monospace code/terminal blocks and exact course module/week references) and support review of mistakes and bookmarked items.
- Track topic-level and style-level performance and question exposure.

### Initial success measures
- Learners can complete a practice session and review explanations.
- The same learner sees materially different questions across sessions where the question supply allows it.
- Generated questions pass automated validation and are not shown when validation fails.
- Learners can identify weaker topics from their results.
- Exam submissions and scoring remain consistent if a learner refreshes or reconnects.

These measures should be converted into measurable targets after a pilot.

## 4. Target users and roles

### Learner
- Practices by topic, question style (definitions, commands/syntax, comparisons, scenarios, or mixed), difficulty, or mixed curriculum.
- Takes timed mock examinations (Module 1, Module 2, or Comprehensive Final presets).
- Reviews answers, explanations, course module/slide citations, flagged items, bookmarked questions, and past results.
- Views progress and topics needing further study.

### Content administrator
- Reviews, edits, approves, rejects, or retires question content across the 1,000+ question bank and AI-generated items.
- Manages curriculum references, question styles, tags, and difficulty.
- Inspects dual-provider AI generation (Gemini / Groq) and validation outcomes, and can trigger bulk question bank expansion.

The MVP may use a limited administrator interface. Administrative access must be restricted and auditable.

## 5. Curriculum scope and question styles

The curriculum is structured into two modules across ten core areas, strictly bounded by the 8 course slide decks in `school notes/`:

### Module 1 — Cloud & DevOps Foundations
1. **Introduction to Computing** (`M1_W1_Introduction to Cloud Computing.pdf`)
2. **Linux Fundamentals** (`M1_W2_Linux Fundamentals.pdf`)
3. **Bash Scripting & Automation** (`M1_ W4_Bash Scripting & Automation.pdf`)
4. **Git & Version Control** (`M1_W3_Git & Version Control - Google Slides.pdf`)

### Module 2 — Cloud Architecture, Services & AI
5. **Core Compute & IAM** (`M2_W1_Core Compute & IAM.pdf`)
6. **Cloud Networking** (`M2_W2_Networking & Security.pdf`)
7. **Cloud Storage** (`M2_W2_Networking & Security.pdf`)
8. **Databases** (`M2_W3_ Databases, Serverless & Application Services.pdf`)
9. **Serverless & Application Services** (`M2_W3_ Databases, Serverless & Application Services.pdf`)
10. **Cloud AI & MLOps** (`M2_W4_ Cloud AI & MLOps Fundamentals.pdf`)

Cloud concepts include compute instances, identity and access management (IAM users, groups, roles, policies), VPCs, subnets, route tables, internet gateways, NAT, security groups, network ACLs, bastion access, object/block/file storage, relational and NoSQL databases, serverless functions, API management, messaging, AI services, and MLOps lifecycle. Every question—whether pre-seeded in the bank or generated by AI—must stay strictly within the jurisdiction of these course materials.

### Supported question styles
Questions must not be limited to scenario-based formats. Every topic in the 1,000+ question bank and AI generator must include a balanced mix of four question styles:
1. **Definition & Concept Recall (~30%):** Direct questions testing terminology, definitions, cloud models, and core concepts from the slides.
2. **Command, Syntax & Service Identification (~25%):** CLI commands, flags, file permissions, script snippets, and output prediction (especially for Linux, Bash, Git, and IAM policies), as well as identifying the right cloud service for a stated function.
3. **Concept Comparison & Distinction (~20%):** Contrasting two or more related concepts taught in the notes (e.g., Security Groups vs. NACLs, IaaS vs. PaaS vs. SaaS, `git merge` vs. `git rebase`, SQL vs. NoSQL).
4. **Scenario, Architecture & Troubleshooting (~25%):** Practical problem-solving, failure diagnosis, security hardening, and cost/architecture trade-offs grounded in the curriculum.

## 6. Functional requirements

### 6.1 Authentication, 1-click guest mode, and learner profile
- **Zero-friction registration & sign-in:** Learners can register with just **Name, Email, and Password** (no email-verification wall required) and are immediately signed into their dashboard upon registration.
- **1-Click "Continue as Guest / Demo Learner" mode:** Both the Landing Page (`/`) and Sign-In Page (`/login`) provide a 1-click **Continue as Guest** button that provisions or signs into a guest learner session immediately, allowing evaluators, instructors, or learners to test practice quizzes and mock exams without typing credentials.
- Guest learners can optionally upgrade/claim their guest session into a permanent registered account later.
- Learners can sign out and access their own practice history, exposure history, and bookmarked questions across devices when signed into a registered account.
- A learner's results, bookmarks, preferences, and exposure history are private.
- Role-based access distinguishes learners (`LEARNER`, `GUEST`) and administrators (`ADMIN`).

### 6.2 Practice modes and mock exam presets
- **Topic practice:** select one or more curriculum topics and optionally filter by **Question Style** (`All Styles`, `Definitions & Concepts`, `Commands & Syntax`, `Comparisons & Architecture`, or `Scenario & Troubleshooting`).
- **Mixed practice:** receive a balanced set across selected topics or the whole curriculum.
- **Review mistakes:** practice concepts and questions associated with previously incorrect answers.
- **Saved / Bookmarked revision:** review and practice questions the learner has bookmarked for revision.
- **Timed mock exam presets:** complete a defined question count within a time limit across three built-in presets (always testing a balanced mix of all question styles):
  - **Module 1 Mock Exam:** Topics 1–4 (Intro to Computing, Linux, Bash Scripting, Git).
  - **Module 2 Mock Exam:** Topics 5–10 (Compute & IAM, Networking, Storage, Databases, Serverless, Cloud AI & MLOps).
  - **Comprehensive Final Mock Exam:** All 10 curriculum topics combined.
- Practice settings include question count, question style filter, difficulty range, feedback timing, and an optional *"Include / Generate Fresh AI Questions"* preference.

### 6.3 Question interaction
- Present one question at a time with a clearly navigable question list.
- Render **monospace code blocks and inline terminal/command syntax** cleanly in question stems, options, and explanations (for Linux commands, Bash scripts, Git CLI, and JSON/YAML policies).
- **Dynamic option order shuffling (`A–D`):** Shuffle the order of options for each session so learners cannot rely on memorizing option positions (`A`, `B`, `C`, `D`) when revisiting a question.
- Allow answer selection, flagging for in-session review, **bookmarking for long-term revision**, and navigation.
- In mock-exam mode, conceal correctness and explanations until submission.
- In learning/practice mode, feedback timing may be immediate or deferred based on session settings.
- Preserve session progress across refreshes and temporary connection loss.
- Submit once; repeated submission attempts must not create duplicate results.

### 6.4 Results and learning feedback
- Show total score, topic-level performance, and breakdown by question style.
- Show the selected answer, correct answer, full explanation, and rationale for why each distractor is wrong.
- Display the **exact course module/week/slide reference** (e.g., *Module 1, Week 2 — Linux File Permissions*) on every reviewed question.
- Allow learners to **bookmark/save** any question or explanation from the results view into their personal *"Saved for Revision"* tab.
- Allow filtering review by all, incorrect, flagged, or bookmarked questions.
- Record question exposure and outcomes for future selection.
- Provide a useful next-step suggestion based on weaker topics without overstating diagnostic certainty.

### 6.5 Question bank and content management (1,000+ pre-seeded questions)
- **1,000+ Pre-Seeded Question Bank:** Pre-populate the PostgreSQL database with **at least 1,000 curated, approved questions** (~100+ questions per topic across all 10 curriculum areas), spanning definitions, commands/syntax, comparisons, and scenarios directly from the 8 PDFs in `school notes/`.
- Store approved questions with topic, subtopic/objective, **question style** (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`), difficulty, cognitive skill, options, answer key, explanations, distractor rationales, and course module/slide source references.
- **Hybrid Bank + AI Growth:** Validated AI-generated questions from Gemini or Groq are saved to the database so the question bank continuously grows beyond the initial 1,000+ questions.
- **Zero-Downtime Fallback:** Whenever AI generation is disabled, quota-limited, or unreachable, the 1,000+ Question Bank serves sessions immediately with zero interruption.
- Administrators can review AI-generated drafts, edit, version, approve, retire, and audit question history.
- A question's source (`SEED_BANK`, `AI_GEMINI`, `AI_GROQ`, `ADMIN`) and validation state must be visible to administrators.

### 6.6 Dual-provider AI-assisted question generation (Google Gemini + Groq)
AI is used to create varied question candidates across all question styles (definitions, commands/syntax, comparisons, and scenarios)—strictly within the jurisdiction of the supplied course materials.

**Dual-Provider Architecture (`Gemini` + `Groq`):**
- Both **Google Gemini (`gemini-2.5-flash`)** and **Groq (`llama-3.3-70b-versatile`)** are wired behind the `QuestionGenerationProvider` interface.
- **Automatic Failover & Load Distribution:** If the primary provider (e.g., Gemini) hits a rate limit, times out, or returns an error, the orchestrator automatically fails over to the secondary provider (e.g., Groq) before falling back to the 1,000+ curated Question Bank. The backend can also alternate/load-balance between Gemini and Groq to maximize free-tier throughput.

**Instant Session Start + Smart Background / Hybrid Generation:**
- Because the database contains **1,000+ pre-seeded questions**, learners should not wait on a slow AI loading spinner on every practice start:
  1. **Instant Start (<200ms):** When sufficient unseen questions exist in the 1,000+ bank for the learner's selected topic/style, the session can start immediately from the bank (or combine unseen bank questions with fresh AI questions).
  2. **Background Replenishment & On-Demand AI:** As learners practice—or when a learner's unseen pool for a topic drops below a threshold, or when they explicitly request fresh AI questions—the backend invokes Gemini/Groq to generate, validate, and persist new questions into the bank.

For an AI generation request, the system will:
1. Read the target topic, question style (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`), difficulty, and count.
2. Retrieve approved curriculum excerpts from the `school notes/` knowledge base and related question patterns.
3. Inspect existing fingerprints and learner exposure history to avoid duplicates.
4. Call the primary AI provider (**Gemini** or **Groq**, with automatic failover to the other) using strict JSON schema enforcement.
5. Validate schema, curriculum alignment (no off-syllabus facts), single best answer, distractor rationales, and similarity against the existing bank.
6. Reject invalid candidates and retry within a bounded budget.
7. Persist accepted questions into the PostgreSQL Question Bank with provenance metadata (`AI_GEMINI` or `AI_GROQ`) and serve them.

**Variety and anti-memorization requirements**
- Avoid re-serving the same question to the same learner when unseen alternatives exist in the 1,000+ bank or via AI generation.
- Dynamically **shuffle option order (`A–D`)** on every session so learners must read the option text rather than memorizing option letters.
- Track both exact question IDs and normalized-content fingerprints.
- Use similarity checks to detect near-duplicates, not only identical text.
- Prefer unseen questions and under-exposed concepts, while keeping topic, question-style, and difficulty balance.
- If the learner exhausts eligible unseen questions and AI generation is unavailable, offer a clearly labeled refreshed set with shuffled options rather than silently claiming every item is brand new.

**Quality and safety requirements**
- Strictly ground all questions in the 8 approved course PDFs in `school notes/`.
- Require exactly one best answer for standard single-select questions.
- Check that distractors are plausible but demonstrably incorrect according to the course notes.
- Reject contradictory, off-syllabus, or ambiguous content.
- Do not expose answer keys before submission in mock exams.
- Keep model prompts, API keys (`GEMINI_API_KEY`, `GROQ_API_KEY`), and private learner history server-side.
- Store provider/model version and validation results for traceability.
- Apply request limits and token/cost budgets to AI generation.

### 6.6.1 Question supply modes
- **1,000+ Pre-seeded Curated Question Bank:** 100+ questions per topic across all 10 topics and all 4 question styles, providing instant (<200ms) session creation and 100% reliable fallback.
- **Hybrid Bank + Dual-AI Generation (Gemini & Groq):** mixes unseen bank questions with fresh AI-generated candidates and saves validated AI questions back into the bank.
- **Variant & Style-Targeted Generation:** generates a new definition, command/syntax challenge, comparison, or scenario around a specific learning objective from `school notes/`.

### 6.7 Progress, bookmarks, and analytics
- Store practice sessions, mock exams, responses, flags, **bookmarked questions**, scores, question-style accuracy, and topic outcomes.
- Display recent activity, topic performance, and a dedicated **Saved / Bookmarked Questions** revision view.
- Use exposure history to inform question selection.
- Avoid ranking learners publicly in the MVP.

## 7. Non-functional requirements

- **Security:** server-side authorization, secure session handling, input validation, rate limits, secret management, and audit logging for privileged actions.
- **Privacy:** collect only data needed for account, learning history, and product operation.
- **Reliability:** persist active exam state, make submission idempotent, fail over between Gemini and Groq automatically, and fall back cleanly to the 1,000+ database question bank.
- **Performance, Cold-Start & Anti-Pause Mitigation:** routine question retrieval from the 1,000+ bank should complete in `<200ms` (powered by a compiled **Golang + Gin** backend); AI generation must have bounded timeouts; the frontend landing (`/`) and login (`/login`) pages must silently fire a background `GET /api/health` wake-up ping to warm up the Render free-tier API, and `GET /api/health` must execute a lightweight Supabase PostgreSQL query (paired with a scheduled keep-alive cron job) so Supabase never pauses the database after 7 days of inactivity.
- **Scalability:** start with a Golang modular monolith (using native goroutines for background AI bank replenishment) and scale stateless application instances as needed.
- **Maintainability:** typed API contracts, automated tests, migrations, clear module boundaries, and documented decisions.
- **Accessibility:** keyboard navigation, readable contrast, monospace code blocks for commands/scripts, labels, and responsive layouts.
- **Cost control:** dual free-tier AI providers (Gemini + Groq), token budgets, generation quotas, automatic persistence/reuse of validated AI items in the database bank, and usage monitoring.

## 8. Proposed technology stack and free-tier deployment

The initial stack prioritizes accessible free-tier services and a simple deployment workflow. Free-tier quotas, sleep behavior, storage limits, and provider policies can change; verify them before launch. “Free tier” does not mean unlimited or guaranteed zero-cost usage.

| Layer | Proposed technology | Deployment / purpose |
|---|---|---|
| Web application | Next.js + TypeScript | Vercel; learner and administrator interface |
| Backend API | **Golang (`Go 1.22+`) + Gin Framework (`gin-gonic/gin`)** | Render; ultra-fast REST API (`~20 MB` RAM), 1,000+ bank selection, option shuffling, and AI orchestration |
| Database | Supabase PostgreSQL | Managed PostgreSQL storing users, sessions, and the 1,000+ question bank (~5 MB) |
| ORM / DB Driver & Migrations | **GORM + `pgx/v5`** (with Go auto-migrations & seed CLI) | Type-safe Go database access, schema migrations, and 1,000+ question seeding |
| AI integration | Dual-provider Go adapter wiring **Google Gemini (`gemini-2.5-flash`)** and **Groq (`llama-3.3-70b-versatile`)** with automatic failover | Server-side only; never expose API keys in the browser |
| Background work | **Native Go Goroutines** with bounded worker channels | Replenish question bank asynchronously without blocking learner session start or requiring Redis |
| Similarity search | PostgreSQL normalized fingerprint + full-text trigram checks initially; evaluate pgvector when needed | Reduce repeated questions without adding another service |
| Version control / CI & Keep-Alive | GitHub + scheduled GitHub Actions keep-alive cron | Source control, automated checks, and periodic `/api/health` ping so Supabase never pauses |
| Local development | Docker Compose (optional) | Consistent local Go API/database tooling |
| Monitoring | Provider logs plus `/api/health` DB heartbeat endpoint | Observe API health, cold starts, Supabase DB connectivity, latency, and Gemini/Groq usage |

### Deployment responsibilities

- **Vercel:** hosts the Next.js frontend. The browser calls the Render Go API using a configured public API base URL and triggers a silent `GET /api/health` warm-up ping on public entry pages (`/`, `/login`).
- **Render:** hosts the compiled **Golang + Gin** API. It handles authentication, authorization, practice sessions, question selection from the 1,000+ bank, option shuffling, scoring, and server-side calls to Google Gemini and Groq.
- **Supabase:** provides the PostgreSQL database holding curriculum excerpts, the 1,000+ pre-seeded questions, growing AI-generated questions, and learner sessions.
- **GORM + `pgx/v5` (Go):** manages schema migrations and the Go seed command (`cmd/seed`) that populates the 1,000+ curriculum-grounded question bank.

### Free-tier operational considerations

- **Render cold-start warm-up (Go advantage):** A free Render web service spins down after 15 minutes of inactivity, but a compiled **Golang + Gin** binary uses only ~15–25 MB of RAM and boots in milliseconds once the container starts. The Next.js frontend silently fires `GET /api/health` when a user opens `/` or `/login` so Render wakes up while the user is browsing or typing credentials.
- **Preventing Supabase 7-day inactivity pause:** Supabase pauses free-tier databases after 7 consecutive days of zero database queries. To prevent this:
  1. `GET /api/health` executes a lightweight PostgreSQL query (`SELECT 1` / cached question count check) against Supabase so every visitor wake-up ping also registers real database activity.
  2. A scheduled **GitHub Actions cron workflow** (`.github/workflows/keep-alive.yml`, or Vercel Cron) pings `GET /api/health` automatically every 2–3 days so Supabase stays active even during quiet periods with zero human visitors.
- Keep database connection limits and pool sizes in mind. Configure `sql.DB` connection pooling (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`) appropriately for Supabase.
- Protect database credentials. The Supabase service-role key and database connection string must never be exposed to client-side code.
- Set Vercel environment variables for the public API URL; set Render environment variables for database URLs, auth secrets, Gemini & Groq API credentials, allowed frontend origins, and environment mode.
- Configure Gin CORS middleware to allow only the deployed frontend origin(s) and local development origin(s).
- Monitor database size (1,000+ questions uses only ~3–5 MB of Supabase's 500 MB free limit), egress, API usage, build minutes, and Gemini/Groq quotas.
- Put all AI calls behind quotas, timeouts (`context.WithTimeout`), token limits, Gemini <-> Groq automatic failover, and instant fallback to the 1,000+ curated question bank.
- Keep development and production credentials separate. Avoid committing `.env` files.

### Environment variables (illustrative names)

**Vercel**
- `NEXT_PUBLIC_API_BASE_URL`

**Render**
- `DATABASE_URL`
- `DIRECT_URL` (only if required by the selected Supabase/Prisma connection setup)
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `FRONTEND_ORIGIN`
- `AI_PRIMARY_PROVIDER` (`gemini` or `groq`, defaults to `gemini`)
- `AI_FALLBACK_PROVIDER` (`groq` or `gemini`, defaults to `groq`)
- `GEMINI_API_KEY`
- `GEMINI_MODEL_ID` (e.g., `gemini-2.5-flash`)
- `GROQ_API_KEY`
- `GROQ_MODEL_ID` (e.g., `llama-3.3-70b-versatile`)

Use the exact connection-string format and pooling configuration recommended by Supabase and Prisma for the deployment environment. Never put secrets in `NEXT_PUBLIC_*` variables.

### Dual AI provider strategy (Google Gemini + Groq)

The application defines an internal `QuestionGenerationProvider` interface with two active adapters wired out of the box: **Google Gemini (`gemini-2.5-flash`)** and **Groq (`llama-3.3-70b-versatile`)**.
- Both providers offer generous free-tier API allowances, fast inference, and structured JSON output enforcement.
- If the primary provider hits a rate limit or network error, the orchestrator automatically fails over to the secondary provider.
- If both providers are exhausted, unconfigured, or unavailable, the system seamlessly serves eligible questions from the **1,000+ pre-seeded Question Bank**.

## 9. MVP scope

### Included
- Learner authentication and roles.
- Curriculum/topic browsing across Module 1 (Topics 1–4) and Module 2 (Topics 5–10).
- Multi-style questions: **Definitions & Recall**, **Commands & Syntax**, **Comparisons & Distinctions**, and **Scenarios & Troubleshooting**—all strictly within the `school notes/` course materials.
- **1,000+ pre-seeded Question Bank** (~100+ questions per topic) stored in PostgreSQL.
- **Dual AI providers (Google Gemini + Groq)** wired with automatic failover, validation, bank persistence, and zero-downtime fallback to the 1,000+ bank.
- Topic practice (with Question Style filter), mixed practice, mistake review, and bookmarked question revision.
- Timed mock exams with **Module 1**, **Module 2**, and **Comprehensive Final** presets, navigation, flags, and persisted progress.
- Dynamic **option order shuffling (`A–D`)** per session.
- Monospace code block and inline command rendering for Linux, Bash, Git, and cloud policies.
- Results, explanations, distractor rationales, exact course module/week citations, and question **bookmarking**.
- Silent `GET /api/health` frontend ping on landing/login pages to wake Render free-tier instances early.
- Basic admin review and bank-generation tools.
- Topic and question-style analytics.

### Deferred
- Social leaderboards and public profiles.
- Payments/subscriptions.
- Mobile native applications.
- Adaptive learning plans beyond topic and style recommendations.
- Complex organization/cohort administration.
- High-stakes proctoring.

## 10. Key user stories

- As a learner, I want to practice a selected topic and filter by question style (definitions, commands/syntax, comparisons, or scenarios) so I can target specific weaknesses.
- As a learner, I want sessions to start quickly from a large 1,000+ question bank while also getting fresh AI-generated questions from Gemini/Groq so I never run out of variety.
- As a learner, I want answer options shuffled each session so I truly understand the answer instead of memorizing option letters.
- As a learner, I want Linux commands, Bash scripts, and Git syntax formatted in clean monospace code blocks so they are easy to read.
- As a learner, I want explanations for every answer, rationales for wrong options, and the exact course module/week reference so I know which slide deck to review.
- As a learner, I want to bookmark tricky questions and explanations so I can revise them before my exam.
- As a learner, I want dedicated timed mock exams for Module 1, Module 2, and the full course that withhold answers until submission.
- As a learner, I want my session to survive a refresh so I do not lose progress.
- As an administrator, I want to manage the 1,000+ question bank and review Gemini/Groq-generated questions and their course sources.

## 11. Acceptance criteria & definition-of-done verification questions

- The database is seeded with **1,000+ approved questions** (~100+ per topic across all 10 curriculum areas) covering definitions, commands/syntax, comparisons, and scenarios strictly from `school notes/`.
- Both **Google Gemini** and **Groq** are wired into the backend generation service with automatic failover between them and graceful fallback to the 1,000+ question bank.
- A learner can start, resume, submit, and review a practice session (with optional question style filtering) or a timed mock exam (Module 1, Module 2, or Comprehensive Final).
- Answer option order (`A–D`) is dynamically shuffled per session and scored accurately against the session's mapped option order.
- Mock mode does not reveal correctness or explanations before final submission.
- Each question is mapped to an approved curriculum topic, question style, and course module/week source reference, and includes a full explanation and distractor rationales.
- Question stems and options render inline code and multi-line code blocks cleanly.
- Learners can bookmark questions and review their saved bookmarks.
- Visitors can enter the dashboard immediately via **1-click "Continue as Guest"** or register with just **Name, Email, and Password** (no email verification wall).
- The landing (`/`) and login (`/login`) pages silently ping `GET /api/health` to warm up the Render backend.
- AI candidates failing validation or straying outside the course curriculum are not served.
- A duplicate submission does not create a second exam result.
- Unauthorized users cannot access another learner's session or administrator functions.

### 11.1 Seven core questions the built project must answer "YES" to

1. **Does it test the actual course curriculum across all question styles (not just scenarios)?**
   - All 10 topics (Module 1 & Module 2) grounded strictly in `school notes/`, testing **Definitions & Recall**, **CLI Commands & Syntax**, **Comparisons**, and **Scenarios**, backed by **1,000+ pre-seeded questions** in Supabase PostgreSQL and filterable by style in Practice Mode.
2. **Does it prevent rote memorization when a learner practices repeatedly?**
   - Unseen-first question selection across returning sessions plus **dynamic per-session option shuffling (`A–D`)** so learners cannot memorize option letters.
3. **Does the Dual-AI engine (Google Gemini + Groq) work seamlessly with a 100% reliable fallback?**
   - Automatic failover between **Gemini (`gemini-2.5-flash`)** and **Groq (`llama-3.3-70b-versatile`)**, automatic persistence of valid AI questions to grow the bank, and `<200ms` zero-downtime fallback to the 1,000+ Question Bank if both AI providers are offline or out of quota.
4. **Do the Timed Mock Exams uphold real exam integrity?**
   - Dedicated **Module 1**, **Module 2**, and **Comprehensive Final** mock exams with authoritative server deadlines, zero answer-key leakage in network payloads prior to submission, full state recovery across page refreshes, and idempotent submission.
5. **Does the feedback actually help a student learn and revise?**
   - Correct answer explanation, per-distractor rationales, **exact course module/week references**, clean **monospace code blocks** for Linux/Bash/Git/IAM snippets, **question bookmarking (`/bookmarks`)**, and topic/style progress analytics.
6. **Can anyone start using it in seconds without friction?**
   - **1-Click "Continue as Guest"** on the Landing and Login pages plus instant Name/Email/Password registration without an email verification wall.
7. **Is it cleanly deployed, secure, and responsive on free-tier hosting?**
   - Silent `GET /api/health` Render cold-start warm-up on `/` and `/login`, responsive desktop & mobile layouts without horizontal overflow, and strict server-side secret isolation.

## 12. Open decisions for review

- Exact default question counts and time limits for the Module 1, Module 2, and Comprehensive Final mock exams (e.g., 30 questions / 45 mins for Module 1 & 2; 60 questions / 90 mins for Comprehensive Final).
- Whether practice mode feedback defaults to immediate (after each question) or deferred (at end of set), while allowing learners to toggle it.
- Whether validated Gemini/Groq questions generated during on-demand sessions are auto-approved into the shared 1,000+ bank immediately upon passing automated validation or queued as `DRAFT` for admin approval.
- Initial pilot audience and admin account provisioning.

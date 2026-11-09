# CloudPrep — Frontend Design & Implementation Specification

**Status:** Draft for architecture review  
**Application:** CloudPrep  
**Frontend deployment:** Vercel  
**Frontend stack:** Next.js App Router, TypeScript, Tailwind CSS

---

## 1. Purpose

This document defines the frontend structure, user experience, page behaviour, reusable components, state management, accessibility expectations, and integration contracts for CloudPrep.

The frontend should make it straightforward for learners to practise Cloud Engineering concepts, sit timed mock examinations, understand their results, and track their progress. The experience should feel consistent across desktop, tablet, and mobile devices.

This document complements:
- `prd.md` — product goals, scope, requirements, and acceptance criteria.
- `architecturaldesign.md` — application modules and high-level architecture.
- `systemdesign.md` — data model, API outline, exam lifecycle, and AI question-generation pipeline.

It does not replace the backend API contract. Request and response schemas must be agreed with the backend before implementation.

## 2. Frontend goals

1. Provide a clear learner journey from dashboard to practice or mock exam and then results.
2. Make question reading, answering, flagging, and navigation easy.
3. Preserve learner progress during temporary network interruptions.
4. Prevent answer keys and explanations from appearing during an active mock exam.
5. Communicate AI-generation and API loading/failure states honestly.
6. Use reusable components and consistent visual rules.
7. Keep the frontend maintainable and avoid unnecessary client-side state.

## 3. Proposed frontend stack

| Concern | Technology | Responsibility |
|---|---|---|
| Framework | Next.js App Router | Routing, layouts, rendering, frontend application |
| Language | TypeScript | Static typing and maintainability |
| Styling | Tailwind CSS | Responsive styling and design tokens |
| Component primitives | shadcn/ui + Radix UI | Accessible dialogs, menus, tabs, dropdowns, tooltips and other primitives |
| Server/API state | TanStack Query | Fetching, caching, mutations, retries and invalidation |
| Forms | React Hook Form | Form state and submission |
| Client validation | Zod | Validate user input and, where practical, API payload shapes |
| Local/global UI state | React state; Zustand only if justified | Temporary UI state shared across distant components |
| Icons | Lucide React | Consistent interface icons |
| Charts | Recharts | Topic performance and progress visualizations |
| Unit/component tests | Vitest + React Testing Library | Component and frontend logic tests |
| End-to-end tests | Playwright | Critical learner journeys and browser behaviour |

Dependencies are proposals, not a requirement to install everything immediately. Add each library only when its use is justified. Backend validation remains authoritative; frontend validation is for usability, not security.

## 4. User roles and navigation

### Learner navigation
- Dashboard
- Practice
- Mock exams (Module 1, Module 2, and Comprehensive Final presets)
- Saved / Bookmarks (for exam revision)
- History
- Progress
- Profile

### Administrator navigation
- Admin overview
- Question bank (1,000+ pre-seeded + Gemini/Groq-generated items)
- AI-generated drafts / review queue & bulk generator
- Curriculum topics
- Generation and validation records

The UI may hide links a user cannot access, but this is not authorization. The backend must enforce every permission independently.

## 5. Route map

| Route | Access | Purpose |
|---|---|---|
| `/` | Public | Product landing page (triggers silent `GET /api/health` warm-up ping) |
| `/login` | Public | Sign in (triggers silent `GET /api/health` warm-up ping) |
| `/register` | Public | Create learner account (triggers silent `GET /api/health` warm-up ping) |
| `/dashboard` | Learner | Overview, recent activity, suggested practice, quick mock exam links |
| `/practice` | Learner | Configure topic, mixed, or question-style-filtered practice |
| `/practice/session/[sessionId]` | Session owner | Active practice session (shuffled options, code blocks, bookmarking) |
| `/exams` | Learner | Browse built-in Mock Exam presets (Module 1, Module 2, Comprehensive Final) |
| `/exams/[examId]` | Learner | Exam details and instructions |
| `/exams/session/[sessionId]` | Session owner | Active timed mock exam |
| `/results/[sessionId]` | Session owner | Score, topic/style breakdown, course citations, bookmarking, and answer review |
| `/bookmarks` | Learner | Saved / bookmarked questions & explanations for revision |
| `/history` | Learner | Previous sessions |
| `/progress` | Learner | Topic and question-style performance and learning trends |
| `/profile` | Learner | Account and preferences |
| `/admin` | Administrator | Content operations overview & 1,000+ bank metrics |
| `/admin/questions` | Administrator | Search and manage 1,000+ question bank & trigger Gemini/Groq expansion |
| `/admin/questions/review` | Administrator | Review AI-generated candidates |
| `/admin/questions/[questionId]` | Administrator | View/edit question and provenance |
| `/admin/curriculum` | Administrator | Manage topics and objectives |

Unauthenticated users visiting protected routes should be redirected to sign-in with a safe return destination. Authorization must be verified by the API and, where appropriate, by server-side route/layout checks.

## 6. Main learner journeys

### 6.1 First-time learner (Zero-friction Sign Up or 1-Click Guest Mode)
1. Open landing page (frontend silently pings `GET /api/health` in the background to wake the Render free-tier API).
2. Either:
   - Click **"Continue as Guest"** for instant 1-click entry into the dashboard, **or**
   - Register with **Name, Email, and Password** (no email verification gate; immediately signed in).
3. Arrive at dashboard.
4. Choose a topic, question style, or mixed practice.
5. Complete a session (served rapidly from the 1,000+ Question Bank with dynamically shuffled options).
6. Review score, explanations, distractor rationales, course module/week citations, and bookmark any tricky questions.

### 6.2 Topic practice (with Question Style filter)
1. Open Practice.
2. Choose one or more topics ( across Module 1 or Module 2).
3. Select **Question Style** filter:
   - *All Styles (Mixed)* — default
   - *Definitions & Core Concepts*
   - *Commands & Syntax* (Linux, Bash, Git, cloud syntax)
   - *Comparisons & Architecture*
   - *Scenario & Troubleshooting*
4. Select question count, optional difficulty, and feedback timing.
5. Start session (instant `<200ms` load from the 1,000+ bank when unseen items exist, or hybrid with Gemini/Groq).
6. Answer, navigate, flag for in-session review, or **bookmark** for long-term revision.
7. Submit or finish the set.
8. Review explanations, course module/week references, and suggested topics to revisit.

### 6.3 Timed mock exam (Module 1, Module 2, and Comprehensive Final presets)
1. Open Mock Exams (`/exams`).
2. Select one of the three built-in presets:
   - **Module 1 Mock Exam:** Intro to Computing, Linux Fundamentals, Bash Scripting, Git & Version Control.
   - **Module 2 Mock Exam:** Core Compute & IAM, Cloud Networking, Cloud Storage, Databases, Serverless & Application Services, Cloud AI & MLOps.
   - **Comprehensive Final Mock Exam:** All 10 curriculum topics combined.
3. Read instructions and start the exam (always serves a balanced mix of definitions, commands/syntax, comparisons, and scenarios with shuffled option order `A–D`).
4. Answer questions and navigate between them.
5. See remaining time and unanswered/flagged counts.
6. Submit manually or allow the server-defined deadline policy to close the session.
7. View results, explanations, and course slide references only after the server confirms submission.

During the active mock exam, the frontend must not receive or display answer keys, correctness flags, or explanations.

### 6.4 Review mistakes & saved bookmarks
1. Open History, Progress, or the **Saved / Bookmarks (`/bookmarks`)** page.
2. Filter saved bookmarks or missed questions by curriculum topic or question style.
3. Review the question, monospace code snippets, full explanation, distractor rationales, and exact course module/week reference, or launch a targeted revision practice session.

## 7. Page-level requirements

### 7.1 Landing page (`/`)
**Purpose:** Explain what CloudPrep does, warm up the Render backend silently, and guide learners to sign in or register.

Suggested sections:
- Header with logo and navigation.
- Hero statement highlighting comprehensive Cloud Engineering preparation across definitions, CLI commands, comparisons, and real-world scenarios—backed by a **1,000+ question bank** and dual-AI variety.
- Short explanation of Practice Mode (with Question Style filters) and Timed Mock Exam presets (Module 1, Module 2, and Final).
- Curriculum overview showing Module 1 (Topics 1–4) and Module 2 (Topics 5–10).
- Feature section: multi-style questions, 1,000+ bank + AI hybrid, shuffled options, code-block formatting, detailed explanations with course module citations, and progress tracking.
- Call to action.
- Footer with relevant product links.
- **Silent cold-start warm-up:** On mount, fire a non-blocking `GET /api/health` request so a sleeping Render free-tier service begins waking immediately.

Avoid unsupported claims such as guaranteed exam success or official affiliation unless confirmed.

### 7.2 Authentication (`/login`, `/register`)
- Fire a silent `GET /api/health` warm-up request on mount if not already triggered.
- **Zero-friction registration:** Require only **Name, Email, and Password**—no email verification gate. Immediately sign the learner in and redirect to `/dashboard`.
- **1-Click "Continue as Guest" button:** Prominent action on `/`, `/login`, and `/register` that calls `POST /api/auth/guest` and takes the visitor straight to `/dashboard` in one click.
- Provide clearly labelled fields and useful inline validation.
- Show password visibility control where appropriate.
- Display server errors in plain language without exposing internals.
- Prevent accidental repeated submissions while a request is pending.
- Support keyboard navigation and password-manager-friendly fields.
- Never store credentials in local storage.
- Follow the authentication strategy agreed with the backend. If cookie-based sessions are used, ensure the frontend request configuration supports credentials and CSRF requirements.

### 7.3 Dashboard (`/dashboard`)
Suggested content:
- Greeting and concise welcome message.
- Quick-start actions: **Start Topic Practice**, **Module 1 Mock Exam**, **Module 2 Mock Exam**, and **Comprehensive Final Mock Exam**.
- Recent practice sessions.
- Topic & question-style progress summary.
- Saved / Bookmarked questions counter and quick link.
- Topics to revisit based on recorded results.

States:
- Loading skeleton.
- First-use empty state with a clear first action.
- Partial-data state if one dashboard section fails.
- Error state with retry.

Do not imply that a learner's topic score is a definitive measure of ability; show attempt counts and context where useful.

### 7.4 Practice configuration (`/practice`)
Controls:
- Topic selection grouped by **Module 1 (Foundations)** and **Module 2 (Cloud Architecture & Services)** (select one, several, a full module, or all 10).
- **Question Style filter:**
  - `All Styles (Mixed)` — default
  - `Definitions & Core Concepts`
  - `Commands & Syntax`
  - `Comparisons & Architecture`
  - `Scenario & Troubleshooting`
- Question count, within server-supported limits.
- Difficulty preference (`Mixed`, `Foundational`, `Intermediate`, `Advanced`).
- Practice feedback timing (`Immediate after each question` or `Deferred until end of set`).
- Optional focus on previously missed objectives or bookmarked questions.
- Optional *"Include Fresh AI-Generated Questions (Gemini / Groq)"* toggle (while defaulting to instant `<200ms` selection from unseen items in the 1,000+ bank + background AI replenishment).

Behaviour:
- Explain that the platform selects unseen questions from the **1,000+ course-grounded Question Bank** and generates validated new candidates via Gemini/Groq where needed, with shuffled option order (`A–D`).
- On start, show a clear loading state while the API creates the session.
- If live AI generation is unavailable, seamlessly proceed using the 1,000+ approved Question Bank.

### 7.5 Active practice (`/practice/session/[sessionId]`)
Core layout:
- Session header with selected topics, question style badge (`Definition`, `Command & Syntax`, `Comparison`, or `Scenario`), and progress.
- Question number and total.
- Question stem with **clean monospace code blocks and inline terminal/command formatting** (e.g., Bash scripts, Linux commands, Git flags, JSON IAM policies).
- Answer options (`A–D`, rendered in the server-supplied shuffled order for this session) as accessible selectable controls supporting inline monospace code.
- Previous/Next navigation.
- Question navigator or compact progress list.
- **Flag for in-session review** button and **Bookmark for revision** button.
- Save/sync status.
- Submit/finish action.

Behaviour:
- Persist selected answers, flags, and bookmarks through the API.
- Debounce only where appropriate; do not delay critical answer persistence excessively.
- Restore current question, shuffled option order, answers, and flags from the server when reopening.
- In practice mode, feedback may be immediate or deferred based on session configuration.
- Do not derive correctness from client-side answer keys.

### 7.6 Mock exam selection & instructions (`/exams`, `/exams/[examId]`)
Show:
- Three featured Mock Exam presets (**Module 1 Mock Exam**, **Module 2 Mock Exam**, and **Comprehensive Final Mock Exam**) with their covered topics, question style balance, question count, and time limit.
- Submission and timeout rules.
- Confirmation that answers, explanations, and course slide citations are unlocked immediately after submission.
- Clear Start Exam action.

The API must supply authoritative exam settings. The frontend must not invent or alter time limits.

### 7.7 Active mock exam (`/exams/session/[sessionId]`)
Core layout:
- Exam title (e.g., *Module 1 Mock Exam*, *Module 2 Mock Exam*, or *Comprehensive Final Mock Exam*).
- Countdown based on the server-provided deadline.
- Question stem (with monospace code/command formatting where applicable) and shuffled answer options (`A–D`).
- Question navigation with answered/unanswered/flagged states.
- Previous/Next controls.
- Flag for review.
- Submit action and confirmation dialog.

Timer requirements:
- Treat the server deadline as authoritative.
- Calculate displayed remaining time using the server deadline and client clock, while accounting for clock drift where possible.
- Do not rely on a browser interval as the authority for exam expiry.
- On refresh or reconnect, fetch the session state and deadline again.
- At zero, submit or close according to the backend's defined policy.
- Prevent accidental double submission in the UI, while the API remains idempotent.

Security:
- The active mock-exam API response must omit correct answers, explanations, and correctness metadata.
- Never bundle answer keys in hidden DOM, JavaScript state, or client-side constants.

### 7.8 Results (`/results/[sessionId]`)
Show:
- Total score and number correct/incorrect/unanswered.
- Completion time where available.
- Topic-level performance and **question-style performance** (Definitions, Commands/Syntax, Comparisons, Scenarios).
- Question-by-question review filterable by All, Incorrect, Flagged, or Bookmarked.
- Correct answer, full explanation, and distractor rationales for why each wrong option is incorrect.
- **Exact course module/week/slide reference badge** (e.g., *Module 1, Week 2 — Linux Fundamentals: File Permissions*) on every question.
- **Bookmark / Save for Revision** toggle on every reviewed question.
- Suggested topics and question styles for further practice.

Only display review details after the backend confirms the session is submitted and the user's role/ownership permits access.

### 7.9 Saved / Bookmarked questions (`/bookmarks`)
- List all questions the learner has bookmarked from practice sessions or exam results.
- Filter by curriculum module/topic and question style (`Definition`, `Command & Syntax`, `Comparison`, `Scenario`).
- Allow expanding the full explanation, distractor rationales, and course module/week reference, or removing a bookmark.
- Allow starting a quick practice drill from bookmarked questions.

### 7.10 History (`/history`)
- List previous sessions with mode/preset, date, score, topic coverage, and status.
- Support pagination or incremental loading.
- Link to results only when available.
- Include empty, loading, and error states.

### 7.11 Progress (`/progress`)
- Topic-level performance chart across Module 1 and Module 2.
- Performance breakdown by **Question Style** (Definitions, Commands & Syntax, Comparisons, Scenarios).
- Attempts and accuracy by topic.
- Recent trend where enough data exists.
- Topics to revisit.
- Clear labels and accessible alternatives to charts.

Avoid misleading charts when there are too few attempts. Show “not enough activity yet” rather than fabricating a trend.

### 7.12 Profile (`/profile`)
- Display and edit only supported profile fields.
- Provide sign-out.
- Provide account/security options when supported by backend.
- Clearly explain any destructive account action and require confirmation.

### 7.13 Administration
Admin pages should support:
- Search/filter the **1,000+ Question Bank** by topic, question style (`DEFINITION`, `COMMAND_SYNTAX`, `COMPARISON`, `SCENARIO`), status, difficulty, and source (`SEED_BANK`, `AI_GEMINI`, `AI_GROQ`, `ADMIN`).
- Trigger on-demand batch AI question generation via **Google Gemini** or **Groq** for specific topics and question styles.
- Inspect question text (including code blocks), answer, explanations, distractor rationales, course module/week source references, and AI generation metadata.
- Edit, approve, reject, and retire questions according to backend permissions.
- Inspect validation failures and similarity warnings.
- Show confirmation before consequential actions.
- Provide audit-friendly feedback after changes.

AI-generated candidates that require human review should be visibly marked as drafts until the backend reports their approved status.

## 8. Component hierarchy

Illustrative structure:

```text
app/
  (public)/
    page.tsx
    login/page.tsx
    register/page.tsx
  (learner)/
    layout.tsx
    dashboard/page.tsx
    practice/page.tsx
    practice/session/[sessionId]/page.tsx
    exams/page.tsx
    exams/[examId]/page.tsx
    exams/session/[sessionId]/page.tsx
    results/[sessionId]/page.tsx
    bookmarks/page.tsx
    history/page.tsx
    progress/page.tsx
    profile/page.tsx
  admin/
    layout.tsx
    page.tsx
    questions/page.tsx
    questions/review/page.tsx
    questions/[questionId]/page.tsx
    curriculum/page.tsx
  api/ (only if a Next.js route handler is intentionally needed)

components/
  layout/
    app-header.tsx
    learner-sidebar.tsx
    admin-sidebar.tsx
    mobile-navigation.tsx
    backend-warmup-ping.tsx
  dashboard/
    recent-sessions.tsx
    topic-progress-card.tsx
    recommended-practice.tsx
  practice/
    practice-configuration-form.tsx
    question-style-selector.tsx
    question-card.tsx
    formatted-code-text.tsx
    answer-option.tsx
    question-navigator.tsx
    session-progress.tsx
    flag-question-button.tsx
    bookmark-question-button.tsx
    save-status.tsx
  exam/
    exam-preset-card.tsx
    exam-instructions.tsx
    exam-timer.tsx
    submit-confirmation-dialog.tsx
  results/
    score-summary.tsx
    topic-performance-chart.tsx
    style-performance-breakdown.tsx
    answer-review-card.tsx
    course-source-badge.tsx
  admin/
    question-table.tsx
    question-editor-form.tsx
    ai-batch-generator-dialog.tsx
    candidate-review-panel.tsx
    validation-status.tsx
  ui/
    ...

features/
  auth/
  curriculum/
  practice/
  exams/
  results/
  bookmarks/
  progress/
  admin/

lib/
  api-client.ts
  query-client.ts
  auth/
  validation/
  utils/

types/
  api.ts
  domain.ts
```

This is a suggested organization, not a mandate to create empty folders. Prefer feature-oriented code ownership and avoid duplicating components.

## 9. State management

### Server state
Use TanStack Query for:
- Current learner profile.
- Curriculum, topics, and mock exam presets.
- Session details (including per-session shuffled option order).
- Session responses and flags.
- Saved / bookmarked questions.
- Results and history.
- Progress summaries (by topic and by question style).
- Admin question bank lists and AI generation/review records.

Use query keys consistently, for example:
- `['me']`
- `['curriculum']`
- `['practice-session', sessionId]`
- `['session-results', sessionId]`
- `['bookmarks', filters]`
- `['session-history', filters]`
- `['progress-summary']`

Invalidate or update relevant queries after mutations. Do not keep a second, divergent copy of server data in Zustand.

### Local UI state
Use React state for:
- Current open dialog.
- Selected question index where it is only navigation state.
- Temporary unsaved form values.
- Expanded/collapsed panels.

Consider Zustand only if a real cross-component client-state requirement emerges. Do not add it by default.

### Session state
The backend is the source of truth for saved answers, shuffled option mapping, session status, and exam deadline. The frontend may optimistically update the selected option or bookmark toggle for responsiveness, but it must reconcile with the server and show a sync error if persistence fails.

## 10. API integration requirements

The frontend should call the Render-hosted **Golang + Gin** API through one typed API client.

Requirements:
- Base URL configured with `NEXT_PUBLIC_API_BASE_URL`.
- Silent background `GET /api/health` warm-up call fired on public entry pages (`/`, `/login`, `/register`) to wake up the Render free-tier Go service *and* trigger a lightweight database query on Supabase PostgreSQL (resetting Supabase's 7-day inactivity pause timer).
- HTTPS in production.
- Explicit CORS allowlist on the API.
- Consistent error response format.
- Request timeout and controlled retry policy.
- Credentials included if cookie-based authentication is selected.
- CSRF token handling if required by the chosen cookie/session design.
- No secret API keys (`GEMINI_API_KEY`, `GROQ_API_KEY`, DB credentials) in frontend environment variables.
- Runtime validation of important API payloads where practical.
- Abort obsolete requests when navigating away or changing filters.

Expected API capabilities are described in `systemdesign.md`. Final request/response types must be agreed with backend implementation; do not guess payload shapes in production code.

### Suggested response handling
- `400`: show field-level or request validation feedback.
- `401`: refresh or redirect to sign-in according to the authentication design.
- `403`: show an access-denied state.
- `404`: show a not-found state.
- `409`: explain a state conflict, such as a session already submitted.
- `429`: show a rate-limit message and respect retry guidance.
- `5xx` / network failure: show retry and preserve local unsynced input where safe.

## 11. 1,000+ Question Bank & Dual-AI Generation UX (Gemini + Groq)

AI generation (via **Google Gemini** and **Groq**) is a server-side process. The browser must never call Gemini or Groq directly.

When starting practice:
- **Instant Start (<200ms):** Because the backend has a **1,000+ pre-seeded Question Bank**, most practice and mock exam sessions start immediately from unseen bank questions while the backend replenishes questions via Gemini/Groq in the background.
- **When live AI generation is active:** Show a concise, non-alarming status such as *"Preparing your practice set..."*
- Do not expose raw prompts, provider errors, Gemini/Groq API keys, or internal validation details to learners.
- If Gemini fails over to Groq or falls back to the 1,000+ curated Question Bank, transition seamlessly without interrupting the learner.
- If fewer questions match a very narrow filter than requested, clearly state the actual number before the session begins.
- Do not imply that AI-generated content is guaranteed error-free.

## 12. Visual design system & code-block formatting

The exact visual identity should be confirmed before high-fidelity implementation. Initial design direction:
- Clean, focused learning interface.
- Strong readability and clear hierarchy.
- Neutral surfaces with one primary accent colour.
- **Monospace code & terminal formatting:** Question stems, options, and explanations must render inline code (e.g., `` `chmod 755` ``, `` `git checkout -b` ``) and multi-line fenced code blocks (e.g., Bash scripts `#!/bin/bash`, JSON IAM policies) using a clear monospace font family with high-contrast background styling and horizontal scroll protection on mobile.
- Distinct badges for **Question Style** (`Definition`, `Command & Syntax`, `Comparison`, `Scenario`) and **Course Source Reference** (e.g., `Module 1, Week 2 — Linux Fundamentals`).
- Avoid excessive gradients, animation, and decorative elements that distract during exams.
- Use a distinct but restrained visual state for answered, unanswered, flagged, bookmarked, and current questions.

### Design tokens to define
- Brand and semantic colours: primary, background, surface, text, muted, border, code-surface, success, warning, error.
- Typography scale: page title, section title, body, caption, question stem, and **monospace code font** (e.g., JetBrains Mono / Fira Code / system monospace).
- Spacing scale and container widths.
- Border radii, shadows, focus rings.
- Button sizes and variants.
- Form field heights and validation states.

Do not use colour alone to communicate answer or question status; pair colour with text, icons, or accessible labels.

## 13. Responsive behaviour

### Desktop
- Sidebar navigation.
- Main content with readable maximum width.
- Exam question area and question navigator can appear side by side when space allows.

### Tablet
- Collapsible navigation.
- Keep question content dominant.
- Navigator may become a drawer or collapsible panel.

### Mobile
- Compact header and bottom navigation or menu.
- Single-column question layout.
- Code blocks scroll cleanly inside their container without causing page-level horizontal overflow.
- Large touch targets.
- Question navigator available through a clearly labelled control.
- Timer and submit action remain visible without covering question content.
- Dialogs fit within viewport and support scrolling.

Test at narrow widths, including 320px, and ensure no horizontal scrolling for core flows.

## 14. Accessibility requirements

- Use semantic headings, landmarks, buttons, and form labels.
- Use semantic `<code>` and `<pre><code>` elements for CLI commands and scripts.
- Ensure all interactions are keyboard accessible.
- Provide visible focus indicators.
- Use accessible names for icon-only buttons (such as Flag and Bookmark toggles).
- Ensure sufficient text and control contrast.
- Announce important timer warnings and submission status without excessively interrupting screen-reader users.
- Do not rely on colour alone for question states.
- Support zoom and responsive text.
- Provide text summaries or tables alongside charts.

## 15. Loading, empty, and error states

Every data-driven page must define:
- Initial loading state (including friendly feedback if the Render free-tier instance is completing its cold-start wake-up).
- Empty state.
- Recoverable error state with retry.
- Permission-denied state where relevant.
- Partial-data state when independent sections can load separately.

For practice/exam sessions, distinguish:
- Saving.
- Saved.
- Save failed / retry required.
- Session expired.
- Session submitted.
- Reconnecting.

Do not discard a learner's selected answer just because a request failed. Preserve it temporarily and clearly indicate that it has not yet been confirmed by the server.

## 16. Frontend security requirements

- Treat all browser input as untrusted.
- Do not rely on hidden navigation or client-side role checks for authorization.
- Do not store passwords, refresh tokens, database credentials, or Gemini/Groq AI secrets in local storage.
- Never include answer keys in active mock-exam payloads or client bundles.
- Render question content and code blocks safely; never use unescaped `dangerouslySetInnerHTML`.
- Apply a restrictive Content Security Policy where compatible with the deployment.
- Avoid unsafe HTML injection.
- Use secure production origins and HTTPS.
- Do not log passwords, tokens, answer keys, or sensitive learner data.
- Ensure logout clears relevant client caches and session state.

## 17. Testing and quality gates

### Unit/component tests
- Practice configuration validation (including Question Style filter and topic selection).
- Question navigator state.
- Monospace code block and inline command rendering (`formatted-code-text.tsx`).
- Shuffled answer option selection, flag, and bookmark interactions.
- Timer display calculations using a fixed server deadline.
- Error and save-status components.
- Role-specific navigation visibility.

### Integration tests
- API client authentication, error handling, and silent `/api/health` warm-up ping.
- Query invalidation after answer save, bookmark toggle, and submission.
- Session resume from server state (preserving shuffled option order).
- Handling of 1,000+ bank instant start and Gemini/Groq failover/fallback states.

### End-to-end tests
1. Open landing/login page (triggering `/api/health`), register/sign in, and reach dashboard.
2. Start topic practice filtered by Question Style (e.g., Commands & Syntax or Definitions), answer questions, bookmark a question, submit, and review course module citations.
3. Open `/bookmarks` and verify the saved question and explanation appear.
4. Start a Module 1, Module 2, or Comprehensive Final mock exam and verify answers/explanations are omitted before submission.
5. Refresh an active session and confirm progress and shuffled option order are restored.
6. Submit twice and confirm the UI displays one final result.
7. Test expired session behaviour.
8. Confirm learner cannot open admin pages or another learner's results.
9. Test the core journeys on desktop and mobile viewport sizes.

### Quality gates before merge
- TypeScript type-check passes.
- Lint passes.
- Unit/component tests pass.
- Critical end-to-end tests pass.
- No secrets committed.
- No mock-exam answer keys shipped to the browser.
- Responsive and keyboard checks completed for changed screens.

## 18. Vercel deployment

- Connect the frontend repository to Vercel.
- Set the production `NEXT_PUBLIC_API_BASE_URL` to the deployed Render API origin.
- Configure preview environment variables separately from production.
- Configure the Render API CORS allowlist for the Vercel production domain and approved preview/local origins.
- Confirm cookie domain, `Secure`, `SameSite`, and CSRF behaviour across the frontend and API domains before relying on cookie authentication.
- Ensure no server-only secret (`GEMINI_API_KEY`, `GROQ_API_KEY`, `DATABASE_URL`, JWT secrets) is prefixed with `NEXT_PUBLIC_`.
- Test preview deployments against a non-production database/environment.
- Verify build, type-check, and tests before production promotion.

Free-tier platform limits and policies may change. The UI should tolerate API cold starts (mitigated by the silent `/api/health` ping) and temporary service unavailability.

## 19. Implementation sequence

1. Confirm visual direction, design tokens (including monospace code typography), and navigation.
2. Initialize Next.js App Router and TypeScript project.
3. Add Tailwind and accessible component primitives.
4. Build public layout, silent `/api/health` warm-up component, and authentication screens.
5. Build learner shell, dashboard, and Module 1 / Module 2 curriculum views.
6. Build practice configuration (with Question Style filter) and active practice flow (with shuffled options, monospace code rendering, and bookmarking).
7. Build Module 1, Module 2, and Final mock exam presets, instructions, timer, navigation, and submission flow.
8. Build results (with course module/week citations and style breakdown), bookmarks (`/bookmarks`), history, and progress pages.
9. Build administrator 1,000+ question bank explorer, review queue, and Gemini/Groq batch generator screens.
10. Add responsive/accessibility refinements and automated tests.
11. Configure Vercel environments and production integration with Render.

## 20. Screen acceptance checklist

- [ ] Every route has an agreed access level.
- [ ] Landing (`/`) and authentication (`/login`, `/register`) pages silently ping `GET /api/health` to wake Render early.
- [ ] Every data-driven screen has loading, empty, error, and retry states.
- [ ] Forms have labels, client validation, and server error handling.
- [ ] Practice configuration supports filtering by Question Style (`All`, `Definitions`, `Commands & Syntax`, `Comparisons`, `Scenarios`).
- [ ] Mock Exams page offers Module 1, Module 2, and Comprehensive Final presets.
- [ ] Question stems, options, and explanations render inline code and multi-line code blocks cleanly.
- [ ] Answer options (`A–D`) render in the session's shuffled order and persist across refreshes.
- [ ] Practice answers, flags, and bookmarks are persisted and recoverable.
- [ ] Mock exam never receives answer keys before submission.
- [ ] Timer is based on the server deadline.
- [ ] Submission handles retries without duplicate results.
- [ ] Results show exact course module/week citations, distractor rationales, and bookmark toggles to the authorized session owner.
- [ ] Saved / Bookmarks page (`/bookmarks`) allows learners to revise bookmarked questions and explanations.
- [ ] Dual-AI (Gemini/Groq) generation status and 1,000+ bank fallback work seamlessly without blocking learners.
- [ ] Admin actions (including 1,000+ bank management and Gemini/Groq batch generation) are protected by backend authorization.
- [ ] Desktop, tablet, and mobile layouts are tested.
- [ ] Keyboard navigation and visible focus are tested.
- [ ] No secrets or sensitive data are exposed in the client bundle.
- [ ] Vercel environment variables and Render CORS settings are verified.

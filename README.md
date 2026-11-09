# CloudPrep

CloudPrep is a comprehensive Cloud Engineering exam-preparation platform. It helps learners prepare for multiple-choice assessments through a rich mix of question styles—including **definitions & concept recall**, **command & syntax questions** (Linux, Bash, Git), **concept comparisons**, and **realistic architecture/troubleshooting scenarios**—strictly grounded in the provided course curriculum (`school notes/`).

## Documentation

- [`prd.md`](prd.md) — product requirements, scope, question styles, 1,000+ question bank, dual-AI strategy (Gemini + Groq), and user stories.
- [`architecturaldesign.md`](architecturaldesign.md) — application architecture, modules, multi-provider AI failover, and key design decisions.
- [`systemdesign.md`](systemdesign.md) — core flows, data model, APIs, option shuffling, AI generation pipeline, reliability, and security design.
- [`frontenddesign.md`](frontenddesign.md) — Next.js App Router UI/UX specification, code-block rendering, mock exam presets, bookmarking, and Render cold-start warm-up.
- [`sprints/DEVELOPMENT_SPRINTS.md`](sprints/DEVELOPMENT_SPRINTS.md) — step-by-step 7-sprint execution roadmap mapped to the Definition of Done.

## Curriculum coverage

The curriculum is organized across **Module 1 (Foundations)** and **Module 2 (Cloud Architecture & Services)**, grounded in the 8 course slide decks in `school notes/`:

### Module 1 — Cloud & DevOps Foundations
1. Introduction to Computing (`M1_W1_Introduction to Cloud Computing.pdf`)
2. Linux Fundamentals (`M1_W2_Linux Fundamentals.pdf`)
3. Bash Scripting & Automation (`M1_ W4_Bash Scripting & Automation.pdf`)
4. Git & Version Control (`M1_W3_Git & Version Control - Google Slides.pdf`)

### Module 2 — Cloud Architecture, Services & AI
5. Core Compute & IAM (`M2_W1_Core Compute & IAM.pdf`)
6. Cloud Networking (`M2_W2_Networking & Security.pdf`)
7. Cloud Storage (`M2_W2_Networking & Security.pdf`)
8. Databases (`M2_W3_ Databases, Serverless & Application Services.pdf`)
9. Serverless & Application Services (`M2_W3_ Databases, Serverless & Application Services.pdf`)
10. Cloud AI & MLOps (`M2_W4_ Cloud AI & MLOps Fundamentals.pdf`)

Questions span multiple styles and organizational contexts strictly within the scope of the course materials:
- **Definitions & Core Concepts** (~30%) — foundational terminology, cloud models, and service definitions.
- **Commands, Syntax & Service Identification** (~25%) — Linux CLI, Bash scripts, Git commands, and cloud service selection.
- **Comparisons & Distinctions** (~20%) — contrasting related concepts (e.g., Security Groups vs. NACLs, IaaS vs. PaaS vs. SaaS, SQL vs. NoSQL).
- **Scenarios & Troubleshooting** (~25%) — practical design, security, cost, and troubleshooting problems across startup, growing business, and enterprise contexts.

## Product principles

- **Multi-style curriculum mastery:** Test definitions, CLI commands/syntax, conceptual comparisons, and practical scenarios—strictly within the jurisdiction of the provided course notes.
- **1,000+ Curated Question Bank + Dual-AI Hybrid (Gemini & Groq):** Start with a pre-seeded bank of **1,000+ questions** (~100+ per curriculum topic). Combine bank questions with AI-generated questions from **Google Gemini** and **Groq** (wired with automatic failover), save validated AI questions to grow the bank, and use the 1,000+ bank as a zero-downtime fallback.
- **Instant session start & smart anti-memorization:** Serve unseen bank questions immediately (<200ms) while replenishing via AI in the background or on demand, and dynamically **shuffle option order (`A–D`)** per session so learners cannot memorize option positions.
- **Module 1, Module 2, and Full Final Mock Exam presets:** Offer dedicated mock exams for Module 1, Module 2, and the full 10-topic curriculum.
- **Rich learning feedback:** Explain why the correct answer is right, why each distractor is wrong, cite the exact course module/week reference, support monospace code/terminal blocks, and allow learners to **bookmark/save** questions for revision.
- **Zero-friction authentication + 1-click Guest Mode:** Provide lightweight Name/Email/Password sign-up (no email verification wall—instant login upon registration) alongside a **1-click "Continue as Guest"** button on the Landing and Login pages so anyone can jump straight into practicing or evaluating the platform immediately.
- **High-performance Golang (Gin) modular monolith:** Built with **Golang + Gin framework + GORM (`pgx/v5`)** for ultra-fast cold starts (`<100ms`), minimal memory usage (`~20 MB` RAM on Render's 512 MB free tier), and native Go goroutines for background AI question replenishment.
- **Render & Supabase anti-sleep / anti-pause keep-alive:** The `GET /api/health` endpoint both wakes the Render Go API when a visitor opens `/` or `/login` *and* executes a lightweight database query on Supabase PostgreSQL, paired with a scheduled GitHub Actions / Vercel Cron keep-alive ping so Supabase never pauses the database after 7 days of inactivity.

## Definition of Done (7 Core Verification Questions)

To count as successfully built, the live platform must answer **"YES"** to these 7 questions:
1. **Multi-style curriculum coverage:** Does it test all 10 topics strictly from `school notes/` across **Definitions**, **CLI Commands & Syntax**, **Comparisons**, and **Scenarios**, backed by **1,000+ pre-seeded questions** in Supabase PostgreSQL?
2. **Anti-memorization:** Does it serve unseen questions across returning sessions and dynamically **shuffle answer options (`A–D`)** every session?
3. **Dual-AI + 100% fallback:** Do **Google Gemini** and **Groq** generate and persist valid questions with automatic failover, while sessions still start in `<200ms` from the 1,000+ bank even if both AI providers are offline?
4. **Mock exam integrity:** Do the **Module 1**, **Module 2**, and **Comprehensive Final** timed mock exams conceal answer keys in network payloads until submission, survive browser refreshes, and prevent duplicate submissions?
5. **Actionable learning feedback:** Does every reviewed question show the correct explanation, why each distractor is wrong, the **exact course module/week citation**, clean **monospace code blocks**, and **bookmarking** for revision?
6. **Zero-friction access:** Can anyone start in 1 second via **1-Click "Continue as Guest"** or instant Name/Email/Password sign-up (no email verification wall)?
7. **Free-tier resilience & security:** Do `/` and `/login` silently ping `GET /api/health` (waking Render and touching Supabase PostgreSQL so it never pauses), is the UI responsive on desktop and mobile, and are all secrets kept strictly server-side?

## Status

This is a planning and design baseline. The current deployment target is **Vercel** (Next.js TypeScript frontend), **Render** (**Golang + Gin** backend API), and **Supabase PostgreSQL** (database), with **Google Gemini** and **Groq** wired for AI question generation. Free-tier limits and provider policies can change, so confirm current limits before deployment.

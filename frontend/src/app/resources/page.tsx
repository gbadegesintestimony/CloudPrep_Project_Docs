"use client";

import React, { useState } from "react";
import Link from "next/link";
import {
  BookOpen,
  FileText,
  Download,
  ExternalLink,
  Search,
  Eye,
  X,
  Layers,
  Sparkles,
  ArrowRight,
  ChevronRight,
  ShieldCheck,
} from "lucide-react";
import { LearnerNav } from "@/components/LearnerNav";

interface CourseResource {
  id: string;
  title: string;
  weekCode: string;
  moduleNumber: 1 | 2;
  moduleLabel: string;
  pdfPath: string;
  fileSize: string;
  pageCount: string;
  description: string;
  keyTopics: string[];
  practiceTopicId: string;
}

const COURSE_RESOURCES: CourseResource[] = [
  {
    id: "m1-w1",
    title: "Introduction to Cloud Computing & AWS Core Services",
    weekCode: "Week 1",
    moduleNumber: 1,
    moduleLabel: "Module 1 • Systems & Tooling",
    pdfPath: "/school-notes/m1-w1-intro-cloud.pdf",
    fileSize: "324 KB",
    pageCount: "25 Slides",
    description:
      "Core cloud deployment models (IaaS, PaaS, SaaS), the AWS Shared Responsibility Model, Global Infrastructure (Regions, AZs, Edge Locations), and fundamental cost-optimization principles.",
    keyTopics: [
      "IaaS vs PaaS vs SaaS",
      "Shared Responsibility Model",
      "AWS Regions & Availability Zones",
      "Fault Tolerance & High Availability",
    ],
    practiceTopicId: "topic-1-cloud-intro",
  },
  {
    id: "m1-w2",
    title: "Linux Fundamentals & System Administration",
    weekCode: "Week 2",
    moduleNumber: 1,
    moduleLabel: "Module 1 • Systems & Tooling",
    pdfPath: "/school-notes/m1-w2-linux-fundamentals.pdf",
    fileSize: "544 KB",
    pageCount: "38 Slides",
    description:
      "Essential command-line proficiency: Linux Filesystem Hierarchy (FHS), permissions (chmod, chown, octal/symbolic), user management, process inspection (ps aux, top, kill), and package administration.",
    keyTopics: [
      "File System Hierarchy (FHS)",
      "Permissions (chmod 750 / 644)",
      "Process Inspection & Signals (SIGTERM)",
      "User & Group Administration",
    ],
    practiceTopicId: "topic-2-linux-fundamentals",
  },
  {
    id: "m1-w3",
    title: "Git & Version Control for Cloud Engineers",
    weekCode: "Week 3",
    moduleNumber: 1,
    moduleLabel: "Module 1 • Systems & Tooling",
    pdfPath: "/school-notes/m1-w3-git-version-control.pdf",
    fileSize: "889 KB",
    pageCount: "42 Slides",
    description:
      "Distributed version control workflows: three-tree architecture (working directory, index, repository), branch strategies, conflict resolution, rebase vs merge, and team collaboration protocols.",
    keyTopics: [
      "Three-Tree Architecture",
      "Fast-Forward vs 3-Way Merge",
      "Git Rebase & History Linearization",
      "Merge Conflict Resolution",
    ],
    practiceTopicId: "topic-3-git-version-control",
  },
  {
    id: "m1-w4",
    title: "Bash Scripting & Automation Pipelines",
    weekCode: "Week 4",
    moduleNumber: 1,
    moduleLabel: "Module 1 • Systems & Tooling",
    pdfPath: "/school-notes/m1-w4-bash-automation.pdf",
    fileSize: "1.02 MB",
    pageCount: "46 Slides",
    description:
      "Production shell scripting: Shebang execution environments, parameters ($#, $@, $?), conditionals, loop constructs, functions, I/O redirection, and robust error management (set -e, trap).",
    keyTopics: [
      "Positional Parameters & Return Codes",
      "Conditionals & Test Operators ([ -f ])",
      "Loops, Arrays & Functions",
      "Error Handling (set -euo pipefail)",
    ],
    practiceTopicId: "topic-4-bash-automation",
  },
  {
    id: "m2-w1",
    title: "Core Compute & AWS Identity and Access Management (IAM)",
    weekCode: "Week 6",
    moduleNumber: 2,
    moduleLabel: "Module 2 • AWS Architecture & Services",
    pdfPath: "/school-notes/m2-w1-compute-iam.pdf",
    fileSize: "130 KB",
    pageCount: "28 Slides",
    description:
      "AWS EC2 compute lifecycle and purchasing strategies (On-Demand, Spot, Reserved), EBS storage types, Security Groups vs NACLs, IAM Users, Roles, Policies, and Least Privilege Architecture.",
    keyTopics: [
      "EC2 Pricing & Instance Lifecycles",
      "Security Groups (Stateful) vs NACLs",
      "IAM Roles, Policies & Instance Profiles",
      "Principle of Least Privilege",
    ],
    practiceTopicId: "topic-6-compute-iam",
  },
  {
    id: "m2-w2",
    title: "Cloud Networking, VPCs & Infrastructure Security",
    weekCode: "Week 7",
    moduleNumber: 2,
    moduleLabel: "Module 2 • AWS Architecture & Services",
    pdfPath: "/school-notes/m2-w2-networking-security.pdf",
    fileSize: "176 KB",
    pageCount: "32 Slides",
    description:
      "Virtual Private Cloud (VPC) topology, public and private subnet design, Internet Gateways, NAT Gateways, Route Tables, CIDR block calculations, and defense-in-depth security.",
    keyTopics: [
      "VPC Topology & CIDR Subnetting",
      "Internet Gateways vs NAT Gateways",
      "Route Tables & Traffic Flow",
      "VPC Peering & Private Endpoints",
    ],
    practiceTopicId: "topic-7-networking-security",
  },
  {
    id: "m2-w3",
    title: "Databases, Serverless & Application Services",
    weekCode: "Week 8",
    moduleNumber: 2,
    moduleLabel: "Module 2 • AWS Architecture & Services",
    pdfPath: "/school-notes/m2-w3-databases-serverless.pdf",
    fileSize: "300 KB",
    pageCount: "36 Slides",
    description:
      "Managed database architectures (RDS Multi-AZ, Read Replicas, Aurora, DynamoDB NoSQL), serverless compute with AWS Lambda, API Gateway integration, and decoupling with SQS and SNS.",
    keyTopics: [
      "RDS Multi-AZ vs Read Replicas",
      "Amazon DynamoDB Partition Keys",
      "AWS Lambda Serverless Event Triggers",
      "SQS Queuing & SNS Pub/Sub Fan-out",
    ],
    practiceTopicId: "topic-8-databases-serverless",
  },
  {
    id: "m2-w4",
    title: "Cloud AI & MLOps Infrastructure",
    weekCode: "Week 9",
    moduleNumber: 2,
    moduleLabel: "Module 2 • AWS Architecture & Services",
    pdfPath: "/school-notes/m2-w4-ai-mlops.pdf",
    fileSize: "200 KB",
    pageCount: "30 Slides",
    description:
      "Enterprise machine learning workloads in the cloud: pipeline architecture, training vs real-time inference compute tradeoffs, GPU acceleration, model registry, and production monitoring.",
    keyTopics: [
      "ML Training vs Inference Compute",
      "GPU Acceleration & Cost Management",
      "Model Registry & Deployment Patterns",
      "Monitoring Drift & Pipeline Automation",
    ],
    practiceTopicId: "topic-10-cloud-ai-mlops",
  },
];

export default function ResourcesPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedModule, setSelectedModule] = useState<"ALL" | 1 | 2>("ALL");
  const [activePdf, setActivePdf] = useState<CourseResource | null>(null);

  const filteredResources = COURSE_RESOURCES.filter((res) => {
    if (selectedModule !== "ALL" && res.moduleNumber !== selectedModule) {
      return false;
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      const matchTitle = res.title.toLowerCase().includes(q);
      const matchDesc = res.description.toLowerCase().includes(q);
      const matchTopic = res.keyTopics.some((t) => t.toLowerCase().includes(q));
      const matchWeek = res.weekCode.toLowerCase().includes(q);
      return matchTitle || matchDesc || matchTopic || matchWeek;
    }
    return true;
  });

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20">
      <LearnerNav />

      <main className="pt-20 px-4 sm:px-6 lg:px-8 py-8 space-y-8 max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="space-y-2 border-b border-sky-100 pb-6">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold uppercase tracking-wider text-sky-600">
              Curriculum Study Materials
            </span>
            <span className="inline-flex items-center gap-1 rounded-full border border-sky-200 bg-sky-50 px-2.5 py-0.5 text-[10px] font-bold text-sky-800">
              <ShieldCheck className="h-3 w-3 text-sky-600" /> Official Class Notes &amp; Slides
            </span>
          </div>
          <h1 className="font-display text-3xl sm:text-4xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3">
            <BookOpen className="h-8 w-8 text-sky-600" />
            Course Slide Decks &amp; Lecture Notes
          </h1>
          <p className="text-sm text-slate-600 max-w-3xl leading-relaxed">
            Review the official course lecture slide decks for all 10 curriculum topics across Module 1 and Module 2.
            Read directly inside the interactive browser reader, download for offline study, or jump directly into practice.
          </p>
        </div>

        {/* Filter & Search Bar */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
          {/* Module Filter Tabs */}
          <div className="flex items-center gap-1.5 rounded-xl border border-sky-100 bg-white p-1 shadow-sm">
            <button
              type="button"
              onClick={() => setSelectedModule("ALL")}
              className={`rounded-lg px-3.5 py-1.5 text-xs font-bold transition-all ${
                selectedModule === "ALL"
                  ? "bg-sky-600 text-white shadow-sm"
                  : "text-slate-600 hover:text-slate-900 hover:bg-slate-50"
              }`}
            >
              All Modules ({COURSE_RESOURCES.length})
            </button>
            <button
              type="button"
              onClick={() => setSelectedModule(1)}
              className={`rounded-lg px-3.5 py-1.5 text-xs font-bold transition-all ${
                selectedModule === 1
                  ? "bg-sky-600 text-white shadow-sm"
                  : "text-slate-600 hover:text-slate-900 hover:bg-slate-50"
              }`}
            >
              Module 1 (Weeks 1–5)
            </button>
            <button
              type="button"
              onClick={() => setSelectedModule(2)}
              className={`rounded-lg px-3.5 py-1.5 text-xs font-bold transition-all ${
                selectedModule === 2
                  ? "bg-sky-600 text-white shadow-sm"
                  : "text-slate-600 hover:text-slate-900 hover:bg-slate-50"
              }`}
            >
              Module 2 (Weeks 6–10)
            </button>
          </div>

          {/* Search Input */}
          <div className="relative sm:w-80">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search topics (e.g. Linux, IAM, Bash)..."
              className="w-full rounded-xl border border-slate-200 bg-white pl-9 pr-4 py-2 text-xs text-slate-900 placeholder-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-500/20 shadow-xs"
            />
            <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-slate-400" />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery("")}
                className="absolute right-3 top-2 text-xs text-slate-400 hover:text-slate-600"
              >
                Clear
              </button>
            )}
          </div>
        </div>

        {/* Resources Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {filteredResources.map((res) => (
            <div
              key={res.id}
              className="flex flex-col justify-between rounded-2xl border border-sky-100 bg-white/95 p-6 shadow-lg shadow-sky-100/50 backdrop-blur-xl transition-all hover:border-sky-300 hover:shadow-xl hover:shadow-sky-200/50 group"
            >
              <div className="space-y-4">
                {/* Meta Header */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="rounded-lg bg-sky-50 px-2.5 py-1 font-mono text-[11px] font-bold text-sky-700 border border-sky-200">
                      {res.weekCode}
                    </span>
                    <span className="text-[11px] font-semibold text-slate-500">
                      {res.moduleLabel}
                    </span>
                  </div>
                  <span className="text-[11px] font-mono text-slate-400 font-medium">
                    {res.fileSize}
                  </span>
                </div>

                {/* Title & Description */}
                <div>
                  <h3 className="font-display text-lg font-bold text-slate-900 group-hover:text-sky-700 transition-colors">
                    {res.title}
                  </h3>
                  <p className="mt-2 text-xs text-slate-600 leading-relaxed font-normal">
                    {res.description}
                  </p>
                </div>

                {/* Core Lecture Points */}
                <div className="space-y-1.5 pt-1">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                    Key Exam Concepts Covered
                  </span>
                  <div className="flex flex-wrap gap-1.5">
                    {res.keyTopics.map((topic, i) => (
                      <span
                        key={i}
                        className="rounded-md bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-700 border border-slate-200/80"
                      >
                        {topic}
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="mt-6 flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 pt-4">
                <button
                  type="button"
                  onClick={() => setActivePdf(res)}
                  className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-4 py-2 text-xs font-bold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/20 transition-all"
                >
                  <Eye className="h-3.5 w-3.5" />
                  <span>Read Slides In-App</span>
                </button>

                <div className="flex items-center gap-2">
                  <a
                    href={res.pdfPath}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[11px] font-semibold text-slate-700 hover:bg-slate-50 hover:text-slate-900 transition-colors shadow-xs"
                    title="Open PDF in a new browser tab"
                  >
                    <ExternalLink className="h-3 w-3 text-slate-400" />
                    <span>Open Tab</span>
                  </a>

                  <a
                    href={res.pdfPath}
                    download
                    className="inline-flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[11px] font-semibold text-slate-700 hover:bg-slate-50 hover:text-slate-900 transition-colors shadow-xs"
                    title="Download slide deck"
                  >
                    <Download className="h-3 w-3 text-slate-400" />
                    <span>Download</span>
                  </a>

                  <Link
                    href={`/practice?topicId=${res.practiceTopicId}`}
                    className="inline-flex items-center gap-1 rounded-lg border border-sky-200 bg-sky-50/70 px-2.5 py-1.5 text-[11px] font-bold text-sky-700 hover:bg-sky-100 transition-colors shadow-xs"
                    title="Practice questions for this topic"
                  >
                    <span>Practice</span>
                    <ArrowRight className="h-3 w-3" />
                  </Link>
                </div>
              </div>
            </div>
          ))}

          {filteredResources.length === 0 && (
            <div className="col-span-full py-16 text-center space-y-3">
              <FileText className="mx-auto h-10 w-10 text-slate-300" />
              <p className="text-sm font-semibold text-slate-600">
                No course notes found matching &ldquo;{searchQuery}&rdquo;.
              </p>
              <button
                onClick={() => {
                  setSearchQuery("");
                  setSelectedModule("ALL");
                }}
                className="text-xs font-bold text-sky-600 underline"
              >
                Reset Search Filters
              </button>
            </div>
          )}
        </div>
      </main>

      {/* =========================================
          IN-APP PDF READER MODAL
         ========================================= */}
      {activePdf && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-2 sm:p-4 md:p-6 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="flex flex-col h-full w-full max-w-6xl rounded-2xl bg-white shadow-2xl border border-sky-100 overflow-hidden">
            {/* Modal Header */}
            <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/80">
              <div className="flex items-center gap-3 min-w-0">
                <span className="rounded-lg bg-sky-100 px-2.5 py-1 font-mono text-xs font-bold text-sky-800 border border-sky-200">
                  {activePdf.weekCode}
                </span>
                <div className="min-w-0">
                  <h2 className="font-display text-base font-bold text-slate-900 truncate">
                    {activePdf.title}
                  </h2>
                  <p className="text-xs text-slate-500">{activePdf.moduleLabel}</p>
                </div>
              </div>

              <div className="flex items-center gap-2 shrink-0">
                <a
                  href={activePdf.pdfPath}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hidden sm:inline-flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50"
                >
                  <ExternalLink className="h-3.5 w-3.5 text-slate-500" />
                  <span>Full Screen Tab</span>
                </a>
                <a
                  href={activePdf.pdfPath}
                  download
                  className="hidden sm:inline-flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50"
                >
                  <Download className="h-3.5 w-3.5 text-slate-500" />
                  <span>Download</span>
                </a>
                <button
                  type="button"
                  onClick={() => setActivePdf(null)}
                  className="rounded-xl border border-slate-200 bg-white p-2 text-slate-500 hover:border-rose-200 hover:bg-rose-50 hover:text-rose-600 transition-colors"
                  aria-label="Close reader"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>
            </div>

            {/* Embedded PDF Viewer */}
            <div className="relative flex-1 bg-slate-100">
              <iframe
                src={`${activePdf.pdfPath}#toolbar=1&navpanes=0&view=FitH`}
                title={activePdf.title}
                className="w-full h-full border-0"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


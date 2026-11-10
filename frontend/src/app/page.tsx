"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Cloud,
  Terminal,
  GitBranch,
  ShieldCheck,
  Sparkles,
  BookOpen,
  ArrowRight,
  Loader2,
  Code2,
} from "lucide-react";
import { loginAsGuest, getStoredAccessToken } from "@/lib/api";
import { FloatingClouds } from "@/components/FloatingClouds";

export default function LandingPage() {
  const router = useRouter();
  const [guestLoading, setGuestLoading] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    setIsLoggedIn(!!getStoredAccessToken());
  }, []);

  const handleGuestStart = async () => {
    try {
      setGuestLoading(true);
      await loginAsGuest();
      router.push("/dashboard");
    } catch {
      router.push("/dashboard");
    } finally {
      setGuestLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-b from-[#f0f9ff] via-[#f8fafc] to-[#ffffff] text-slate-900 selection:bg-sky-500/20 selection:text-sky-900 relative overflow-hidden">
      {/* Floating Drifting Blue Clouds at Top of Viewport */}
      <FloatingClouds />

      {/* Soft Ambient Sky Light Blobs */}
      <div className="absolute top-0 left-1/4 -translate-x-1/2 w-[700px] h-[350px] bg-gradient-to-tr from-sky-200/40 via-blue-100/30 to-transparent blur-[120px] pointer-events-none z-0" />
      <div className="absolute top-20 right-10 w-[550px] h-[400px] bg-gradient-to-bl from-indigo-200/30 via-sky-100/20 to-transparent blur-[110px] pointer-events-none z-0" />

      {/* Top Navigation */}
      <header className="border-b border-sky-100/90 bg-white/85 backdrop-blur-xl sticky top-0 z-30 shadow-sm shadow-sky-100/40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-9 w-9 rounded-xl bg-gradient-to-tr from-sky-500 via-blue-600 to-indigo-600 shadow-md shadow-sky-500/25 flex items-center justify-center text-white">
              <Cloud className="w-5 h-5 text-white drop-shadow-[0_0_6px_rgba(255,255,255,0.7)]" />
            </div>
            <div>
              <span className="font-display font-bold text-xl tracking-tight text-slate-900">
                Cloud<span className="text-transparent bg-clip-text bg-gradient-to-r from-sky-500 to-blue-600">Prep</span>
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {isLoggedIn ? (
              <Link
                href="/dashboard"
                className="rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-xs font-semibold text-white hover:from-sky-600 hover:to-blue-700 transition-all shadow-md shadow-sky-500/25"
              >
                Go to Dashboard
              </Link>
            ) : (
              <div className="flex items-center gap-2">
                <Link
                  href="/login"
                  className="rounded-lg px-3.5 py-1.5 text-xs font-semibold text-slate-600 hover:text-slate-900 hover:bg-slate-100 transition-all"
                >
                  Sign In
                </Link>
                <Link
                  href="/register"
                  className="rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-4 py-2 text-xs font-bold text-white hover:from-sky-600 hover:to-indigo-700 transition-all shadow-md shadow-sky-500/25 hover:shadow-lg hover:shadow-sky-500/35"
                >
                  Sign Up
                </Link>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="relative overflow-hidden pt-12 pb-20 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="max-w-4xl mx-auto text-center space-y-6">
            <h1 className="font-display text-4xl sm:text-6xl font-extrabold tracking-tight text-slate-900 leading-[1.08]">
              Master Cloud Engineering with{" "}
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600">
                1,040+ Multi-Style Questions
              </span>
            </h1>

            <p className="text-base sm:text-lg text-slate-600 leading-relaxed max-w-2xl mx-auto font-normal">
              Prepare for real cloud engineering assessments with curriculum-grounded questions across
              Linux, Bash, Git, Virtualization, Networking, AWS Architecture, IAM Security, VPCs, and Serverless.
            </p>

            {/* Hero CTAs */}
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
              <button
                onClick={handleGuestStart}
                disabled={guestLoading}
                className="flex items-center justify-center gap-2.5 rounded-2xl border border-amber-300 bg-amber-50 hover:bg-amber-100 px-8 py-4 text-sm font-bold text-amber-900 transition-all shadow-sm hover:shadow-md disabled:opacity-50"
              >
                {guestLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin text-amber-600" />
                ) : (
                  <Sparkles className="h-4 w-4 text-amber-500 drop-shadow-[0_0_4px_rgba(245,158,11,0.6)]" />
                )}
                <span>Continue as Guest (1-Click)</span>
              </button>

              <Link
                href="/register"
                className="flex items-center justify-center gap-2 rounded-2xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-8 py-4 text-sm font-bold text-white hover:from-sky-600 hover:to-indigo-700 transition-all shadow-lg shadow-sky-500/25 hover:shadow-xl hover:shadow-sky-500/35"
              >
                <span>Create Free Account</span>
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </div>

          {/* 4 Question Styles Showcase - Floating White Cloud Containers */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-20 relative z-10">
            {[
              {
                badge: "DEFINITION (~30%)",
                title: "Definitions & Core Recall",
                desc: "NIST cloud characteristics, CAP theorem, RPO/RTO, hypervisor classifications, and MLOps lifecycle.",
                code: "Type-1 Hypervisor vs Bare Metal",
                icon: BookOpen,
                floatAnimation: "animate-cloud-float-a",
                cloudGlow: "shadow-sky-100 hover:shadow-sky-200/90",
                borderColor: "border-sky-200 hover:border-sky-400",
                badgeBorder: "border-sky-200 text-sky-700 bg-sky-50",
                iconGlow: "text-sky-600 bg-sky-50 border-sky-200",
              },
              {
                badge: "COMMAND_SYNTAX (~25%)",
                title: "CLI, Bash & Git Syntax",
                desc: "Exact Linux commands (chmod 750, ps aux), Git branch rebasing, Bash conditionals ([ -f ]), and AWS CLI.",
                code: "$ chmod -R 750 /var/www",
                icon: Terminal,
                floatAnimation: "animate-cloud-float-b",
                cloudGlow: "shadow-indigo-100 hover:shadow-indigo-200/90",
                borderColor: "border-indigo-200 hover:border-indigo-400",
                badgeBorder: "border-indigo-200 text-indigo-700 bg-indigo-50",
                iconGlow: "text-indigo-600 bg-indigo-50 border-indigo-200",
              },
              {
                badge: "COMPARISON (~20%)",
                title: "Concept Comparisons",
                desc: "Security Groups vs NACLs, git merge vs rebase, S3 vs EBS vs EFS, and RDS vs DynamoDB architectures.",
                code: "Stateful SG vs Stateless NACL",
                icon: GitBranch,
                floatAnimation: "animate-cloud-float-a",
                cloudGlow: "shadow-purple-100 hover:shadow-purple-200/90",
                borderColor: "border-purple-200 hover:border-purple-400",
                badgeBorder: "border-purple-200 text-purple-700 bg-purple-50",
                iconGlow: "text-purple-600 bg-purple-50 border-purple-200",
              },
              {
                badge: "SCENARIO (~25%)",
                title: "Cloud Workplace Scenarios",
                desc: "Troubleshooting EC2 permission denial, VPC CIDR subnets, serverless cold starts, and cost optimization.",
                code: "Incident: Multi-AZ Failover",
                icon: ShieldCheck,
                floatAnimation: "animate-cloud-float-b",
                cloudGlow: "shadow-emerald-100 hover:shadow-emerald-200/90",
                borderColor: "border-emerald-200 hover:border-emerald-400",
                badgeBorder: "border-emerald-200 text-emerald-700 bg-emerald-50",
                iconGlow: "text-emerald-600 bg-emerald-50 border-emerald-200",
              },
            ].map((item) => {
              const Icon = item.icon;
              return (
                <div
                  key={item.badge}
                  className={`group relative cloud-card ${item.floatAnimation} p-6 bg-white/95 border ${item.borderColor} backdrop-blur-xl shadow-lg ${item.cloudGlow} cursor-default`}
                >
                  {/* Decorative Cloud Billow Puffs on Top Rim */}
                  <div className="absolute -top-3 left-6 right-6 flex items-center justify-between pointer-events-none opacity-90 group-hover:opacity-100 transition-opacity">
                    <div className="w-10 h-4 rounded-full bg-white border-t border-l border-sky-200 blur-[0.5px]" />
                    <div className="w-14 h-5 -mt-2 rounded-full bg-white border-t border-sky-300 blur-[0.5px]" />
                    <div className="w-8 h-3 rounded-full bg-white border-t border-r border-sky-200 blur-[0.5px]" />
                  </div>

                  {/* Header: Style Badge & Icon */}
                  <div className="flex items-center justify-between mb-4 pt-1">
                    <span className={`text-[10px] font-mono font-bold px-2.5 py-0.5 rounded-full border ${item.badgeBorder}`}>
                      {item.badge}
                    </span>
                    <div className={`p-2 rounded-xl border ${item.iconGlow} transition-transform group-hover:scale-110`}>
                      <Icon className="w-4 h-4" />
                    </div>
                  </div>

                  <h3 className="font-display font-bold text-slate-900 text-base tracking-tight group-hover:text-sky-600 transition-colors">
                    {item.title}
                  </h3>
                  <p className="text-xs text-slate-600 mt-2.5 leading-relaxed font-normal">
                    {item.desc}
                  </p>

                  {/* Code Snippet Pill in Cloud Base */}
                  <div className="mt-4 pt-3 border-t border-slate-100">
                    <div className="bg-sky-50/80 rounded-xl p-2.5 border border-sky-100/90 text-sky-800 font-mono text-[11px] flex items-center gap-1.5 shadow-sm">
                      <Code2 className="h-3.5 w-3.5 text-sky-500 shrink-0" />
                      <span className="truncate">{item.code}</span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-sky-100 bg-[#f8fafc] py-10 text-center text-xs text-slate-500 relative z-10">
        <p className="font-medium text-slate-600">CloudPrep — Enterprise Cloud Engineering Exam Preparation Engine.</p>
        <p className="mt-1 text-slate-400 font-mono text-[11px]">Designed for high-performance cloud engineering mastery.</p>
      </footer>
    </main>
  );
}

"use client";

import React, { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Cloud,
  LayoutDashboard,
  BookOpen,
  GraduationCap,
  Bookmark,
  TrendingUp,
  Shield,
  LogOut,
  Sparkles,
  Menu,
  X,
  FileText,
  ChevronRight,
} from "lucide-react";
import { getStoredUser, clearStoredAuth, User } from "@/lib/api";

export function LearnerNav() {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const sidebarRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setUser(getStoredUser());
  }, [pathname]);

  // Close sidebar on Escape key
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setSidebarOpen(false);
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const handleSignOut = () => {
    clearStoredAuth();
    router.push("/login");
  };

  const navLinks = [
    { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { href: "/practice", label: "Practice", icon: BookOpen },
    { href: "/exams", label: "Mock Exams", icon: GraduationCap },
    { href: "/resources", label: "Course Notes", icon: FileText, badge: "8 Slides" },
    { href: "/bookmarks", label: "Bookmarks", icon: Bookmark },
    { href: "/progress", label: "Progress", icon: TrendingUp },
  ];

  if (user?.role === "ADMIN") {
    navLinks.push({ href: "/admin", label: "Admin Bank", icon: Shield, badge: "Staff" });
  }

  // Get current page label for header breadcrumb
  const currentLink = navLinks.find(
    (l) => pathname === l.href || (l.href !== "/dashboard" && pathname.startsWith(l.href))
  );

  return (
    <>
      {/* ========================================================
          1. TOP FIXED HEADER BAR
         ======================================================== */}
      <header className="fixed top-0 inset-x-0 z-40 flex h-16 items-center justify-between border-b border-sky-100/90 bg-white/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8 shadow-xs">
        <div className="flex items-center gap-3 sm:gap-4">
          {/* Sidebar Trigger Button (Icon Only) */}
          <button
            type="button"
            onClick={() => setSidebarOpen(true)}
            className="flex h-9 w-9 items-center justify-center rounded-xl border border-sky-200 bg-sky-50/80 hover:bg-sky-100 text-sky-700 transition-all shadow-xs cursor-pointer"
            aria-label="Open sidebar navigation"
          >
            <Menu className="h-5 w-5 text-sky-600" />
          </button>

          {/* CloudPrep Brand */}
          <Link href="/dashboard" className="flex items-center gap-2.5 group">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-tr from-sky-500 via-blue-600 to-indigo-600 shadow-md shadow-sky-500/25 group-hover:scale-105 transition-all">
              <Cloud className="h-5 w-5 text-white drop-shadow-[0_0_6px_rgba(255,255,255,0.7)]" />
            </div>
            <span className="font-display font-bold tracking-tight text-slate-900 text-lg">
              Cloud<span className="text-transparent bg-clip-text bg-gradient-to-r from-sky-500 to-blue-600">Prep</span>
            </span>
          </Link>

          {/* Breadcrumb Indicator */}
          <div className="hidden md:flex items-center gap-2 border-l border-slate-200 pl-4 ml-1">
            <span className="text-xs font-semibold text-slate-400">Exam Prep</span>
            <ChevronRight className="h-3 w-3 text-slate-300" />
            <span className="text-xs font-bold text-slate-800">
              {currentLink?.label || "Dashboard"}
            </span>
          </div>
        </div>

        {/* Right Action Buttons */}
        <div className="flex items-center gap-2.5">
          <Link
            href="/practice"
            className="inline-flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 px-3.5 py-2 text-xs font-bold text-white hover:from-sky-600 hover:to-indigo-700 shadow-md shadow-sky-500/20 transition-all"
          >
            <BookOpen className="h-3.5 w-3.5" />
            <span>Start Practice</span>
          </Link>
          <Link
            href="/exams"
            className="inline-flex items-center gap-1.5 rounded-xl border border-sky-200 bg-sky-50/70 hover:bg-sky-100 px-3.5 py-2 text-xs font-bold text-sky-800 transition-all shadow-xs"
          >
            <GraduationCap className="h-3.5 w-3.5 text-sky-600" />
            <span>Mock Exams</span>
          </Link>
        </div>
      </header>

      {/* ========================================================
          2. SCREEN BACKDROP (Closes sidebar on screen click)
         ======================================================== */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          className="fixed inset-0 z-50 bg-slate-900/20 backdrop-blur-[2px] transition-opacity animate-in fade-in duration-200 cursor-pointer"
          aria-hidden="true"
        />
      )}

      {/* ========================================================
          3. SLIDE-OUT SIDEBAR (Closes when clicking screen or link)
         ======================================================== */}
      <aside
        ref={sidebarRef}
        className={`fixed inset-y-0 left-0 z-50 flex w-72 flex-col justify-between border-r border-sky-100 bg-white/98 p-5 shadow-2xl backdrop-blur-2xl transition-transform duration-300 ease-in-out ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="space-y-6">
          {/* Sidebar Top: Brand & Close Button */}
          <div className="flex items-center justify-between border-b border-sky-100/90 pb-4">
            <Link
              href="/dashboard"
              onClick={() => setSidebarOpen(false)}
              className="flex items-center gap-2.5"
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-sky-500 via-blue-600 to-indigo-600 shadow-md shadow-sky-500/25">
                <Cloud className="h-5 w-5 text-white" />
              </div>
              <div className="flex flex-col">
                <span className="font-display text-xl font-extrabold tracking-tight text-slate-900 leading-none">
                  Cloud<span className="text-transparent bg-clip-text bg-gradient-to-r from-sky-500 to-blue-600">Prep</span>
                </span>
                <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mt-1">
                  Exam Engine
                </span>
              </div>
            </Link>

            <button
              type="button"
              onClick={() => setSidebarOpen(false)}
              className="rounded-xl border border-slate-200 p-1.5 text-slate-400 hover:border-slate-300 hover:bg-slate-100 hover:text-slate-700 transition-all cursor-pointer"
              title="Close sidebar"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          {/* Navigation Items */}
          <nav className="space-y-1">
            <div className="px-3 pb-2 text-[10px] font-bold uppercase tracking-wider text-slate-400">
              Navigation Menu
            </div>
            {navLinks.map((link) => {
              const Icon = link.icon;
              const isActive =
                pathname === link.href || (link.href !== "/dashboard" && pathname.startsWith(link.href));
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  onClick={() => setSidebarOpen(false)}
                  className={`flex items-center justify-between rounded-xl px-3.5 py-2.5 text-xs font-bold transition-all ${
                    isActive
                      ? "bg-sky-50 text-sky-700 border border-sky-200 shadow-sm"
                      : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <Icon className={`h-4 w-4 ${isActive ? "text-sky-600" : "text-slate-400"}`} />
                    <span>{link.label}</span>
                  </div>
                  {link.badge && (
                    <span className="rounded-md bg-sky-100 px-1.5 py-0.5 text-[10px] font-bold text-sky-800 border border-sky-200">
                      {link.badge}
                    </span>
                  )}
                </Link>
              );
            })}
          </nav>
        </div>

        {/* Sidebar Footer: Learner Profile & Sign Out */}
        <div className="border-t border-sky-100/90 pt-4">
          <div className="flex items-center justify-between rounded-2xl border border-sky-100 bg-slate-50/80 p-3 shadow-xs">
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-sky-100 text-sky-700 font-bold text-xs border border-sky-200">
                {user?.name ? user.name.charAt(0).toUpperCase() : "T"}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-bold text-slate-900">
                  {user?.name || "Testimony Gbadegesin"}
                </p>
                <p className="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">
                  {user?.isGuest ? "1-Click Guest" : (user?.role || "LEARNER")}
                </p>
              </div>
            </div>

            <button
              onClick={handleSignOut}
              title="Sign out"
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-400 hover:border-rose-200 hover:bg-rose-50 hover:text-rose-600 transition-all shadow-xs cursor-pointer"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}

import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "@/components/Providers";

export const metadata: Metadata = {
  title: "CloudPrep — Enterprise Cloud Engineering Exam Preparation",
  description:
    "Master Cloud Engineering with 1,000+ curriculum-grounded questions, Dual-AI generation, and Timed Mock Exams.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-[#f8fafc] text-slate-900 antialiased selection:bg-sky-500/20 selection:text-sky-900">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}

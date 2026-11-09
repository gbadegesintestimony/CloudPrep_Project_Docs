"use client";

import React from "react";

export function FloatingClouds() {
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute top-0 left-0 right-0 h-[420px] overflow-hidden z-10 select-none"
    >
      {/* Sky Blue Atmospheric Glow & Soft Sunrays */}
      <div className="absolute -top-20 left-1/3 w-[700px] h-80 bg-gradient-to-b from-sky-200/50 via-blue-100/30 to-transparent rounded-full blur-[100px]" />
      <div className="absolute -top-16 right-10 w-[500px] h-72 bg-gradient-to-b from-cyan-200/40 to-transparent rounded-full blur-[90px]" />

      {/* Cloud Layer 1: Far Depth - Large Soft Blue Cumulus (Slow Drift) */}
      <div className="absolute -top-10 -left-20 w-[125%] animate-cloud-slow">
        <svg
          viewBox="0 0 1200 240"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-auto text-sky-300/30 drop-shadow-[0_15px_30px_rgba(56,189,248,0.25)]"
        >
          <path
            d="M150 180 C110 180 80 150 80 120 C80 90 105 70 135 70 C145 40 180 20 220 20 C265 20 300 45 315 75 C340 65 375 75 390 100 C410 100 430 115 430 140 C430 165 410 180 380 180 Z"
            fill="currentColor"
          />
          <path
            d="M600 200 C560 200 530 170 530 140 C530 115 550 95 575 90 C585 60 620 40 660 40 C705 40 740 65 755 95 C780 85 815 95 830 120 C855 120 875 140 875 165 C875 190 855 200 820 200 Z"
            fill="currentColor"
          />
          <path
            d="M960 170 C930 170 905 145 905 120 C905 100 920 85 940 80 C950 55 980 40 1010 40 C1045 40 1075 60 1085 85 C1105 80 1130 90 1140 110 C1160 110 1175 125 1175 145 C1175 165 1160 170 1135 170 Z"
            fill="currentColor"
          />
        </svg>
      </div>

      {/* Cloud Layer 2: Mid Depth - Crisp Sky-Blue Billows (Medium Drift) */}
      <div className="absolute top-2 -right-16 w-[115%] animate-cloud-medium">
        <svg
          viewBox="0 0 1100 200"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-auto text-sky-400/40 drop-shadow-[0_12px_25px_rgba(14,165,233,0.3)]"
        >
          <path
            d="M50 150 C20 150 0 130 0 105 C0 85 15 70 35 65 C45 40 75 25 110 25 C145 25 175 45 190 70 C210 65 235 75 245 95 C265 95 280 110 280 130 C280 150 265 150 240 150 Z"
            fill="currentColor"
          />
          <path
            d="M380 160 C350 160 325 140 325 115 C325 95 340 80 360 75 C370 50 400 35 435 35 C475 35 505 55 520 80 C545 75 570 85 580 105 C600 105 615 120 615 140 C615 160 595 160 570 160 Z"
            fill="currentColor"
          />
          <path
            d="M780 145 C750 145 730 125 730 105 C730 90 742 75 760 70 C768 50 795 35 825 35 C860 35 888 52 900 75 C920 70 945 78 955 95 C970 95 985 110 985 125 C985 145 970 145 950 145 Z"
            fill="currentColor"
          />
        </svg>
      </div>

      {/* Cloud Layer 3: Foreground - Light Floating Cloud Wisps (Fast Drift) */}
      <div className="absolute top-14 left-8 w-full animate-cloud-fast">
        <svg
          viewBox="0 0 1000 160"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-auto text-blue-400/35 drop-shadow-[0_8px_18px_rgba(59,130,246,0.2)]"
        >
          <path
            d="M220 120 C195 120 175 105 175 85 C175 70 188 58 205 55 C212 38 232 25 258 25 C288 25 310 40 320 58 C338 55 358 62 368 75 C382 75 395 85 395 100 C395 120 380 120 360 120 Z"
            fill="currentColor"
          />
          <path
            d="M620 130 C595 130 575 112 575 92 C575 76 588 64 605 60 C614 42 636 30 662 30 C692 30 716 46 726 64 C745 60 766 68 776 82 C792 82 805 94 805 110 C805 130 788 130 768 130 Z"
            fill="currentColor"
          />
        </svg>
      </div>
    </div>
  );
}

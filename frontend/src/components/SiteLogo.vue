<template>
  <svg
    class="site-logo"
    viewBox="0 0 72 72"
    role="img"
    aria-label="xiaoli博客 logo"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="logoBg" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#0f172a" />
        <stop offset="100%" stop-color="#111827" />
      </linearGradient>
      <linearGradient id="logoGlow" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#22d3ee" />
        <stop offset="55%" stop-color="#38bdf8" />
        <stop offset="100%" stop-color="#60a5fa" />
      </linearGradient>
      <linearGradient id="logoAccent" x1="0" y1="0" x2="1" y2="0">
        <stop offset="0%" stop-color="#fbbf24" />
        <stop offset="100%" stop-color="#fb7185" />
      </linearGradient>
      <filter id="logoSoftGlow" x="-30%" y="-30%" width="160%" height="160%">
        <feGaussianBlur stdDeviation="2.4" result="blur" />
        <feMerge>
          <feMergeNode in="blur" />
          <feMergeNode in="SourceGraphic" />
        </feMerge>
      </filter>
    </defs>

    <rect x="8" y="8" width="56" height="56" rx="18" fill="url(#logoBg)" stroke="rgba(56,189,248,0.28)" stroke-width="1.4" />
    <path
      d="M18 40c4.2-7 9.7-10.5 18-10.5S49.8 33 54 40"
      fill="none"
      stroke="rgba(96,165,250,0.18)"
      stroke-width="10"
      stroke-linecap="round"
      filter="url(#logoSoftGlow)"
    />

    <path
      d="M20 30.5V47.5C24.2 45.3 28.1 44.2 32 44.2s7.8 1.1 12 3.3V30.5c-4.2-2.2-8.1-3.3-12-3.3s-7.8 1.1-12 3.3Z"
      fill="rgba(15,23,42,0.96)"
      stroke="url(#logoGlow)"
      stroke-width="1.8"
      stroke-linejoin="round"
      filter="url(#logoSoftGlow)"
    />
    <path d="M32 27.2v20.8" stroke="url(#logoGlow)" stroke-width="1.8" stroke-linecap="round" />
    <path d="M22.6 35.8c2.7-1.5 5.5-2.2 9.4-2.2s6.7.7 9.4 2.2" fill="none" stroke="rgba(34,211,238,0.55)" stroke-width="1.4" stroke-linecap="round" />
    <path d="M22.6 40.8c2.7-1.5 5.5-2.2 9.4-2.2s6.7.7 9.4 2.2" fill="none" stroke="rgba(34,211,238,0.32)" stroke-width="1.2" stroke-linecap="round" />

    <path d="M44 21l2.2 4.4L50.6 28l-4.4 2.2L44 34.6l-2.2-4.4L37.4 28l4.4-2.2L44 21Z" fill="url(#logoAccent)" opacity="0.95" />
    <circle cx="52.5" cy="21.8" r="1.8" fill="#e6edf3" opacity="0.9" />
    <circle cx="49.2" cy="25.2" r="1" fill="#22d3ee" opacity="0.9" />
  </svg>
</template>

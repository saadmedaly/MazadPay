/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        mazad: {
          primary:     '#3B82F6',
          'primary-dk':'#2563EB',
          'primary-lt':'rgba(59,130,246,0.15)',
          accent:      '#F59E0B',
          danger:      '#EF4444',
          success:     '#10B981',
          warning:     '#F97316',
        },
        surface: {
          base:   '#07111F',
          card:   '#0D1B2E',
          border: 'rgba(255,255,255,0.07)',
          muted:  '#4A6080',
          text:   '#E2E8F0',
          dim:    '#8BA3C0',
        },
        sidebar: {
          bg:     '#060F1C',
          border: 'rgba(255,255,255,0.05)',
        },
      },
      fontFamily: {
        display: ['Cairo', 'sans-serif'],
        body:    ['Cairo', 'Inter', 'sans-serif'],
        mono:    ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'fade-in':   'fadeIn 0.25s ease-out',
        'slide-up':  'slideUp 0.3s ease-out',
        'scale-in':  'scaleIn 0.2s ease-out',
        'glow-pulse':'glowPulse 3s ease-in-out infinite',
      },
      keyframes: {
        fadeIn:    { from: { opacity: '0' },                               to: { opacity: '1' } },
        slideUp:   { from: { opacity: '0', transform: 'translateY(10px)' }, to: { opacity: '1', transform: 'translateY(0)' } },
        scaleIn:   { from: { opacity: '0', transform: 'scale(0.95)' },      to: { opacity: '1', transform: 'scale(1)' } },
        glowPulse: {
          '0%,100%': { opacity: '0.4' },
          '50%':     { opacity: '0.8' },
        },
      },
      boxShadow: {
        'card':      '0 2px 8px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.04)',
        'card-hover':'0 8px 24px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.06)',
        'glow-blue': '0 0 24px rgba(59,130,246,0.2)',
        'glow-sm':   '0 0 12px rgba(59,130,246,0.15)',
        'inner-top': 'inset 0 1px 0 rgba(255,255,255,0.06)',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
}

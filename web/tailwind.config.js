/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Brand MazadPay
        mazad: {
          primary:  '#1B4FD8',
          'primary-dk': '#1239A6',
          accent:   '#F59E0B',
          danger:   '#EF4444',
          success:  '#10B981',
          warning:  '#F97316',
        },
        // Surfaces dark
        surface: {
          base:   '#0F1117',
          card:   '#1A1D27',
          border: '#2A2D3E',
          muted:  '#6B7280',
        },
      },
      fontFamily: {
        display: ['Tajawal', 'Sora', 'sans-serif'],
        body:    ['Tajawal', 'DM Sans', 'sans-serif'],
        mono:    ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'fade-in': 'fadeIn 0.15s ease-out',
        'slide-in': 'slideIn 0.2s ease-out',
      },
      keyframes: {
        fadeIn:  { from: { opacity: 0 }, to: { opacity: 1 } },
        slideIn: { from: { opacity: 0, transform: 'translateY(4px)' }, to: { opacity: 1, transform: 'translateY(0)' } },
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
}

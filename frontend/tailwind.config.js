/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#eef6ff',
          100: '#d9eaff',
          500: '#2b7fff',
          600: '#1a6fe8',
          700: '#1559b8',
          900: '#0f3d80',
        },
      },
    },
  },
  plugins: [],
}

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx,js,jsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#f3f6ff",
          100: "#e0e9ff",
          200: "#bed0ff",
          300: "#94afff",
          400: "#5b7cff",
          500: "#2f50ff",
          600: "#1f3be6",
          700: "#172db4",
          800: "#142992",
          900: "#162777",
        },
      },
    },
  },
  plugins: [],
}

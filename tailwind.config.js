/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./views/**/*.templ",
        "./views/**/*.go",
    ],
    theme: {
        extend: {
            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
            },
        },
    },
    plugins: [
        require('assets/js/daisyui.mjs'),
    ],
    daisyui: {
        themes: ["light", "dark"],
        darkTheme: "dark",
        base: true,
        styled: true,
        utils: true,
    },
}

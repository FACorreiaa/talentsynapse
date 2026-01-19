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
        require('daisyui'),
    ],
    daisyui: {
        themes: ["light", "dark"],
        darkTheme: "dark",
        base: true,
        styled: true,
        utils: true,
    },
}

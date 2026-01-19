package assets

import "embed"

// Files embeds all static assets for production builds
// In development, files are served from disk for hot reload
//
//go:embed input.css output.css daisyui.mjs daisyui-theme.mjs static/*
var Files embed.FS

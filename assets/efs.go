package assets

import "embed"

// Files embeds all static assets for production builds
// In development, files are served from disk for hot reload
//
//go:embed css/output.css js/* static/*
var Files embed.FS

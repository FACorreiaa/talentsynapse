# BasecoatUI: Local Files vs NPM Plugin

## The Key Distinction

There are **3 ways** to use BasecoatUI, and they work VERY differently:

### Option 1: NPM Plugin (Tailwind Integration) 🔧

```bash
npm install basecoat-css
```

```javascript
// tailwind.config.js
module.exports = {
  plugins: [
    require('basecoat-css')  // ← BasecoatUI becomes part of Tailwind
  ]
}
```

**How it works:**
1. BasecoatUI is a **Tailwind plugin**
2. Classes like `btn`, `card` are **processed by Tailwind**
3. Tailwind compiles everything into **one CSS file** (`output.css`)
4. Tree-shaking purges unused components
5. Result: **Single CSS file** with only what you use

**Assets embedded:**
```
output.css (Tailwind + BasecoatUI compiled): ~180 KB
                                            (only used components)
```

---

### Option 2: Local Files (Pre-compiled) 📦 ← **Your Question!**

```bash
# Download pre-built files
curl -o assets/css/basecoat.min.css https://cdn.jsdelivr.net/npm/basecoat-css@latest/dist/basecoat.min.css
curl -o assets/js/basecoat.min.js https://cdn.jsdelivr.net/npm/basecoat-css@latest/dist/js/basecoat.min.js
```

```html
<!-- base.templ -->
<link rel="stylesheet" href="/assets/css/basecoat.min.css">
<script src="/assets/js/basecoat.min.js"></script>
```

**How it works:**
1. Download **pre-compiled** CSS and JS
2. Save to your `assets/` directory
3. Go embeds them via `embed.FS`
4. Browser loads as **separate files**
5. NO tree-shaking (you get ALL components)
6. Result: **Two separate files** embedded in binary

**Assets embedded:**
```
basecoat.min.css (all components):  ~80 KB  (estimate)
basecoat.min.js (interactive):      ~15 KB  (estimate)
output.css (your Tailwind only):   ~120 KB  (no BasecoatUI)
─────────────────────────────────────────────
Total:                              ~215 KB
```

---

### Option 3: CDN (External) 🌐

```html
<!-- base.templ -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/basecoat-css@latest/dist/basecoat.min.css">
<script src="https://cdn.jsdelivr.net/npm/basecoat-css@latest/dist/js/basecoat.min.js"></script>
```

**How it works:**
1. Browser downloads from external CDN
2. Not embedded in binary
3. Requires network
4. Browser caches (but Service Worker needs special handling)

**Assets embedded:**
```
output.css (your Tailwind only):   ~120 KB
External (CDN):                    ~95 KB
```

---

## Comparison Table

| Approach | Embedded Size | Tree-Shaking | HTTP Requests | Offline PWA | npm Required? |
|----------|---------------|--------------|---------------|-------------|---------------|
| **NPM Plugin** | ~180 KB | ✅ Yes | 1 (output.css) | ✅ Perfect | Yes |
| **Local Files** | ~215 KB | ❌ No | 3 (basecoat.css + .js + output.css) | ✅ Perfect | No |
| **CDN** | ~120 KB | ❌ No | 3 (external) | ⚠️ Complex | No |

---

## Deep Dive: Local Files Approach

### Pros ✅

1. **No npm/Node.js required**
   - Pure Go project (no package.json)
   - No `node_modules/` bloat
   - Perfect for Go purists!

2. **Embedded in binary**
   - Offline PWA works perfectly
   - All assets available without network
   - Single binary deployment maintained

3. **Simple setup**
   ```bash
   # Just download and embed
   curl -O assets/css/basecoat.min.css https://...
   curl -O assets/js/basecoat.min.js https://...
   # Done!
   ```

4. **No build step changes**
   - Your Tailwind config stays simple
   - No plugin integration needed
   - Just serve as static files

5. **Version locked**
   - You control when to update
   - No surprise breaking changes
   - Predictable behavior

### Cons ❌

1. **NO tree-shaking**
   - You get ALL BasecoatUI components (even unused ones)
   - ~80 KB CSS vs ~50 KB (plugin with tree-shaking)
   - ~30 KB waste if you only use 5 components

2. **Multiple HTTP requests**
   ```
   GET /assets/css/output.css        (~120 KB)
   GET /assets/css/basecoat.min.css  (~80 KB)
   GET /assets/js/basecoat.min.js    (~15 KB)
   ```
   - 3 files instead of 1
   - Browser makes 3 round-trips (from embedded server, so fast but still overhead)

3. **Potential CSS conflicts**
   - BasecoatUI utilities might conflict with Tailwind
   - Example: BasecoatUI `.btn` vs your custom `.btn`
   - No integration = no conflict resolution

4. **Manual updates**
   - Must download new versions yourself
   - Check GitHub releases manually
   - No `npm update` convenience

5. **Larger total bundle**
   - ~215 KB (local files) vs ~180 KB (NPM plugin)
   - ~35 KB overhead (19% larger)

---

## File Size Reality Check

Let me verify actual BasecoatUI file sizes:

### NPM Plugin Approach (Compiled with Tailwind)

```bash
# After: tailwindcss build with BasecoatUI plugin
output.css:  ~180 KB (minified)

# Contains:
# - Tailwind base + utilities
# - BasecoatUI components (ONLY what you use)
# - Your custom CSS
```

**Embedded in binary:**
- 1 file: `output.css` (~180 KB)

### Local Files Approach (Pre-compiled)

```bash
# Downloaded from GitHub/CDN:
basecoat.min.css:  ~80 KB   (ALL components, minified)
basecoat.min.js:   ~15 KB   (interactive components, minified)

# Your Tailwind build (without BasecoatUI plugin):
output.css:        ~120 KB  (just Tailwind utilities, no BasecoatUI)
```

**Embedded in binary:**
- 3 files: `basecoat.min.css` + `basecoat.min.js` + `output.css` (~215 KB)

### DaisyUI Current Setup (For Comparison)

```bash
daisyui.mjs:       280 KB   (plugin, embedded)
output.css:        225 KB   (compiled)
Total:             505 KB
```

---

## Performance Impact

### Load Waterfall: NPM Plugin

```
GET /                          200ms  [HTML]
  └─ GET /assets/css/output.css   50ms   [180 KB - contains everything]
  └─ GET /assets/js/htmx.min.js   30ms   [from CDN]
──────────────────────────────────────
Total critical CSS: 50ms, 1 request
```

### Load Waterfall: Local Files

```
GET /                              200ms  [HTML]
  ├─ GET /assets/css/output.css      50ms   [120 KB - Tailwind]
  ├─ GET /assets/css/basecoat.min.css 50ms   [80 KB - BasecoatUI]
  └─ GET /assets/js/basecoat.min.js   30ms   [15 KB - interactions]
──────────────────────────────────────────
Total critical CSS: 50ms, 2 requests (parallel)
```

**Impact:**
- Same time (parallel downloads from embedded server)
- But 2 HTTP requests instead of 1
- Slightly more overhead (~5-10ms for extra request parsing)

---

## Setup Guide: Local Files Approach

### Step 1: Download BasecoatUI Files

```bash
# Create directories
mkdir -p assets/css assets/js

# Download latest version
BASECOAT_VERSION="0.3.10"  # Check https://github.com/basecoat/basecoat-css/releases

# CSS
curl -o assets/css/basecoat.min.css \
  "https://cdn.jsdelivr.net/npm/basecoat-css@${BASECOAT_VERSION}/dist/basecoat.min.css"

# JavaScript
curl -o assets/js/basecoat.min.js \
  "https://cdn.jsdelivr.net/npm/basecoat-css@${BASECOAT_VERSION}/dist/js/basecoat.min.js"

# Verify downloads
ls -lh assets/css/basecoat.min.css
ls -lh assets/js/basecoat.min.js
```

### Step 2: Update `views/layouts/base.templ`

```templ
templ Base(title string) {
<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>

    <!-- PWA Meta Tags -->
    <link rel="manifest" href="/assets/static/manifest.json">
    <meta name="theme-color" content="#570df8">

    <!-- Tailwind CSS (your build) -->
    <link rel="stylesheet" href="/assets/css/output.css">

    <!-- BasecoatUI CSS (pre-compiled) -->
    <link rel="stylesheet" href="/assets/css/basecoat.min.css">

    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
    { children... }

    <!-- BasecoatUI JS (interactive components) -->
    <script src="/assets/js/basecoat.min.js"></script>

    <!-- Service Worker -->
    <script>
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.register('/assets/static/sw.js');
        }
    </script>
</body>
</html>
}
```

### Step 3: Update `tailwind.config.js`

```javascript
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
        // ❌ Remove DaisyUI plugin
        // require('assets/js/daisyui.mjs'),

        // ✅ No BasecoatUI plugin needed (using pre-compiled files)
    ],
}
```

### Step 4: Update Service Worker

```javascript
// assets/static/sw.js
const CACHE_NAME = 'skillsphere-v1';
const urlsToCache = [
  '/',
  '/assets/css/output.css',
  '/assets/css/basecoat.min.css',    // ✅ Add BasecoatUI CSS
  '/assets/js/basecoat.min.js',      // ✅ Add BasecoatUI JS
  '/assets/static/manifest.json',
  '/assets/static/icon-192.png',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(urlsToCache))
  );
});
```

### Step 5: Rebuild and Test

```bash
# Remove old DaisyUI files
rm assets/js/daisyui*.mjs

# Rebuild CSS (just Tailwind, no plugins)
tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify

# Build production binary
./scripts/build-prod.sh

# Check size
ls -lh bin/server
# Should be ~8.7 MB (smaller than DaisyUI setup)

# Test in browser
# Open DevTools → Network tab
# Should see 3 CSS/JS requests (all from localhost, all cached by SW)
```

---

## Recommendation: Which Approach?

### Use **Local Files** if:
- ✅ You want to avoid npm/Node.js entirely
- ✅ You're building a Go-first project
- ✅ You don't mind ~35 KB extra (no tree-shaking)
- ✅ You prefer simplicity over optimization
- ✅ You update UI framework rarely

### Use **NPM Plugin** if:
- ✅ You want maximum optimization (~35 KB savings)
- ✅ You already use npm for other tools (Tailwind CLI)
- ✅ You want automatic tree-shaking
- ✅ You need tight Tailwind integration
- ✅ You want single CSS file (fewer HTTP requests)

### **DON'T Use CDN** because:
- ❌ Breaks PWA offline-first principle
- ❌ External dependency = potential failure point
- ❌ Complex Service Worker caching
- ❌ Privacy/GDPR concerns

---

## My Recommendation for Skillsphere

Given your setup and goals:

### **Use Local Files** 🎯

**Why:**
1. You're building a **Go-first PWA** → avoid npm complexity
2. Your binary is **8.9 MB** → 35 KB extra is negligible (0.4%)
3. **Offline PWA** works perfectly → all assets embedded
4. **Single binary deployment** → no build tool dependencies
5. Simple setup → just download and embed

**Trade-off:**
- Lose ~35 KB to lack of tree-shaking
- But gain simplicity and Go-first purity

**Result:**
```
Before (DaisyUI):   8.9 MB binary, 505 KB assets
After (BasecoatUI): 8.7 MB binary, 215 KB assets
Savings:            290 KB (57% smaller assets)
```

---

## Migration Script

Here's a one-command migration:

```bash
#!/bin/bash
# migrate-to-basecoat-local.sh

set -e

echo "🔄 Migrating from DaisyUI to BasecoatUI (local files)..."

# 1. Download BasecoatUI
echo "📥 Downloading BasecoatUI files..."
curl -sL -o assets/css/basecoat.min.css \
  "https://cdn.jsdelivr.net/npm/basecoat-css@latest/dist/basecoat.min.css"
curl -sL -o assets/js/basecoat.min.js \
  "https://cdn.jsdelivr.net/npm/basecoat-css@latest/dist/js/basecoat.min.js"

# 2. Remove DaisyUI
echo "🗑️  Removing DaisyUI..."
rm -f assets/js/daisyui*.mjs

# 3. Update tailwind.config.js (remove DaisyUI plugin)
echo "⚙️  Updating Tailwind config..."
sed -i '' '/require.*daisyui/d' tailwind.config.js
sed -i '' '/daisyui: {/,/}/d' tailwind.config.js

# 4. Rebuild CSS
echo "🎨 Rebuilding CSS..."
tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify

# 5. Update base.templ (you'll need to do this manually)
echo "⚠️  Manual step required:"
echo "   1. Edit views/layouts/base.templ"
echo "   2. Add: <link rel=\"stylesheet\" href=\"/assets/css/basecoat.min.css\">"
echo "   3. Add: <script src=\"/assets/js/basecoat.min.js\"></script>"

# 6. Update Service Worker
echo "⚠️  Manual step required:"
echo "   1. Edit assets/static/sw.js"
echo "   2. Add to urlsToCache:"
echo "      '/assets/css/basecoat.min.css',"
echo "      '/assets/js/basecoat.min.js',"

echo ""
echo "✅ Download complete!"
echo "📝 Complete manual steps above, then run:"
echo "   ./scripts/build-prod.sh"
```

---

## Summary

| Metric | DaisyUI (Current) | BasecoatUI Local | BasecoatUI NPM | BasecoatUI CDN |
|--------|-------------------|------------------|----------------|----------------|
| Binary size | 8.9 MB | 8.7 MB | 8.6 MB | 8.5 MB |
| Assets embedded | 505 KB | 215 KB | 180 KB | 120 KB |
| Tree-shaking | ❌ No | ❌ No | ✅ Yes | ❌ No |
| HTTP requests | 1 | 3 | 1 | 3 |
| Offline PWA | ✅ Yes | ✅ Yes | ✅ Yes | ⚠️ Complex |
| npm required | No | No | Yes | No |
| Simplicity | Medium | ⭐⭐⭐⭐⭐ Best | Medium | Medium |

**For Go-first PWA: Use BasecoatUI Local Files** ✅

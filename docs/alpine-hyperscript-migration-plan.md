# Alpine.js & Hyperscript Migration Plan

## Executive Summary

After analyzing all JavaScript in your templ files, I've identified **4 distinct JavaScript patterns** that can be significantly improved using Alpine.js and Hyperscript, following the **Locality of Behavior (LoB)** philosophy.

**Current State:**
- 5 inline `<script>` blocks with vanilla JavaScript
- Mix of event handlers, state management, and DOM manipulation
- Logic scattered between inline onclick handlers and script blocks
- Some code duplication (password toggle logic)

**Recommended Approach:**
- Use **Alpine.js** for stateful components (theme toggle, dropdowns)
- Use **Hyperscript** for simple event-driven actions (password visibility, flash messages)

---

## JavaScript Inventory

### 1. **Navbar Component** (`views/components/navbar.templ`)

**Current JavaScript (Lines 89-142):**
```javascript
// Theme toggle with localStorage
// Dropdown menu with close-on-outside-click
// Global functions: toggleTheme(), toggleDropdown()
```

**Analysis:**
- **State Management**: Theme (dark/light), Dropdown (open/closed)
- **Events**: Click handlers, outside click detection
- **DOM Manipulation**: Class toggling, localStorage access
- **Scope**: Global functions polluting namespace

**LoB Violations:**
- Logic separated from HTML elements
- Global functions require finding script block to understand behavior
- State not co-located with UI

---

### 2. **Flash Message Component** (`views/components/flash_message.templ`)

**Current JavaScript (Lines 19-32):**
```javascript
// Auto-dismiss after 5 seconds with fade animation
// IIFE that runs on page load
```

**Analysis:**
- **Events**: Timer-based auto-dismiss
- **DOM Manipulation**: Style changes, element removal
- **Behavior**: One-time execution on load

**LoB Violations:**
- Timer logic not visible on the element
- Animation timing buried in script block

---

### 3. **Service Worker Registration** (`views/layouts/base.templ`)

**Current JavaScript (Lines 30-38):**
```javascript
// PWA service worker registration
if ('serviceWorker' in navigator) { ... }
```

**Analysis:**
- **Special Case**: This is fine as-is
- **Reason**: Service worker registration is app-level infrastructure, not UI behavior
- **Recommendation**: Keep unchanged

---

### 4. **Password Toggle** (`views/pages/auth/login.templ` & `register.templ`)

**Current JavaScript (Lines 174-191 in login, 208-225 in register):**
```javascript
function togglePassword(btn) {
  // Toggle input type and icon visibility
}
```

**Analysis:**
- **State**: Password visible/hidden
- **Events**: Button click
- **DOM Manipulation**: Input type, icon visibility
- **Issue**: **DUPLICATED** in two files

**LoB Violations:**
- State not tracked on element
- Function defined globally in script block
- Code duplication across pages

---

## Migration Strategy

### Priority 1: Password Toggle (Highest Impact)
**Why First:** Duplicated code, simple behavior, high LoB benefit

**Current:**
```html
<button onclick="togglePassword(this)">...</button>
<script>
  function togglePassword(btn) { ... }
</script>
```

**Recommended: Hyperscript** (Best fit for simple toggle action)
```html
<div _="
  init set $passwordVisible to false
  on click from <button/>
    if $passwordVisible
      set <input/>'s @type to 'password'
      add .hidden to <.eye-closed/>
      remove .hidden from <.eye-open/>
      set $passwordVisible to false
    else
      set <input/>'s @type to 'text'
      add .hidden to <.eye-open/>
      remove .hidden from <.eye-closed/>
      set $passwordVisible to true
    end
">
  <input type="password" />
  <button>
    <svg class="eye-open">...</svg>
    <svg class="eye-closed hidden">...</svg>
  </button>
</div>
```

**Alternative: Alpine.js** (More explicit for maintainers unfamiliar with Hyperscript)
```html
<div x-data="{ show: false }" data-password-toggle>
  <input :type="show ? 'text' : 'password'" />
  <button @click="show = !show">
    <svg class="eye-open" :class="{ 'hidden': show }">...</svg>
    <svg class="eye-closed" :class="{ 'hidden': !show }">...</svg>
  </button>
</div>
```

**Benefits:**
- ✅ Eliminates code duplication
- ✅ State visible on element
- ✅ No global functions
- ✅ Behavior co-located with markup

---

### Priority 2: Dropdown Menu (High Impact)
**Why Second:** Complex state, multiple interactions, global event listeners

**Current Issues:**
- Global `toggleDropdown()` function
- Document-level click listener
- Manual state tracking across all dropdowns

**Recommended: Alpine.js** (Complex state + reactivity)
```html
<div x-data="{ open: false }" @click.away="open = false" data-dropdown>
  <button @click="open = !open">
    <span>{ userName }</span>
    <svg :class="{ 'rotate-180': open }">...</svg>
  </button>

  <div x-show="open"
       x-transition
       class="dropdown-menu">
    <a href="/profile">Profile</a>
    <a href="/settings">Settings</a>
    <form action="/logout" method="POST">
      <button type="submit">Sign Out</button>
    </form>
  </div>
</div>
```

**Benefits:**
- ✅ Automatic close-on-outside-click with `@click.away`
- ✅ Built-in transitions with `x-transition`
- ✅ No manual event listener cleanup
- ✅ State scoped to component
- ✅ Arrow rotation tied to state reactively

**Why Alpine.js over Hyperscript:**
- Multiple UI elements reacting to same state (arrow, menu)
- Need for click-away behavior (Alpine has this built-in)
- Transition support

---

### Priority 3: Theme Toggle (Medium Impact)
**Why Third:** Global state, localStorage integration, affects entire page

**Current Issues:**
- IIFE runs on page load
- Global `toggleTheme()` function
- Manual localStorage manipulation

**Recommended: Alpine.js** (State + localStorage + global scope)
```html
<div x-data="themeToggle()" class="...">
  <button @click="toggle" aria-label="Toggle theme">
    <svg x-show="!isDark">☀️ Light</svg>
    <svg x-show="isDark">🌙 Dark</svg>
  </button>
</div>

<script>
function themeToggle() {
  return {
    isDark: localStorage.getItem('theme') === 'dark' ||
            !localStorage.getItem('theme'),

    init() {
      // Apply theme on load
      this.$watch('isDark', value => {
        document.documentElement.classList.toggle('dark', value);
        localStorage.setItem('theme', value ? 'dark' : 'light');
      });
      // Trigger initial application
      document.documentElement.classList.toggle('dark', this.isDark);
    },

    toggle() {
      this.isDark = !this.isDark;
    }
  }
}
</script>
```

**Benefits:**
- ✅ State co-located with component
- ✅ Automatic localStorage sync via watcher
- ✅ Reactive icon display
- ✅ Still accessible globally via Alpine store if needed

**Alternative: Alpine Store** (If you need theme access in multiple components)
```html
<!-- In base.templ -->
<script>
document.addEventListener('alpine:init', () => {
  Alpine.store('theme', {
    dark: localStorage.getItem('theme') === 'dark',
    toggle() {
      this.dark = !this.dark;
      localStorage.setItem('theme', this.dark ? 'dark' : 'light');
      document.documentElement.classList.toggle('dark', this.dark);
    }
  });
});
</script>

<!-- In navbar.templ -->
<button @click="$store.theme.toggle()">
  <svg x-show="!$store.theme.dark">☀️</svg>
  <svg x-show="$store.theme.dark">🌙</svg>
</button>
```

**Why Alpine.js over Hyperscript:**
- Need persistent state (localStorage)
- Multiple UI elements react to theme
- Watcher pattern for side effects

---

### Priority 4: Flash Message Auto-Dismiss (Low Impact)
**Why Last:** Simple, works well, low complexity

**Current:**
Works fine but logic not visible on element

**Recommended: Hyperscript** (Simple timer + animation)
```html
<div
  data-flash-message
  class="flash-message"
  _="on load wait 5s then transition opacity to 0 over 300ms
     then transition transform to 'translateY(-10px)' over 300ms
     then remove me"
>
  <span>{ message }</span>
  <button _="on click remove closest <[data-flash-message]/>">×</button>
</div>
```

**Benefits:**
- ✅ Auto-dismiss logic visible on element
- ✅ No separate script block needed
- ✅ Manual dismiss with one-liner

**Alternative: Alpine.js with x-init**
```html
<div
  x-data="{ show: true }"
  x-show="show"
  x-transition.opacity.duration.300ms
  x-init="setTimeout(() => show = false, 5000)"
  data-flash-message
>
  <span>{ message }</span>
  <button @click="show = false">×</button>
</div>
```

**Why Hyperscript preferred here:**
- More readable for simple timeline of events
- "wait, then, then, then" reads like plain English
- No need for state variable

---

## Decision Matrix

| Feature | Current | Alpine.js | Hyperscript | Recommendation | Reason |
|---------|---------|-----------|-------------|----------------|---------|
| **Password Toggle** | ❌ Script + onclick | ✅ Good | ✅ Good | **Hyperscript** | Simple state, action-focused, less verbose |
| **Dropdown Menu** | ❌ Script + listener | ✅✅ Excellent | ⚠️ Verbose | **Alpine.js** | Complex state, click-away, transitions |
| **Theme Toggle** | ❌ Script + IIFE | ✅✅ Excellent | ❌ Not suitable | **Alpine.js** | localStorage, global state, watchers |
| **Flash Messages** | ⚠️ Works but hidden | ✅ Good | ✅✅ Excellent | **Hyperscript** | Timeline of events, self-documenting |
| **Service Worker** | ✅ Fine as-is | ❌ Overkill | ❌ Overkill | **Keep current** | Infrastructure, not UI behavior |

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
**Goal:** Set up libraries and create reusable components

1. ✅ **Libraries already included** in base.templ:
   - Alpine.js
   - Hyperscript
   - HTMX

2. **Create shared Alpine components** (`assets/js/components.js`):
   ```javascript
   // Theme toggle store
   document.addEventListener('alpine:init', () => {
     Alpine.store('theme', { /* ... */ });
   });

   // Reusable component functions
   window.passwordToggle = () => ({ /* ... */ });
   window.dropdown = () => ({ /* ... */ });
   ```

3. **Test in isolation:**
   - Create a test page with each component
   - Verify behavior matches current implementation

### Phase 2: Password Toggle Migration (Week 1-2)
**Goal:** Eliminate duplicated password toggle code

1. Migrate `login.templ` password toggle to Hyperscript
2. Migrate `register.templ` password toggle to Hyperscript
3. Remove duplicate script blocks
4. Test both pages thoroughly

**Success Criteria:**
- ✅ No JavaScript in `<script>` tags for password toggle
- ✅ Behavior identical to current implementation
- ✅ Zero code duplication

### Phase 3: Dropdown Migration (Week 2)
**Goal:** Simplify dropdown with Alpine.js

1. Replace navbar dropdown JavaScript
2. Remove global event listeners
3. Test dropdown open/close behavior
4. Test click-away functionality

**Success Criteria:**
- ✅ No global `toggleDropdown()` function
- ✅ No document-level event listeners
- ✅ Dropdown state visible in markup

### Phase 4: Theme Toggle Migration (Week 2-3)
**Goal:** Make theme state reactive and visible

1. Create Alpine.js theme store
2. Migrate theme toggle in navbar
3. Remove IIFE and global function
4. Test localStorage persistence

**Success Criteria:**
- ✅ Theme state accessible in devtools
- ✅ No IIFE on page load
- ✅ localStorage sync automatic

### Phase 5: Flash Messages Migration (Week 3)
**Goal:** Make auto-dismiss behavior self-documenting

1. Replace IIFE with Hyperscript
2. Test auto-dismiss timing
3. Test manual dismiss

**Success Criteria:**
- ✅ Dismiss logic visible on element
- ✅ No script block in component

### Phase 6: Testing & Refinement (Week 3-4)
**Goal:** Ensure stability and document patterns

1. Cross-browser testing
2. Performance comparison
3. Create component documentation
4. Add to style guide

---

## Code Size Comparison

### Before Migration
```
navbar.templ:        54 lines of JavaScript
flash_message.templ: 14 lines of JavaScript
login.templ:         17 lines of JavaScript
register.templ:      17 lines of JavaScript
-------------------------------------------
Total:               102 lines of script blocks
```

### After Migration
```
navbar.templ:        ~8 lines of Alpine.js attributes
flash_message.templ: ~3 lines of Hyperscript
login.templ:         ~5 lines of Alpine/Hyperscript
register.templ:      ~5 lines of Alpine/Hyperscript
components.js:       ~50 lines (shared, reusable)
-------------------------------------------
Total markup:        21 lines (80% reduction in script blocks)
Shared code:         50 lines (reusable across app)
```

**Net benefit:**
- 80% reduction in inline scripts
- Better reusability
- Improved maintainability
- LoB principles followed

---

## Risks & Mitigation

### Risk 1: Learning Curve
**Risk:** Team unfamiliar with Alpine.js/Hyperscript syntax
**Mitigation:**
- Start with documentation and examples
- Migrate one component at a time
- Pair programming for first migrations
- Create internal cheat sheet

### Risk 2: Debugging Complexity
**Risk:** Harder to debug inline Alpine/Hyperscript vs scripts
**Mitigation:**
- Use Alpine DevTools browser extension
- Add `x-data` to make component boundaries visible
- Document common patterns

### Risk 3: Bundle Size
**Risk:** Adding frameworks increases page weight
**Mitigation:**
- Alpine.js: 15KB gzipped (already included)
- Hyperscript: 10KB gzipped (already included)
- Both are already loaded, no additional cost

### Risk 4: Breaking Changes
**Risk:** Migration introduces bugs
**Mitigation:**
- Migrate one component per PR
- Add automated tests for each component
- Keep old code commented temporarily
- Use feature flags for gradual rollout

---

## Success Metrics

### Quantitative
- [ ] 80%+ reduction in `<script>` block lines
- [ ] 0 duplicate JavaScript functions
- [ ] 0 global function pollution
- [ ] Page load time unchanged or improved
- [ ] Bundle size increase < 5KB

### Qualitative
- [ ] New developers can understand behavior without finding scripts
- [ ] Component state visible in browser devtools
- [ ] No "action at a distance" - behavior co-located with markup
- [ ] Code review time reduced

---

## Recommended Reading

1. **Alpine.js Docs:** https://alpinejs.dev/
   - Focus on: `x-data`, `x-show`, `@click`, `x-transition`

2. **Hyperscript Docs:** https://hyperscript.org/
   - Focus on: event handlers, `wait`, `then`, `remove`

3. **LoB Philosophy:** https://htmx.org/essays/locality-of-behaviour/
   - Understanding the "why" behind this approach

---

## Next Steps

1. **Review this plan** with team
2. **Prioritize phases** based on business needs
3. **Assign owners** for each migration phase
4. **Create tickets** in project management tool
5. **Set up Alpine DevTools** for debugging
6. **Schedule kickoff meeting** to discuss approach

---

## Questions for Discussion

1. Do we want to use Alpine.js or Hyperscript for password toggle? (I recommend Hyperscript)
2. Should theme be an Alpine Store (global) or component-scoped?
3. Do we want to keep HTMX for dynamic content loading? (Yes, it complements these tools)
4. What's our timeline for completing this migration?
5. Do we want to create a component library file (`components.js`)?

---

## Conclusion

This migration will significantly improve code maintainability by following LoB principles. Each UI behavior will be self-documenting and co-located with its markup, reducing cognitive load and making the codebase more approachable for new developers.

**Primary Benefits:**
1. **Locality of Behavior** - Logic lives on elements
2. **No Code Duplication** - Shared patterns extracted
3. **Better Developer Experience** - State visible in devtools
4. **Smaller Script Blocks** - 80% reduction
5. **Improved Maintainability** - Easier to understand and modify

**Estimated Effort:** 2-3 weeks with proper testing
**Risk Level:** Low (can roll back component-by-component)
**ROI:** High (long-term maintainability gains)

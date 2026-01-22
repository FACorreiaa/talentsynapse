# Hyperscript to Surreal Migration Plan

**Date Created:** January 22, 2026
**Status:** Planning Phase
**Estimated Effort:** 2-3 hours

---

## Executive Summary

This document outlines the plan to migrate from Hyperscript to Surreal.js for client-side interactivity. Surreal is a tiny jQuery alternative (320 lines, no dependencies) that provides better ergonomics for vanilla JavaScript with inline Locality of Behavior.

**Current State:**
- Hyperscript is loaded (~10KB gzipped) but only used in 3 locations
- 3 `_=` directive usages across 3 template files
- Surreal.js already downloaded (16KB file) but not yet integrated

**Goal:**
- Replace all hyperscript directives with Surreal.js inline `<script>` tags
- Remove hyperscript.min.js dependency from base.templ
- Maintain 100% Locality of Behavior (LoB) compliance

---

## Current Hyperscript Usage Analysis

### Location 1: Navbar Dropdown Menu
**File:** `internal/app/views/components/navbar.templ:23-25`

**Current Code:**
```html
<button
    class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-muted hover:bg-muted/80 transition-all duration-300"
    _="on click toggle .hidden on #dropdown-menu
       on click toggle .rotate-180 on #dropdown-chevron
       on click elsewhere add .hidden to #dropdown-menu then remove .rotate-180 from #dropdown-chevron"
>
```

**Functionality:**
1. Toggle `.hidden` class on `#dropdown-menu`
2. Toggle `.rotate-180` class on `#dropdown-chevron`
3. On click elsewhere: close dropdown and reset chevron rotation

**Complexity:** Medium (requires click-away handler)

---

### Location 2: Theme Toggle Button
**File:** `internal/app/views/components/navbar.templ:66-73`

**Current Code:**
```html
<button
    class="btn btn-sm btn-icon btn-ghost"
    aria-label="Toggle theme"
    _="on click
       js
         document.documentElement.classList.toggle('dark');
         const isDark = document.documentElement.classList.contains('dark');
         localStorage.setItem('theme', isDark ? 'dark' : 'light');
       end
       toggle .hidden on .sun-icon in me
       toggle .hidden on .moon-icon in me"
>
```

**Functionality:**
1. Toggle `dark` class on `<html>` element
2. Save theme preference to localStorage
3. Toggle icon visibility (sun/moon)

**Complexity:** Medium (involves localStorage and multiple element toggles)

---

### Location 3: Password Toggle (Login Page)
**File:** `internal/app/views/pages/auth/login.templ:99-102`

**Current Code:**
```html
<button
    type="button"
    _="on click
        toggle the *type of #password between 'password' and 'text'
        toggle .hidden on .eye-open in me
        toggle .hidden on .eye-closed in me"
    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
>
```

**Functionality:**
1. Toggle input type between 'password' and 'text'
2. Toggle visibility of eye-open icon
3. Toggle visibility of eye-closed icon

**Complexity:** Simple

---

### Location 4: Password Toggle (Register Page)
**File:** `internal/app/views/pages/auth/register.templ:84-87`

**Current Code:**
```html
<button
    type="button"
    _="on click
        toggle the *type of #password between 'password' and 'text'
        toggle .hidden on .eye-open in me
        toggle .hidden on .eye-closed in me"
    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
>
```

**Functionality:** Identical to Location 3

**Complexity:** Simple

---

### Location 5: Flash Messages (Potential)
**File:** `internal/app/views/components/flash_message.templ`

**Current Code:** No hyperscript directives found currently

**Notes:**
- The migration-completed.md document mentions flash messages should use hyperscript
- Current implementation uses `data-flash-message` attributes but no `_=` directives
- Flash message auto-dismiss may be handled by external JavaScript or not implemented yet

---

## Surreal.js Capabilities

### Core Features Needed for Migration

1. **Event Handling**
   - `me().on("click", ev => { ... })`
   - Clean, chainable event handlers

2. **Class Manipulation**
   - `me().classAdd('hidden')`
   - `me().classRemove('hidden')`
   - `me().classToggle('hidden')`

3. **Attribute Manipulation**
   - `me().attribute('type', 'text')`
   - Perfect for password toggle

4. **DOM Selection**
   - `me()` - get parent element of script (Locality of Behavior)
   - `me('#id')` - select by ID
   - `any('.class')` - select all matching elements

5. **Global Conveniences**
   - `sleep(ms)` - async setTimeout
   - `tick()` - await requestAnimationFrame
   - `halt(event)` - stopPropagation + preventDefault

---

## Migration Strategy by Component

### Priority 1: Password Toggle (Both Login & Register)
**Reason:** Simplest migration, duplicated code, highest confidence

**Hyperscript (Current):**
```html
<button
    type="button"
    _="on click
        toggle the *type of #password between 'password' and 'text'
        toggle .hidden on .eye-open in me
        toggle .hidden on .eye-closed in me"
>
    <svg class="w-5 h-5 eye-open">...</svg>
    <svg class="w-5 h-5 eye-closed hidden">...</svg>
</button>
```

**Surreal (Proposed):**
```html
<button type="button" class="...">
    <svg class="w-5 h-5 eye-open">...</svg>
    <svg class="w-5 h-5 eye-closed hidden">...</svg>
    <script>
        me().on("click", ev => {
            const input = me('#password')
            const isPassword = input.attribute('type') === 'password'
            input.attribute('type', isPassword ? 'text' : 'password')
            me('.eye-open', me(ev)).classToggle('hidden')
            me('.eye-closed', me(ev)).classToggle('hidden')
        })
    </script>
</button>
```

**Benefits:**
- Eliminates code duplication (same script works in both pages)
- Clear, readable vanilla JavaScript
- Maintains LoB (script inside button element)

---

### Priority 2: Theme Toggle
**Reason:** Medium complexity, global state management

**Hyperscript (Current):**
```html
<button
    aria-label="Toggle theme"
    _="on click
       js
         document.documentElement.classList.toggle('dark');
         const isDark = document.documentElement.classList.contains('dark');
         localStorage.setItem('theme', isDark ? 'dark' : 'light');
       end
       toggle .hidden on .sun-icon in me
       toggle .hidden on .moon-icon in me"
>
    <svg class="h-5 w-5 sun-icon">...</svg>
    <svg class="h-5 w-5 moon-icon hidden">...</svg>
</button>
```

**Surreal (Proposed):**
```html
<button aria-label="Toggle theme" class="btn btn-sm btn-icon btn-ghost">
    <svg class="h-5 w-5 sun-icon">...</svg>
    <svg class="h-5 w-5 moon-icon hidden">...</svg>
    <script>
        me().on("click", ev => {
            document.documentElement.classList.toggle('dark')
            const isDark = document.documentElement.classList.contains('dark')
            localStorage.setItem('theme', isDark ? 'dark' : 'light')

            me('.sun-icon', me(ev)).classToggle('hidden')
            me('.moon-icon', me(ev)).classToggle('hidden')
        })
    </script>
</button>
```

**Benefits:**
- More readable than hyperscript's `js...end` blocks
- Direct JavaScript, no special syntax
- Maintains LoB

---

### Priority 3: Dropdown Menu
**Reason:** Most complex, requires click-away handler

**Hyperscript (Current):**
```html
<button
    _="on click toggle .hidden on #dropdown-menu
       on click toggle .rotate-180 on #dropdown-chevron
       on click elsewhere add .hidden to #dropdown-menu then remove .rotate-180 from #dropdown-chevron"
>
```

**Surreal (Proposed - Option A: Inline):**
```html
<div class="relative ml-2" id="user-dropdown">
    <button class="flex items-center gap-2...">
        <!-- avatar/user info -->
        <svg id="dropdown-chevron" class="w-4 h-4...">...</svg>
        <script>
            me().on("click", ev => {
                me('#dropdown-menu').classToggle('hidden')
                me('#dropdown-chevron').classToggle('rotate-180')
            })
        </script>
    </button>
    <div id="dropdown-menu" class="hidden absolute right-0...">
        <!-- menu items -->
    </div>
    <script>
        // Click away handler
        document.addEventListener('click', (e) => {
            const dropdown = me('#user-dropdown')
            if (!dropdown.contains(e.target)) {
                me('#dropdown-menu').classAdd('hidden')
                me('#dropdown-chevron').classRemove('rotate-180')
            }
        })
    </script>
</div>
```

**Surreal (Proposed - Option B: Shared Function):**
```html
<div class="relative ml-2" id="user-dropdown">
    <button class="flex items-center gap-2...">
        <!-- avatar/user info -->
        <svg id="dropdown-chevron" class="w-4 h-4...">...</svg>
        <script>
            me().on("click", ev => {
                halt(ev) // Prevent event bubbling
                me('#dropdown-menu').classToggle('hidden')
                me('#dropdown-chevron').classToggle('rotate-180')
            })
        </script>
    </button>
    <div id="dropdown-menu" class="hidden absolute right-0...">
        <!-- menu items -->
    </div>
</div>

<!-- In base.templ or shared script -->
<script>
    // Global click-away handler for all dropdowns
    document.addEventListener('click', (e) => {
        if (!e.target.closest('#user-dropdown')) {
            me('#dropdown-menu')?.classAdd('hidden')
            me('#dropdown-chevron')?.classRemove('rotate-180')
        }
    })
</script>
```

**Benefits:**
- More explicit than hyperscript's "elsewhere"
- Can be extracted to reusable pattern
- Better debugging capabilities

**Challenges:**
- Hyperscript's "elsewhere" is very ergonomic
- Surreal requires manual click-away implementation
- Potential for multiple event listeners if not careful

---

### Priority 4: Flash Messages (If Needed)
**Status:** Currently no hyperscript implementation found

**If implementing auto-dismiss:**

**Surreal (Proposed):**
```html
<div class="flash-message" data-flash-message>
    <div class="flex items-center gap-3">
        <!-- icon and message -->
        <button data-flash-close class="p-1...">
            <svg class="w-4 h-4">...</svg>
            <script>
                me().on("click", ev => {
                    me('[data-flash-message]', document).remove()
                })
            </script>
        </button>
    </div>
    <script>
        (async () => {
            await sleep(5000) // Wait 5 seconds
            const flash = me('[data-flash-message]', document)
            flash.styles({ opacity: '0', transition: 'opacity 300ms' })
            await sleep(300)
            flash.styles({ transform: 'translateY(-10px)', transition: 'transform 300ms' })
            await sleep(300)
            flash.remove()
        })()
    </script>
</div>
```

**Benefits:**
- Clear timeline of events using async/await
- Uses Surreal's `sleep()` helper
- Maintainable animation sequence

---

## Migration Checklist

### Phase 1: Preparation
- [x] Audit all hyperscript usage locations
- [x] Document current functionality
- [x] Download Surreal.js (already done)
- [ ] Add Surreal.js to base.templ
- [ ] Test Surreal.js loads correctly
- [ ] Verify no conflicts with HTMX or Basecoat

### Phase 2: Password Toggle Migration
**Files:** `login.templ`, `register.templ`

- [ ] Migrate login.templ password toggle
- [ ] Test login page password visibility toggle
- [ ] Migrate register.templ password toggle
- [ ] Test register page password visibility toggle
- [ ] Verify icons switch correctly
- [ ] Cross-browser test (Chrome, Firefox, Safari)

### Phase 3: Theme Toggle Migration
**Files:** `navbar.templ`

- [ ] Migrate theme toggle to Surreal
- [ ] Test theme toggle functionality
- [ ] Verify localStorage persistence
- [ ] Test on page reload
- [ ] Verify icons toggle correctly
- [ ] Test with browser DevTools

### Phase 4: Dropdown Menu Migration
**Files:** `navbar.templ`

- [ ] Migrate dropdown toggle to Surreal
- [ ] Implement click-away handler
- [ ] Test dropdown open/close
- [ ] Test click-away functionality
- [ ] Verify chevron rotation
- [ ] Test with multiple rapid clicks

### Phase 5: Cleanup
- [ ] Remove hyperscript.min.js from base.templ
- [ ] Remove hyperscript.min.js file from assets
- [ ] Update build scripts if needed
- [ ] Update documentation
- [ ] Create PR with all changes

### Phase 6: Testing
- [ ] Full regression test all pages
- [ ] Test on mobile devices
- [ ] Test offline PWA functionality
- [ ] Performance comparison (before/after)
- [ ] Verify no console errors
- [ ] Test with different screen sizes

---

## Code Size Comparison

### Before Migration
```
Libraries:
- hyperscript.min.js:    ~10 KB (gzipped)
- Total:                 ~10 KB

Inline Directives:
- navbar.templ:          3 _= directives (~150 chars)
- login.templ:           1 _= directive (~100 chars)
- register.templ:        1 _= directive (~100 chars)
- Total:                 ~350 characters inline
```

### After Migration
```
Libraries:
- surreal.min.js:        ~16 KB raw (~5 KB gzipped)
- Total:                 ~5 KB (50% smaller!)

Inline Scripts:
- navbar.templ:          ~30 lines of Surreal JS
- login.templ:           ~8 lines of Surreal JS
- register.templ:        ~8 lines of Surreal JS
- Total:                 ~46 lines of vanilla JS
```

**Net Impact:**
- Library size: -5 KB (50% reduction)
- Inline code: +~40 lines (but more readable)
- Overall: Smaller bundle, more explicit code

---

## Hyperscript vs Surreal Syntax Comparison

### Event Handling
```javascript
// Hyperscript
_="on click toggle .hidden on #menu"

// Surreal
me().on("click", ev => {
    me('#menu').classToggle('hidden')
})
```

### Class Manipulation
```javascript
// Hyperscript
_="toggle .hidden on .icon in me"

// Surreal
me('.icon', me()).classToggle('hidden')
```

### Attribute Changes
```javascript
// Hyperscript
_="toggle the *type of #password between 'password' and 'text'"

// Surreal
const input = me('#password')
input.attribute('type', input.attribute('type') === 'password' ? 'text' : 'password')
```

### Click Elsewhere
```javascript
// Hyperscript
_="on click elsewhere add .hidden to #menu"

// Surreal
document.addEventListener('click', (e) => {
    if (!e.target.closest('#container')) {
        me('#menu').classAdd('hidden')
    }
})
```

### Animations/Delays
```javascript
// Hyperscript
_="on load wait 5s then transition opacity to 0 over 300ms then remove me"

// Surreal
(async () => {
    await sleep(5000)
    me().styles({ opacity: '0', transition: 'opacity 300ms' })
    await sleep(300)
    me().remove()
})()
```

---

## Risks & Mitigations

### Risk 1: Surreal.js Learning Curve
**Risk:** Team unfamiliar with Surreal.js syntax
**Mitigation:**
- Surreal uses vanilla JS patterns (easier than hyperscript syntax)
- Good documentation at https://gnat.github.io/surreal/
- More searchable/debuggable than hyperscript

### Risk 2: Click-Away Implementation
**Risk:** Manual click-away is more complex than hyperscript "elsewhere"
**Mitigation:**
- Create reusable pattern/helper function
- Document the pattern clearly
- Test thoroughly across browsers

### Risk 3: Bundle Size Increase
**Risk:** Surreal.js is 16KB vs hyperscript's 10KB raw
**Mitigation:**
- Gzipped: Surreal is ~5KB vs hyperscript ~10KB (actually smaller!)
- Better ergonomics justify any small size difference

### Risk 4: Breaking Changes
**Risk:** Migration introduces bugs
**Mitigation:**
- Migrate one component at a time
- Test each component thoroughly
- Keep hyperscript until all migrations complete
- Can roll back component-by-component

### Risk 5: HTMX/Basecoat Conflicts
**Risk:** Surreal may conflict with existing libraries
**Mitigation:**
- All three libraries are designed to coexist
- Surreal doesn't use global $ (unlike jQuery)
- Test in development environment first

---

## Testing Strategy

### Unit Testing (Manual)
1. **Password Toggle**
   - Click eye icon → input type changes
   - Click again → type reverts
   - Icons switch visibility correctly

2. **Theme Toggle**
   - Click theme button → dark mode toggles
   - Reload page → theme persists
   - Icons switch correctly

3. **Dropdown Menu**
   - Click avatar → dropdown opens
   - Click outside → dropdown closes
   - Chevron rotates correctly
   - Multiple rapid clicks handled properly

### Integration Testing
1. **Page Load**
   - Surreal.js loads before inline scripts
   - No console errors on any page
   - All functionality works on first visit

2. **Navigation**
   - Test all pages with Surreal components
   - Verify no conflicts with HTMX page transitions
   - Flash messages work after navigation

3. **Offline PWA**
   - Service worker caches surreal.min.js
   - All functionality works offline
   - No network errors in DevTools

### Cross-Browser Testing
- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Mobile Safari (iOS)
- Chrome Android

---

## Implementation Timeline

### Day 1: Setup & Password Toggle (1-2 hours)
1. Add surreal.min.js to base.templ (15 min)
2. Test Surreal loads correctly (15 min)
3. Migrate login.templ password toggle (20 min)
4. Test login page (15 min)
5. Migrate register.templ password toggle (15 min)
6. Test register page (15 min)

### Day 2: Theme Toggle & Dropdown (1-2 hours)
1. Migrate theme toggle (30 min)
2. Test theme persistence (15 min)
3. Migrate dropdown menu (45 min)
4. Test dropdown + click-away (30 min)

### Day 3: Cleanup & Testing (1 hour)
1. Remove hyperscript.min.js (15 min)
2. Full regression testing (30 min)
3. Documentation updates (15 min)

**Total Estimated Time:** 2-3 hours

---

## Rollback Plan

If critical issues arise:

### Option 1: Immediate Rollback
1. Re-add hyperscript.min.js to base.templ
2. Revert all component changes
3. Git revert if changes were committed

### Option 2: Selective Rollback
1. Keep surreal.min.js loaded
2. Revert specific component one at a time
3. Allows partial migration if some components work

### Option 3: Hybrid Approach
1. Keep both libraries loaded temporarily
2. Migrate components one at a time
3. Remove hyperscript only when 100% migrated

---

## Success Metrics

### Quantitative
- [ ] 100% of hyperscript directives replaced
- [ ] Zero console errors
- [ ] Bundle size reduced by ~5 KB
- [ ] All tests passing

### Qualitative
- [ ] Code is more readable (vanilla JS vs hyperscript syntax)
- [ ] Easier to debug (standard JS tools work)
- [ ] Maintains Locality of Behavior
- [ ] Team can understand without hyperscript knowledge

---

## Documentation Updates Needed

1. **Update base.templ script loading order**
2. **Update PWA_JAVASCRIPT_DEPENDENCIES.md**
   - Remove hyperscript section
   - Add surreal section
3. **Update alpine-hyperscript-migration-plan.md**
   - Note that hyperscript has been replaced by Surreal
4. **Create surreal-patterns.md** (optional)
   - Document common Surreal patterns used in project
   - Click-away dropdown pattern
   - Toggle visibility pattern
   - Password toggle pattern

---

## Conclusion

Migrating from Hyperscript to Surreal.js will:

1. **Reduce bundle size** by ~50% (10KB → 5KB gzipped)
2. **Improve readability** (vanilla JS vs custom syntax)
3. **Enhance debuggability** (standard browser tools)
4. **Maintain LoB** (all behavior stays co-located)
5. **Remove learning curve** (vanilla JS vs hyperscript syntax)

**Recommendation:** Proceed with migration in three phases:
1. Password toggles (simplest, highest confidence)
2. Theme toggle (medium complexity)
3. Dropdown menu (most complex, requires careful testing)

**Risk Level:** Low
**Estimated Effort:** 2-3 hours
**ROI:** High (smaller bundle, better DX, more maintainable)

---

## Next Steps

1. Review this plan with team
2. Get approval to proceed
3. Create feature branch: `feature/migrate-hyperscript-to-surreal`
4. Start with Phase 1: Add Surreal to base.templ
5. Migrate components in priority order
6. Test thoroughly
7. Create PR for review

---

## Questions for Discussion

1. Should we keep both libraries during migration or remove hyperscript immediately?
2. Do we want to create reusable Surreal helper functions (e.g., for click-away)?
3. Should we migrate flash messages to Surreal or leave them as vanilla JS?
4. What's our browser support policy? (affects testing matrix)
5. Do we want to add Surreal examples to our component style guide?

---

## References

- **Surreal.js Documentation:** https://gnat.github.io/surreal/
- **Surreal GitHub:** https://github.com/gnat/surreal
- **Hyperscript Documentation:** https://hyperscript.org/
- **Project Migration Docs:** `docs/alpine-hyperscript-migration-plan.md`
- **Completed Migration:** `docs/migration-completed.md`

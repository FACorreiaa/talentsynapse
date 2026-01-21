# Testing Checklist for Alpine.js & Hyperscript Migration

## Password Toggle

### Login Page (`/login`)
- [ ] Eye icon visible next to password field
- [ ] Clicking eye icon changes input type from `password` to `text`
- [ ] Password becomes visible when type is `text`
- [ ] Eye icon changes from open to closed
- [ ] Clicking again toggles back
- [ ] No JavaScript errors in console

### Register Page (`/register`)
- [ ] Eye icon visible next to password field
- [ ] Same toggle behavior as login page
- [ ] Confirm password field (if added) works independently
- [ ] No JavaScript errors in console

**Expected Behavior:**
```
Click → Password visible → Icon changes
Click → Password hidden → Icon changes back
```

---

## Dropdown Menu (User Menu)

### When Authenticated
- [ ] User avatar/name button visible in navbar
- [ ] Clicking button opens dropdown menu
- [ ] Dropdown menu shows: Profile, Settings, Sign Out
- [ ] Arrow icon rotates 180° when open
- [ ] Clicking outside dropdown closes it automatically
- [ ] Smooth transition animation (fade in/out)
- [ ] Multiple clicks toggle correctly
- [ ] ESC key closes dropdown (native browser behavior)
- [ ] No JavaScript errors in console

### When Not Authenticated
- [ ] "Sign In" and "Get Started" buttons visible
- [ ] No dropdown menu present
- [ ] No JavaScript errors

**Expected Behavior:**
```
Click avatar → Dropdown opens → Arrow rotates
Click outside → Dropdown closes → Arrow resets
```

---

## Theme Toggle

### Initial Load
- [ ] Theme loads from localStorage
- [ ] If no preference: defaults to dark mode
- [ ] Correct icon shows (sun for dark, moon for light)
- [ ] `<html class="dark">` present for dark mode

### Toggle Functionality
- [ ] Clicking theme button toggles dark/light mode
- [ ] Icon changes immediately (sun ↔ moon)
- [ ] Page colors change immediately
- [ ] Theme saved to localStorage
- [ ] Refresh page: theme persists
- [ ] No JavaScript errors in console

### DevTools Check
- [ ] Open Alpine DevTools (if installed)
- [ ] Navigate to Stores tab
- [ ] See `theme` store with `dark: true/false`
- [ ] Toggle theme and watch store update

**Expected Behavior:**
```
Click → Dark mode OFF → Light colors → Sun icon → Save
Refresh → Light mode persists
Click → Dark mode ON → Dark colors → Moon icon → Save
```

---

## Flash Messages

### Success Message
- [ ] Green flash message appears (e.g., after successful registration)
- [ ] Message contains text and checkmark icon
- [ ] Close button (×) visible on right
- [ ] Clicking × removes message immediately
- [ ] If not clicked: auto-dismisses after 5 seconds
- [ ] Fade-out animation (opacity → 0)
- [ ] Slide-up animation (translateY)
- [ ] Message removed from DOM after animation
- [ ] No JavaScript errors

### Error Message
- [ ] Red flash message appears (e.g., login error)
- [ ] Message contains text and error icon
- [ ] Same dismiss behavior as success message

### Warning Message
- [ ] Yellow flash message appears
- [ ] Same dismiss behavior

**Expected Behavior:**
```
Load page → Flash appears → Wait 5s → Fade out → Slide up → Removed
OR
Load page → Flash appears → Click × → Immediately removed
```

---

## Service Worker (PWA)

- [ ] Console shows "✅ Service Worker registered"
- [ ] No errors during registration
- [ ] PWA install prompt available (if applicable)
- [ ] Offline functionality works (if implemented)

---

## Cross-Browser Testing

### Desktop Browsers
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)
- [ ] Edge (latest)

### Mobile Browsers
- [ ] iOS Safari
- [ ] Chrome on Android
- [ ] Firefox on Android

---

## Performance Testing

### Page Load
- [ ] No visible lag on page load
- [ ] Theme applies immediately (no flash of wrong theme)
- [ ] Dropdown initializes without delay
- [ ] No console errors or warnings

### Interactions
- [ ] Button clicks feel instant
- [ ] Animations smooth (60fps)
- [ ] No memory leaks (check DevTools Memory tab)
- [ ] Alpine DevTools shows reasonable component tree

---

## Console Checks

Open browser console and verify:

### No Errors
```
✓ No red error messages
✓ No "undefined is not a function"
✓ No "Cannot read property"
✓ No Alpine.js errors
✓ No Hyperscript errors
```

### Expected Messages
```
✅ Service Worker registered
(Optional) Alpine.js loaded messages
```

---

## Alpine DevTools (If Installed)

### Install Extension
1. Chrome: https://chrome.google.com/webstore/detail/alpine-devtools/
2. Firefox: https://addons.mozilla.org/en-US/firefox/addon/alpine-js-devtools/

### Check Components
- [ ] Open DevTools → Alpine tab
- [ ] See `dropdown()` component when user menu present
- [ ] Check `open: false/true` state when toggling
- [ ] See `theme` store in Stores tab
- [ ] Check `dark: false/true` when toggling

---

## Edge Cases

### Rapid Clicking
- [ ] Click password toggle 10 times rapidly → No errors
- [ ] Click dropdown 10 times rapidly → No errors
- [ ] Click theme toggle 10 times rapidly → No errors

### Multiple Flash Messages
- [ ] If multiple flash messages appear → All dismiss independently
- [ ] Stacked messages don't interfere with each other

### Navigation During Animations
- [ ] Navigate away while flash message is fading → No errors

### localStorage Issues
- [ ] Clear localStorage → Theme defaults to dark
- [ ] Set invalid theme value → Falls back to default
- [ ] localStorage disabled (private browsing) → Theme still works (just doesn't persist)

---

## Accessibility

### Keyboard Navigation
- [ ] Tab to theme button → Works
- [ ] Tab to dropdown button → Works
- [ ] Enter/Space on dropdown → Opens
- [ ] ESC closes dropdown
- [ ] Tab through dropdown items → Works

### Screen Readers
- [ ] Theme button has `aria-label="Toggle theme"`
- [ ] Dropdown menu accessible via keyboard
- [ ] Flash messages announced by screen reader

---

## Regression Tests

### Ensure Old Behavior Preserved
- [ ] Login still works
- [ ] Register still works
- [ ] Logout still works
- [ ] Navigation still works
- [ ] Forms still submit
- [ ] No broken links

---

## Known Issues / Expected Behavior

### Hyperscript Syntax
- Hyperscript uses `_=` attribute (not a typo!)
- Valid HTML5 attribute
- May look unfamiliar but works correctly

### Alpine.js Store Initialization
- Theme store loads on Alpine init
- Slight delay possible on very slow connections
- Theme may flash if not in localStorage (expected)

### Flash Message Transitions
- CSS transitions may not work in very old browsers
- Graceful degradation: message still dismisses, just no animation

---

## Testing Priority

### Priority 1 (Must Test)
1. ✅ Password toggle (most visible user-facing feature)
2. ✅ Theme toggle (affects entire site)
3. ✅ Flash messages (important for feedback)

### Priority 2 (Should Test)
4. ✅ Dropdown menu (only visible when authenticated)
5. ✅ Cross-browser compatibility

### Priority 3 (Nice to Have)
6. ✅ Edge cases and rapid clicking
7. ✅ Alpine DevTools inspection
8. ✅ Accessibility

---

## Test Report Template

```markdown
## Test Report

**Date:** YYYY-MM-DD
**Tester:** [Name]
**Browser:** [Chrome/Firefox/Safari] Version
**Device:** [Desktop/Mobile]

### Results
- Password Toggle: ✅ PASS / ❌ FAIL
- Dropdown Menu: ✅ PASS / ❌ FAIL
- Theme Toggle: ✅ PASS / ❌ FAIL
- Flash Messages: ✅ PASS / ❌ FAIL
- No Console Errors: ✅ PASS / ❌ FAIL

### Issues Found
1. [Description of issue if any]

### Notes
[Any additional observations]
```

---

## Automated Testing (Future)

Consider adding:
1. **Playwright tests** for end-to-end functionality
2. **Visual regression tests** for theme switching
3. **Unit tests** for Alpine.js store logic

Example test structure:
```javascript
test('password toggle works', async ({ page }) => {
  await page.goto('/login');
  await page.click('[data-password-toggle] button');
  const inputType = await page.getAttribute('#password', 'type');
  expect(inputType).toBe('text');
});
```

---

## Success Criteria

All items checked = Migration successful! ✅

**If any issues found:**
1. Document in GitHub issue
2. Prioritize fix based on impact
3. Rollback if critical (see rollback plan in migration-completed.md)

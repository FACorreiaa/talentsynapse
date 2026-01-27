// TalentSynapse Custom JavaScript
// Handles flash messages, mobile interactions, and UI enhancements

(function() {
    'use strict';

    // ═══════════════════════════════════════════════════════════════════════
    // Flash Message Handler
    // ═══════════════════════════════════════════════════════════════════════

    function initFlashMessages() {
        // Handle flash message auto-dismiss and close button
        document.querySelectorAll('[data-flash-message]').forEach(function(flash) {
            const duration = parseInt(flash.dataset.duration) || 5000;
            const closeBtn = flash.querySelector('[data-flash-close]');

            // Auto-dismiss after duration
            const timer = setTimeout(function() {
                dismissFlash(flash);
            }, duration);

            // Close button click handler
            if (closeBtn) {
                closeBtn.addEventListener('click', function(e) {
                    e.preventDefault();
                    e.stopPropagation();
                    clearTimeout(timer);
                    dismissFlash(flash);
                });

                // Ensure button is clickable on mobile
                closeBtn.style.touchAction = 'manipulation';
                closeBtn.style.cursor = 'pointer';
            }

            // Also allow clicking anywhere on the flash to dismiss
            flash.addEventListener('click', function(e) {
                // Don't dismiss if clicking on a link or button inside
                if (e.target.tagName !== 'A' && e.target.tagName !== 'BUTTON') {
                    clearTimeout(timer);
                    dismissFlash(flash);
                }
            });
        });
    }

    function dismissFlash(flashElement) {
        flashElement.style.animation = 'slideOutRight 0.3s ease-in-out';
        setTimeout(function() {
            flashElement.remove();
        }, 300);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Mobile Menu Toggle (if needed)
    // ═══════════════════════════════════════════════════════════════════════

    function initMobileMenu() {
        const menuToggle = document.querySelector('[data-mobile-menu-toggle]');
        const mobileMenu = document.querySelector('[data-mobile-menu]');

        if (menuToggle && mobileMenu) {
            menuToggle.addEventListener('click', function(e) {
                e.preventDefault();
                mobileMenu.classList.toggle('hidden');
            });
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Tabs Handler (for mobile-responsive tabs)
    // ═══════════════════════════════════════════════════════════════════════

    function initTabs() {
        document.querySelectorAll('[data-tabs]').forEach(function(tabContainer) {
            const tabs = tabContainer.querySelectorAll('[data-tab]');
            const panels = tabContainer.querySelectorAll('[data-tab-panel]');

            tabs.forEach(function(tab, index) {
                tab.addEventListener('click', function(e) {
                    e.preventDefault();

                    // Remove active state from all tabs
                    tabs.forEach(function(t) {
                        t.classList.remove('active');
                        t.setAttribute('aria-selected', 'false');
                    });

                    // Hide all panels
                    panels.forEach(function(p) {
                        p.classList.add('hidden');
                    });

                    // Activate clicked tab
                    tab.classList.add('active');
                    tab.setAttribute('aria-selected', 'true');

                    // Show corresponding panel
                    const panelId = tab.dataset.tab;
                    const panel = document.querySelector(`[data-tab-panel="${panelId}"]`);
                    if (panel) {
                        panel.classList.remove('hidden');
                    }
                });
            });
        });
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Touch/Click Enhancement for Mobile
    // ═══════════════════════════════════════════════════════════════════════

    function enhanceMobileInteractions() {
        // Add touch-action to all interactive elements for better mobile UX
        document.querySelectorAll('button, a, [role="button"]').forEach(function(el) {
            if (!el.style.touchAction) {
                el.style.touchAction = 'manipulation';
            }
        });

        // Prevent double-tap zoom on buttons
        document.querySelectorAll('button').forEach(function(button) {
            button.addEventListener('touchend', function(e) {
                e.preventDefault();
                this.click();
            }, {passive: false});
        });
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Initialize All
    // ═══════════════════════════════════════════════════════════════════════

    function init() {
        initFlashMessages();
        initMobileMenu();
        initTabs();
        enhanceMobileInteractions();
    }

    // Run on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // Re-initialize flash messages after HTMX swaps
    if (typeof htmx !== 'undefined') {
        document.body.addEventListener('htmx:afterSwap', function() {
            initFlashMessages();
        });
    }

})();

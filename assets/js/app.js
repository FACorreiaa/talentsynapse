// TalentSynapse Custom JavaScript
// Handles flash messages, mobile interactions, and UI enhancements

// ═══════════════════════════════════════════════════════════════════════
// Global Mobile Menu Functions (must be global for onclick handlers)
// ═══════════════════════════════════════════════════════════════════════

function toggleMobileMenu() {
    const menu = document.getElementById('mobile-menu');
    const backdrop = document.getElementById('mobile-menu-backdrop');
    const toggle = document.getElementById('mobile-menu-toggle');

    if (menu && backdrop) {
        menu.classList.toggle('hidden');
        backdrop.classList.toggle('hidden');

        if (toggle) {
            const hamburger = toggle.querySelector('.hamburger-icon');
            const closeIcon = toggle.querySelector('.close-icon');
            if (hamburger && closeIcon) {
                hamburger.classList.toggle('hidden');
                closeIcon.classList.toggle('hidden');
            }
        }
    }
}

function closeMobileMenu() {
    const menu = document.getElementById('mobile-menu');
    const backdrop = document.getElementById('mobile-menu-backdrop');
    const toggle = document.getElementById('mobile-menu-toggle');

    if (menu) menu.classList.add('hidden');
    if (backdrop) backdrop.classList.add('hidden');

    if (toggle) {
        const hamburger = toggle.querySelector('.hamburger-icon');
        const closeIcon = toggle.querySelector('.close-icon');
        if (hamburger) hamburger.classList.remove('hidden');
        if (closeIcon) closeIcon.classList.add('hidden');
    }
}

// ═══════════════════════════════════════════════════════════════════════
// Global Tabs Dropdown Functions (for mobile tabs burger menu)
// ═══════════════════════════════════════════════════════════════════════

function toggleTabsDropdown(button) {
    const container = button.parentElement;
    const menu = container.querySelector('.tabs-dropdown-menu');
    const chevron = button.querySelector('.tabs-dropdown-chevron');
    const isExpanded = button.getAttribute('aria-expanded') === 'true';

    if (menu) {
        menu.classList.toggle('hidden');
        button.setAttribute('aria-expanded', !isExpanded);

        if (chevron) {
            chevron.classList.toggle('rotate-180');
        }
    }
}

function closeTabsDropdown(clickedTab) {
    const container = clickedTab.closest('.tabs-container');
    if (container) {
        const button = container.querySelector('[aria-haspopup="true"]');
        const menu = container.querySelector('.tabs-dropdown-menu');
        const chevron = container.querySelector('.tabs-dropdown-chevron');

        if (menu) menu.classList.add('hidden');
        if (button) button.setAttribute('aria-expanded', 'false');
        if (chevron) chevron.classList.remove('rotate-180');
    }
}

// Close tabs dropdown when clicking outside
document.addEventListener('click', function(e) {
    const dropdowns = document.querySelectorAll('.tabs-container');
    dropdowns.forEach(function(container) {
        const button = container.querySelector('[aria-haspopup="true"]');
        const menu = container.querySelector('.tabs-dropdown-menu');

        if (button && menu && !container.contains(e.target)) {
            menu.classList.add('hidden');
            button.setAttribute('aria-expanded', 'false');
            const chevron = container.querySelector('.tabs-dropdown-chevron');
            if (chevron) chevron.classList.remove('rotate-180');
        }
    });
});

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

    // Re-initialize after HTMX swaps
    if (typeof htmx !== 'undefined') {
        document.body.addEventListener('htmx:afterSwap', function() {
            initFlashMessages();
            initMobileMenu();
            initTabs();
        });

        // Close mobile menu before navigation
        document.body.addEventListener('htmx:beforeRequest', function() {
            closeMobileMenu();
        });
    }

})();

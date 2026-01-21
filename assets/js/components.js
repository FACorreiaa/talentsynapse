/**
 * Shared Alpine.js Components for SkillSphere
 * Following Locality of Behavior (LoB) principles
 */

document.addEventListener('alpine:init', () => {
  // Global Theme Store
  Alpine.store('theme', {
    dark: localStorage.getItem('theme') !== 'light',

    init() {
      // Apply initial theme
      this.apply();
    },

    apply() {
      document.documentElement.classList.toggle('dark', this.dark);
    },

    toggle() {
      this.dark = !this.dark;
      localStorage.setItem('theme', this.dark ? 'dark' : 'light');
      this.apply();
    }
  });

  // Initialize theme on load
  Alpine.store('theme').init();
});

/**
 * Dropdown Component
 * Usage: x-data="dropdown()"
 */
window.dropdown = function() {
  return {
    open: false,

    toggle() {
      this.open = !this.open;
    },

    close() {
      this.open = false;
    }
  };
};

/**
 * Password Toggle Component (Alpine.js version - if preferred over Hyperscript)
 * Usage: x-data="passwordToggle()"
 */
window.passwordToggle = function() {
  return {
    visible: false,

    toggle() {
      this.visible = !this.visible;
    },

    get type() {
      return this.visible ? 'text' : 'password';
    }
  };
};

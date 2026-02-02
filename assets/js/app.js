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
    // Push Notification Management
    // ═══════════════════════════════════════════════════════════════════════

    async function initPushNotifications() {
        // Check if service worker and push are supported
        if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
            console.log('[Push] Service Worker or Push not supported');
            return;
        }

        try {
            // Wait for service worker to be ready
            const registration = await navigator.serviceWorker.ready;

            // Check if already subscribed
            const existingSubscription = await registration.pushManager.getSubscription();

            if (existingSubscription) {
                console.log('[Push] Already subscribed');
                return;
            }

            // Check notification permission
            if (Notification.permission === 'denied') {
                console.log('[Push] Notifications are blocked');
                return;
            }

            // If permission hasn't been granted, we'll wait for user interaction
            // This should be triggered by a button click, not automatically
            // You can expose a function globally for this
            window.requestNotificationPermission = async function() {
                try {
                    const permission = await Notification.requestPermission();

                    if (permission !== 'granted') {
                        console.log('[Push] Permission not granted');
                        return false;
                    }

                    await subscribeToPush(registration);
                    return true;
                } catch (error) {
                    console.error('[Push] Error requesting permission:', error);
                    return false;
                }
            };

            // Auto-subscribe if already granted
            if (Notification.permission === 'granted') {
                await subscribeToPush(registration);
            }

        } catch (error) {
            console.error('[Push] Initialization error:', error);
        }
    }

    async function subscribeToPush(registration) {
        try {
            // Get VAPID public key from server
            const response = await fetch('/api/push/vapid-key');
            const { publicKey } = await response.json();

            // Subscribe to push
            const subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: urlBase64ToUint8Array(publicKey)
            });

            // Send subscription to server
            const subResponse = await fetch('/api/push/subscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(subscription)
            });

            if (subResponse.ok) {
                console.log('[Push] Successfully subscribed');
            } else {
                console.error('[Push] Failed to store subscription on server');
            }
        } catch (error) {
            console.error('[Push] Subscription error:', error);
        }
    }

    // Helper function to convert VAPID key
    function urlBase64ToUint8Array(base64String) {
        const padding = '='.repeat((4 - base64String.length % 4) % 4);
        const base64 = (base64String + padding)
            .replace(/\-/g, '+')
            .replace(/_/g, '/');

        const rawData = window.atob(base64);
        const outputArray = new Uint8Array(rawData.length);

        for (let i = 0; i < rawData.length; ++i) {
            outputArray[i] = rawData.charCodeAt(i);
        }
        return outputArray;
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Initialize All
    // ═══════════════════════════════════════════════════════════════════════

    function init() {
        initFlashMessages();
        initMobileMenu();
        initTabs();
        enhanceMobileInteractions();
        initPushNotifications();
    }

    // Run on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // ═══════════════════════════════════════════════════════════════════════
    // HTMX Navigation Handlers
    // ═══════════════════════════════════════════════════════════════════════

    // Re-initialize after HTMX swaps
    if (typeof htmx !== 'undefined') {
        // Mark navbar as loaded after first page load to prevent animation replay
        document.addEventListener('DOMContentLoaded', function() {
            setTimeout(function() {
                const navbar = document.querySelector('.nav-premium');
                if (navbar) {
                    navbar.setAttribute('data-loaded', 'true');
                }
            }, 1000); // Wait for initial animations to complete
        });

        document.body.addEventListener('htmx:afterSwap', function() {
            initFlashMessages();
            initMobileMenu();
            initTabs();
            initChatTypingIndicator();
        });

        // Before HTMX navigation: close mobile menu and mark navbar
        document.body.addEventListener('htmx:beforeRequest', function() {
            closeMobileMenu();
            // Ensure navbar stays stable during navigation
            const navbar = document.querySelector('.nav-premium');
            if (navbar) {
                navbar.setAttribute('data-loaded', 'true');
            }
        });

        // Initialize typing indicator on initial load
        initChatTypingIndicator();
    }

})();

// ═══════════════════════════════════════════════════════════════════════
// Chat Typing Indicator
// ═══════════════════════════════════════════════════════════════════════

function initChatTypingIndicator() {
    const messageInput = document.getElementById('message-input');
    const typingIndicator = document.getElementById('typing-indicator');
    const messagesContainer = document.getElementById('messages-container');

    if (!messageInput || !messagesContainer) {
        return; // Not on chat page
    }

    let typingTimer;
    const typingTimeout = 1000; // 1 second debounce
    let isTyping = false;
    let ws = null;

    // Get WebSocket connection from HTMX ws extension
    function getWebSocket() {
        // HTMX ws extension stores connection on the element
        const wsElement = messagesContainer;
        if (wsElement && wsElement._ws) {
            return wsElement._ws;
        }
        return null;
    }

    function sendTypingIndicator(typing) {
        ws = getWebSocket();
        if (ws && ws.readyState === WebSocket.OPEN) {
            const conversationID = messagesContainer.getAttribute('ws-connect').split('conversation=')[1];
            const message = {
                type: 'typing',
                conversation_id: conversationID,
                is_typing: typing
            };
            ws.send(JSON.stringify(message));
            isTyping = typing;
        }
    }

    // Handle input events
    messageInput.addEventListener('input', function() {
        clearTimeout(typingTimer);

        if (!isTyping) {
            sendTypingIndicator(true);
        }

        typingTimer = setTimeout(function() {
            sendTypingIndicator(false);
        }, typingTimeout);
    });

    // Stop typing indicator on blur or form submit
    messageInput.addEventListener('blur', function() {
        clearTimeout(typingTimer);
        if (isTyping) {
            sendTypingIndicator(false);
        }
    });

    // Listen for typing events from WebSocket
    if (messagesContainer) {
        // Use HTMX ws extension's message handling
        messagesContainer.addEventListener('htmx:wsAfterMessage', function(event) {
            try {
                const data = JSON.parse(event.detail.message);
                if (data.type === 'typing' && typingIndicator) {
                    if (data.is_typing) {
                        typingIndicator.style.display = 'flex';
                        // Auto-scroll to show typing indicator
                        messagesContainer.scrollTop = messagesContainer.scrollHeight;
                    } else {
                        typingIndicator.style.display = 'none';
                    }
                }
            } catch (e) {
                // Not JSON or not a typing message, ignore
            }
        });
    }
}

// ═══════════════════════════════════════════════════════════════════════
// Voice Recording Functions
// ═══════════════════════════════════════════════════════════════════════

let voiceRecorder = null;
let recordingTimer = null;
let recordingStartTime = null;
let recordingAnalyser = null;
let animationFrameId = null;

class VoiceRecorder {
    constructor() {
        this.mediaRecorder = null;
        this.audioChunks = [];
        this.stream = null;
        this.audioContext = null;
    }

    async startRecording() {
        try {
            this.stream = await navigator.mediaDevices.getUserMedia({ audio: true });

            // Create audio context for visualization
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
            const source = this.audioContext.createMediaStreamSource(this.stream);
            recordingAnalyser = this.audioContext.createAnalyser();
            recordingAnalyser.fftSize = 64;
            source.connect(recordingAnalyser);

            this.mediaRecorder = new MediaRecorder(this.stream, {
                mimeType: this.getSupportedMimeType()
            });

            this.audioChunks = [];

            this.mediaRecorder.ondataavailable = (event) => {
                if (event.data.size > 0) {
                    this.audioChunks.push(event.data);
                }
            };

            this.mediaRecorder.onstop = () => {
                const mimeType = this.mediaRecorder.mimeType || 'audio/webm';
                const audioBlob = new Blob(this.audioChunks, { type: mimeType });
                this.uploadVoiceMessage(audioBlob, mimeType);
                this.cleanup();
            };

            this.mediaRecorder.start(100); // Collect data every 100ms
            return true;
        } catch (err) {
            console.error('Failed to start recording:', err);
            alert('Could not access microphone. Please check permissions.');
            return false;
        }
    }

    getSupportedMimeType() {
        const types = [
            'audio/webm;codecs=opus',
            'audio/webm',
            'audio/ogg;codecs=opus',
            'audio/mp4',
        ];
        for (const type of types) {
            if (MediaRecorder.isTypeSupported(type)) {
                return type;
            }
        }
        return 'audio/webm';
    }

    stopRecording() {
        if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
            this.mediaRecorder.stop();
        }
    }

    cancelRecording() {
        if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
            this.mediaRecorder.stop();
        }
        this.audioChunks = [];
        this.cleanup();
    }

    cleanup() {
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
            this.stream = null;
        }
        if (this.audioContext) {
            this.audioContext.close();
            this.audioContext = null;
        }
        recordingAnalyser = null;
    }

    async uploadVoiceMessage(audioBlob, mimeType) {
        // Get conversation ID from URL
        const pathParts = window.location.pathname.split('/');
        const chatIndex = pathParts.indexOf('chat');
        if (chatIndex === -1 || !pathParts[chatIndex + 1]) {
            console.error('Could not determine conversation ID');
            return;
        }
        const conversationID = pathParts[chatIndex + 1];

        // Calculate duration
        const duration = Math.floor((Date.now() - recordingStartTime) / 1000);

        const formData = new FormData();
        const extension = mimeType.includes('webm') ? 'webm' :
                         mimeType.includes('ogg') ? 'ogg' :
                         mimeType.includes('mp4') ? 'm4a' : 'webm';
        formData.append('file', audioBlob, `voice-message.${extension}`);
        formData.append('type', 'voice');
        formData.append('duration', duration.toString());

        try {
            const response = await fetch(`/chat/${conversationID}/voice`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error('Failed to upload voice message');
            }

            // Get the HTML response and append to messages
            const html = await response.text();
            const messagesList = document.getElementById('messages-list');
            if (messagesList) {
                messagesList.insertAdjacentHTML('beforeend', html);
                const container = document.getElementById('messages-container');
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            }
        } catch (err) {
            console.error('Failed to upload voice message:', err);
            alert('Failed to send voice message. Please try again.');
        }
    }
}

function startVoiceRecording() {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        alert('Voice recording is not supported in this browser.');
        return;
    }

    voiceRecorder = new VoiceRecorder();
    voiceRecorder.startRecording().then(started => {
        if (started) {
            showRecordingUI();
            startRecordingTimer();
            animateRecordingWaveform();
        }
    });
}

function stopVoiceRecording() {
    if (voiceRecorder) {
        voiceRecorder.stopRecording();
    }
    hideRecordingUI();
    stopRecordingTimer();
    stopWaveformAnimation();
}

function cancelVoiceRecording() {
    if (voiceRecorder) {
        voiceRecorder.cancelRecording();
    }
    hideRecordingUI();
    stopRecordingTimer();
    stopWaveformAnimation();
}

function showRecordingUI() {
    const form = document.getElementById('message-form');
    const recordingUI = document.getElementById('voice-recording-ui');
    if (form) form.classList.add('hidden');
    if (recordingUI) recordingUI.classList.remove('hidden');
}

function hideRecordingUI() {
    const form = document.getElementById('message-form');
    const recordingUI = document.getElementById('voice-recording-ui');
    if (form) form.classList.remove('hidden');
    if (recordingUI) recordingUI.classList.add('hidden');
}

function startRecordingTimer() {
    recordingStartTime = Date.now();
    const timeDisplay = document.getElementById('recording-time');

    recordingTimer = setInterval(() => {
        const elapsed = Math.floor((Date.now() - recordingStartTime) / 1000);
        const mins = Math.floor(elapsed / 60);
        const secs = elapsed % 60;
        if (timeDisplay) {
            timeDisplay.textContent = `${mins}:${secs.toString().padStart(2, '0')}`;
        }
    }, 1000);
}

function stopRecordingTimer() {
    if (recordingTimer) {
        clearInterval(recordingTimer);
        recordingTimer = null;
    }
    recordingStartTime = null;
}

function animateRecordingWaveform() {
    const bars = document.querySelectorAll('#recording-waveform .recording-bar');
    if (!bars.length) return;

    function animate() {
        if (recordingAnalyser) {
            const dataArray = new Uint8Array(recordingAnalyser.frequencyBinCount);
            recordingAnalyser.getByteFrequencyData(dataArray);

            bars.forEach((bar, i) => {
                const value = dataArray[i % dataArray.length] || 0;
                const height = Math.max(4, (value / 255) * 28);
                bar.style.height = `${height}px`;
            });
        } else {
            // Fallback animation when no analyser
            bars.forEach((bar, i) => {
                const height = Math.random() * 24 + 4;
                bar.style.height = `${height}px`;
            });
        }
        animationFrameId = requestAnimationFrame(animate);
    }
    animate();
}

function stopWaveformAnimation() {
    if (animationFrameId) {
        cancelAnimationFrame(animationFrameId);
        animationFrameId = null;
    }
}

// Voice message playback
let currentAudio = null;
let currentPlayButton = null;

function toggleVoicePlayback(button) {
    const audioUrl = button.dataset.audioUrl;
    if (!audioUrl) return;

    const playIcon = button.querySelector('.play-icon');
    const pauseIcon = button.querySelector('.pause-icon');
    const container = button.closest('.voice-message-player');
    const currentTimeEl = container?.querySelector('.voice-current-time');

    // If clicking the same button that's currently playing
    if (currentAudio && currentPlayButton === button) {
        if (currentAudio.paused) {
            currentAudio.play();
            playIcon?.classList.add('hidden');
            pauseIcon?.classList.remove('hidden');
        } else {
            currentAudio.pause();
            playIcon?.classList.remove('hidden');
            pauseIcon?.classList.add('hidden');
        }
        return;
    }

    // Stop any currently playing audio
    if (currentAudio) {
        currentAudio.pause();
        currentAudio.currentTime = 0;
        if (currentPlayButton) {
            const oldPlayIcon = currentPlayButton.querySelector('.play-icon');
            const oldPauseIcon = currentPlayButton.querySelector('.pause-icon');
            oldPlayIcon?.classList.remove('hidden');
            oldPauseIcon?.classList.add('hidden');
        }
    }

    // Create and play new audio
    currentAudio = new Audio(audioUrl);
    currentPlayButton = button;

    currentAudio.ontimeupdate = () => {
        if (currentTimeEl) {
            const mins = Math.floor(currentAudio.currentTime / 60);
            const secs = Math.floor(currentAudio.currentTime % 60);
            currentTimeEl.textContent = `${mins}:${secs.toString().padStart(2, '0')}`;
        }
    };

    currentAudio.onended = () => {
        playIcon?.classList.remove('hidden');
        pauseIcon?.classList.add('hidden');
        if (currentTimeEl) currentTimeEl.textContent = '0:00';
        currentAudio = null;
        currentPlayButton = null;
    };

    currentAudio.onerror = () => {
        console.error('Failed to play audio');
        playIcon?.classList.remove('hidden');
        pauseIcon?.classList.add('hidden');
        currentAudio = null;
        currentPlayButton = null;
    };

    currentAudio.play().then(() => {
        playIcon?.classList.add('hidden');
        pauseIcon?.classList.remove('hidden');
    }).catch(err => {
        console.error('Failed to play audio:', err);
    });
}


// ═══════════════════════════════════════════════════════════════════════
// File Upload Functions
// ═══════════════════════════════════════════════════════════════════════

async function handleFileUpload(event) {
    const file = event.target.files[0];
    if (!file) return;

    // Check file size (max 10MB)
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
        alert('File is too large. Maximum size is 10MB.');
        event.target.value = '';
        return;
    }

    // Get conversation ID from URL
    const pathParts = window.location.pathname.split('/');
    const chatIndex = pathParts.indexOf('chat');
    if (chatIndex === -1 || !pathParts[chatIndex + 1]) {
        console.error('Could not determine conversation ID');
        return;
    }
    const conversationID = pathParts[chatIndex + 1];

    // Show upload progress
    showUploadProgress(file.name);

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`/chat/${conversationID}/upload`, {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            throw new Error('Upload failed');
        }

        // Get the HTML response and append to messages
        const html = await response.text();
        const messagesList = document.getElementById('messages-list');
        if (messagesList) {
            messagesList.insertAdjacentHTML('beforeend', html);
            const container = document.getElementById('messages-container');
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        }

        hideUploadProgress();
    } catch (err) {
        console.error('File upload failed:', err);
        alert('Failed to upload file. Please try again.');
        hideUploadProgress();
    } finally {
        // Reset file input
        event.target.value = '';
    }
}

function showUploadProgress(fileName) {
    const container = document.getElementById('messages-container');
    if (!container) return;

    const progressHTML = `
        <div id="upload-progress" class="flex items-center gap-3 p-4 bg-muted/50 rounded-lg animate-pulse">
            <div class="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center">
                <svg class="w-5 h-5 text-primary animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            </div>
            <div class="flex-1">
                <p class="text-sm font-medium">Uploading...</p>
                <p class="text-xs text-muted-foreground">${fileName}</p>
            </div>
        </div>
    `;

    const messagesList = document.getElementById('messages-list');
    if (messagesList) {
        messagesList.insertAdjacentHTML('beforeend', progressHTML);
        container.scrollTop = container.scrollHeight;
    }
}

function hideUploadProgress() {
    const progress = document.getElementById('upload-progress');
    if (progress) {
        progress.remove();
    }
}

// Image modal for viewing full-size images
function openImageModal(imageSrc) {
    const modal = document.createElement('div');
    modal.id = 'image-modal';
    modal.className = 'fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4';
    modal.onclick = () => modal.remove();

    modal.innerHTML = `
        <div class="relative max-w-4xl max-h-full">
            <img src="${imageSrc}" class="max-w-full max-h-[90vh] object-contain rounded-lg" alt="Full size image"/>
            <button onclick="event.stopPropagation(); this.closest('#image-modal').remove()" 
                    class="absolute top-4 right-4 w-10 h-10 rounded-full bg-white/20 hover:bg-white/30 flex items-center justify-center transition-colors">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
            </button>
        </div>
    `;

    document.body.appendChild(modal);
}

// Close image modal on Escape key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        const modal = document.getElementById('image-modal');
        if (modal) modal.remove();
    }
});

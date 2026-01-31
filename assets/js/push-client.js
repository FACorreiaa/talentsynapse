// Push Notification Client Logic

// Helper to convert VAPID key
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

// Check if Push is supported
function isPushSupported() {
    return 'serviceWorker' in navigator && 'PushManager' in window;
}

// Request permission and subscribe
async function subscribeToPush() {
    if (!isPushSupported()) {
        console.log('Push messaging is not supported');
        return;
    }

    try {
        const permission = await Notification.requestPermission();
        if (permission !== 'granted') {
            console.log('Notification permission denied');
            return;
        }

        const registration = await navigator.serviceWorker.ready;

        // Get VAPID key from server
        const response = await fetch('/api/push/vapid-key');
        const { publicKey } = await response.json();
        const convertedVapidKey = urlBase64ToUint8Array(publicKey);

        // Subscribe
        const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: convertedVapidKey
        });

        console.log('User is subscribed:', subscription);

        // Send subscription to server
        await fetch('/api/push/subscribe', {
            method: 'POST',
            body: JSON.stringify(subscription),
            headers: {
                'Content-Type': 'application/json'
            }
        });

        console.log('Subscription sent to server');
        alert('Notifications enabled!');

    } catch (err) {
        console.error('Failed to subscribe the user: ', err);
    }
}

// Expose globally
window.subscribeToPush = subscribeToPush;

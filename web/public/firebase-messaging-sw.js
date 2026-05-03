importScripts('https://www.gstatic.com/firebasejs/10.8.0/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/10.8.0/firebase-messaging-compat.js');

 firebase.initializeApp({
  apiKey: "AIzaSyAo1Wg3SyIwrM1WUkb1I28pqSbyuXmkC_s",
  authDomain: "test-mazadpay.firebaseapp.com",
  projectId: "test-mazadpay",
  storageBucket: "test-mazadpay.firebasestorage.app",
  messagingSenderId: "389117490995",
  appId: "1:389117490995:web:9abc3ecb0990fb1f23f046"
});

const messaging = firebase.messaging();

// Gestion des messages en arrière-plan
messaging.onBackgroundMessage((payload) => {
  console.log('[firebase-messaging-sw.js] Message reçu en arrière-plan:', payload);

  const notificationTitle = payload.notification?.title || 'MazadPay';
  const notificationOptions = {
    body: payload.notification?.body || '',
    icon: '/logo192.png',
    badge: '/logo192.png',
    tag: payload.data?.notification_id || 'default',
    requireInteraction: false,
    data: payload.data || {}
  };

  return self.registration.showNotification(notificationTitle, notificationOptions);
});

// Gestion du clic sur la notification
self.addEventListener('notificationclick', (event) => {
  console.log('[firebase-messaging-sw.js] Notification cliquée:', event);

  event.notification.close();

  const notificationData = event.notification.data;
  let url = '/';

   if (notificationData?.type === 'auction_approved') {
    url = `/auctions/${notificationData.auction_id}`;
  } else if (notificationData?.type === 'auction_ending_soon') {
    url = `/auctions/${notificationData.auction_id}`;
  } else if (notificationData?.type === 'bid_placed') {
    url = `/auctions/${notificationData.auction_id}`;
  } else if (notificationData?.type === 'auction_won') {
    url = `/auctions/${notificationData.auction_id}`;
  }

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      // Si un onglet est déjà ouvert, le focaliser
      for (const client of clientList) {
        if (client.url.includes(self.location.origin) && 'focus' in client) {
          return client.focus().then(() => client.navigate(url));
        }
      }
       if (clients.openWindow) {
        return clients.openWindow(url);
      }
    })
  );
});

 self.addEventListener('install', (event) => {
  console.log('[firebase-messaging-sw.js] Service Worker installé');
  self.skipWaiting();
});

// Activation du Service Worker
self.addEventListener('activate', (event) => {
  console.log('[firebase-messaging-sw.js] Service Worker activé');
  event.waitUntil(clients.claim());
});

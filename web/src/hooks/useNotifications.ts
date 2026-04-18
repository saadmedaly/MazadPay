import { useEffect, useState } from 'react';
import { messaging, requestNotificationPermission } from '../lib/firebase';
import { onMessage } from 'firebase/messaging';
import { useAuthStore } from '../stores/authStore';
import api from '../api/client';

export const useNotifications = () => {
  const { user, isAuthenticated } = useAuthStore();
  const [permission, setPermission] = useState<NotificationPermission>(Notification.permission);
  const [fcmToken, setFcmToken] = useState<string | null>(null);

  useEffect(() => {
    if (isAuthenticated && user) {
      handlePermission();
    }
  }, [isAuthenticated, user]);

  const handlePermission = async () => {
    const token = await requestNotificationPermission();
    if (token) {
      setFcmToken(token);
      setPermission('granted');
      
      // Sauvegarder le token sur le backend
      try {
        await api.post('/v1/api/notifications/push-tokens', {
          fcm_token: token,
          device_id: navigator.userAgent, // Simple device ID
          platform: 'web'
        });
      } catch (err) {
        console.error('Failed to sync FCM token:', err);
      }
    }
  };

  useEffect(() => {
    if (!messaging) return;

    const unsubscribe = onMessage(messaging, (payload) => {
      console.log('Message received in foreground: ', payload);
      // On peut ajouter une notification locale ou mettre à jour un store ici
      if (Notification.permission === 'granted') {
        new Notification(payload.notification?.title || 'Notification', {
          body: payload.notification?.body,
          icon: '/logo192.png'
        });
      }
    });

    return () => unsubscribe();
  }, []);

  return { permission, fcmToken, requestPermission: handlePermission };
};
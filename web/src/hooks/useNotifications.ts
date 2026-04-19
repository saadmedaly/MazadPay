import { useEffect, useState, useRef } from 'react';
import { messaging, requestNotificationPermission } from '../lib/firebase';
import { onMessage } from 'firebase/messaging';
import { useAuthStore } from '../stores/authStore';
import api from '../api/client';

export const useNotifications = () => {
  const { user, isAuthenticated } = useAuthStore();
  const [permission, setPermission] = useState<NotificationPermission | null>(null);
  const [fcmToken, setFcmToken] = useState<string | null>(null);
  const initialized = useRef(false);

  useEffect(() => {
    const initPermissions = async () => {
      if (initialized.current === true) return;
      if (!isAuthenticated || !user) return;
      initialized.current = true;

      const token = await requestNotificationPermission();
      if (token) {
        setFcmToken(token);
        setPermission('granted');

        try {
          await api.post('/v1/api/notifications/push-tokens', {
            fcm_token: token,
            device_id: navigator.userAgent,
            platform: 'web'
          });
        } catch (err) {
          console.error('Failed to sync FCM token:', err);
        }
      }
    };

    initPermissions();
  }, [isAuthenticated, user]);

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

  const requestPermission = async () => {
    if (initialized.current === true) return;
    if (!isAuthenticated || !user) return;
    initialized.current = true;

    const token = await requestNotificationPermission();
    if (token) {
      setFcmToken(token);
      setPermission('granted');

      try {
        await api.post('/v1/api/notifications/push-tokens', {
          fcm_token: token,
          device_id: navigator.userAgent,
          platform: 'web'
        });
      } catch (err) {
        console.error('Failed to sync FCM token:', err);
      }
    }
  };

  return { permission, fcmToken, requestPermission };
};
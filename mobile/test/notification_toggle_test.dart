import 'package:flutter_test/flutter_test.dart';

// Customer request #6B: Notifications toggle in the Account page. Audit
// found this already fully implemented and wired (mobile/lib/pages/
// account_profile_page.dart's _toggleNotifications -> UserApi.
// updateNotificationPrefs -> PUT /users/me/notification-prefs, protected by
// jwtMiddleware, persisted server-side via
// UserService.UpdateNotificationSettings(userID, enabled) -- scoped to the
// authenticated user's own row). The initial value is loaded from
// _userData['notifications_enabled'] on page load (_loadUserProfile), so it
// survives app restarts since it's a server-side preference, not a local-
// only flag. No code was changed for #6B -- this test covers the
// optimistic-update-with-revert-on-failure state machine
// (_toggleNotifications's core logic), reproduced here rather than pumping
// the real widget, since _toggleNotifications lives on a State that
// constructs UserApi()/Dio/dotenv directly in initState.
class _ToggleState {
  bool enabled;
  bool isToggling;
  _ToggleState({required this.enabled, this.isToggling = false});
}

Future<void> toggleNotifications(
  _ToggleState state,
  bool newValue,
  Future<bool> Function(bool) persist,
) async {
  if (state.isToggling) return;
  final previous = state.enabled;
  state.enabled = newValue;
  state.isToggling = true;

  final success = await persist(newValue);

  state.isToggling = false;
  if (!success) {
    state.enabled = previous;
  }
}

void main() {
  group('notification toggle optimistic update with revert-on-failure', () {
    test('initial state reflects the loaded value', () {
      final state = _ToggleState(enabled: true);
      expect(state.enabled, true);
    });

    test('toggling ON updates immediately (optimistic) before the network call resolves', () async {
      final state = _ToggleState(enabled: false);
      final future = toggleNotifications(state, true, (_) async {
        // Mid-flight: optimistic update should already be visible.
        expect(state.enabled, true);
        expect(state.isToggling, true);
        return true;
      });
      await future;
      expect(state.enabled, true);
      expect(state.isToggling, false);
    });

    test('toggling OFF and a successful persist keeps the new value', () async {
      final state = _ToggleState(enabled: true);
      await toggleNotifications(state, false, (_) async => true);
      expect(state.enabled, false);
      expect(state.isToggling, false);
    });

    test('a failed persist reverts to the previous value, never leaving a state the backend did not save', () async {
      final state = _ToggleState(enabled: false);
      await toggleNotifications(state, true, (_) async => false);
      expect(state.enabled, false); // reverted
      expect(state.isToggling, false);
    });

    test('a second toggle call while one is already in flight is a no-op (double-toggle safety)', () async {
      final state = _ToggleState(enabled: false, isToggling: true);
      var persistCalled = false;
      await toggleNotifications(state, true, (_) async {
        persistCalled = true;
        return true;
      });
      expect(persistCalled, false);
      expect(state.enabled, false); // untouched
    });
  });
}

import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/settings_page.dart';

// Client feedback: Bug D contract hardening. SettingsPage's GET response
// parsing previously read response.data!['settings'] -- a nested key that
// never existed in the backend's actual response shape
// ({"success":true,"data":{...settings fields...}}) -- so it silently
// always fell back to hardcoded local defaults, regardless of what the
// server had stored. mergeSettingsResponse is the extracted, pure merge
// logic now used by _loadSettings; these tests prove it directly without
// pumping a widget tree or mocking the network layer.
void main() {
  group('mergeSettingsResponse (Bug D contract hardening)', () {
    test('server values populate the page for the three real backend-wired settings', () {
      const localDefaults = {
        'notifications_push': true,
        'notifications_email': true,
        'notifications_sms': false,
        'public_profile': true,
        'show_bid_history': true,
      };
      const serverData = {
        'user_id': 'abc',
        'currency': 'MRU',
        'theme': 'dark',
        'language': 'fr',
        'notifications_email': false,
        'notifications_push': false,
        'notifications_sms': true,
        'two_factor_enabled': true,
      };

      final merged = mergeSettingsResponse(localDefaults, serverData);

      expect(merged['notifications_push'], false);
      expect(merged['notifications_email'], false);
      expect(merged['notifications_sms'], true);
    });

    test('theme and language from the server are present in the merged map (even though no UI reads them yet)', () {
      const localDefaults = {'notifications_push': true};
      const serverData = {'theme': 'dark', 'language': 'fr'};

      final merged = mergeSettingsResponse(localDefaults, serverData);

      expect(merged['theme'], 'dark');
      expect(merged['language'], 'fr');
    });

    test('notifications_push false from the server is not coerced to true or omitted', () {
      const localDefaults = {'notifications_push': true};
      const serverData = {'notifications_push': false};

      final merged = mergeSettingsResponse(localDefaults, serverData);

      expect(merged['notifications_push'], isFalse);
      expect(merged.containsKey('notifications_push'), isTrue);
    });

    test('local-only keys with no backend column survive when the server omits them', () {
      const localDefaults = {
        'notifications_push': true,
        'public_profile': true,
        'show_bid_history': true,
      };
      const serverData = {'notifications_push': false};

      final merged = mergeSettingsResponse(localDefaults, serverData);

      expect(merged['public_profile'], true);
      expect(merged['show_bid_history'], true);
      expect(merged['notifications_push'], false);
    });

    test('an empty server response leaves local defaults entirely intact (safe fallback)', () {
      const localDefaults = {
        'notifications_push': true,
        'notifications_email': true,
        'notifications_sms': false,
      };

      final merged = mergeSettingsResponse(localDefaults, const {});

      expect(merged, localDefaults);
    });

    test('the old contract (a nested "settings" key) is no longer required -- server fields are read directly', () {
      const localDefaults = {'notifications_push': true};
      // This is the ACTUAL shape response.data has (no nested "settings"
      // wrapper) -- the bug was reading response.data!['settings'], which
      // always evaluated to null against a payload shaped like this.
      const serverData = {'notifications_push': false, 'theme': 'auto'};

      final merged = mergeSettingsResponse(localDefaults, serverData);

      expect(merged['notifications_push'], false);
      expect(merged.containsKey('settings'), isFalse);
    });
  });
}

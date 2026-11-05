import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/login_page.dart';
import 'package:mezadpay/providers/auth_provider.dart';
import 'package:mezadpay/services/auth_service.dart';
import 'package:mezadpay/services/fcm_service.dart';
import 'package:mezadpay/services/global_websocket_service.dart';
import 'package:mezadpay/services/user_api.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Self-service account deletion (App Store guideline 5.1.1(v)), shared by
/// AccountProfilePage and SettingsPage: confirm → DELETE /users/me → wipe the
/// local session → Login with the whole stack removed. Nothing is sent before
/// the user explicitly confirms.
class AccountDeletionFlow {
  AccountDeletionFlow({UserApi? api, Future<void> Function()? clearLocalSession, WidgetBuilder? loginBuilder})
      : _api = api ?? UserApi(),
        _clearLocalSession = clearLocalSession ?? clearLocalAccountData,
        _loginBuilder = loginBuilder ?? ((_) => const LoginPage());

  final UserApi _api;
  final Future<void> Function() _clearLocalSession;
  final WidgetBuilder _loginBuilder;

  /// Returns true when the account was deleted.
  Future<bool> start(BuildContext context) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(l10n.delete_account_confirm_title),
        content: Text(l10n.delete_account_confirm_body),
        actions: [
          TextButton(
            key: const Key('delete-account-cancel'),
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: Text(l10n.delete_account_cancel),
          ),
          TextButton(
            key: const Key('delete-account-confirm'),
            onPressed: () => Navigator.of(dialogContext).pop(true),
            child: Text(l10n.delete_account_confirm_button, style: const TextStyle(color: Colors.red, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
    if (confirmed != true || !context.mounted) return false;

    final response = await _api.deleteAccount();
    if (!context.mounted) return false;

    if (!response.success) {
      final reason = blockedReason(l10n, response.error?.code);
      if (reason != null) {
        await showDialog<void>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            title: Text(l10n.delete_account_blocked_title),
            content: Text(reason),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(dialogContext).pop(),
                child: Text(l10n.delete_account_blocked_ok),
              ),
            ],
          ),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.delete_account_failed), backgroundColor: Colors.red),
        );
      }
      return false;
    }

    try {
      ProviderScope.containerOf(context, listen: false).invalidate(authNotifierProvider);
    } catch (_) {
      // No ProviderScope above (e.g. isolated widget tests): nothing to reset.
    }
    await _clearLocalSession();
    if (!context.mounted) return true;

    final messenger = ScaffoldMessenger.of(context);
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(builder: _loginBuilder),
      (route) => false,
    );
    messenger.showSnackBar(
      SnackBar(content: Text(l10n.delete_account_success), backgroundColor: Colors.green),
    );
    return true;
  }

  /// Localized explanation for a 409 account_deletion_blocked_* code, or null.
  static String? blockedReason(AppLocalizations l10n, String? code) {
    switch (code) {
      case 'account_deletion_blocked_balance':
        return l10n.delete_account_blocked_balance;
      case 'account_deletion_blocked_hold':
        return l10n.delete_account_blocked_hold;
      case 'account_deletion_blocked_pending_deposit':
        return l10n.delete_account_blocked_pending_deposit;
      case 'account_deletion_blocked_pending_withdrawal':
        return l10n.delete_account_blocked_pending_withdrawal;
      case 'account_deletion_blocked_leading_bid':
        return l10n.delete_account_blocked_leading_bid;
      case 'account_deletion_blocked_unpaid_win':
        return l10n.delete_account_blocked_unpaid_win;
      case 'account_deletion_blocked_active_listing':
        return l10n.delete_account_blocked_active_listing;
    }
    return null;
  }
}

/// Local data tied to the deleted account. Onboarding and language choices
/// belong to the device, not the account, and are kept.
const accountScopedPrefKeys = [
  'fcm_token',
  'device_id',
  'cached_location',
  'has_shown_location_modal',
  'local_favorites',
  'local_favorite_auctions',
  'favorites_pending_sync',
];

/// Wipes everything the deleted account left on this device: every secure
/// storage entry (JWT, refresh token, user id), the realtime/WebSocket
/// connections, the FCM token and account-scoped preferences.
Future<void> clearLocalAccountData() async {
  await AuthService().clearAll(); // also stops RealtimeSyncService
  GlobalWebsocketService().disconnect();
  await FCMService().clearLocalToken();
  final prefs = await SharedPreferences.getInstance();
  for (final key in accountScopedPrefKeys) {
    await prefs.remove(key);
  }
}

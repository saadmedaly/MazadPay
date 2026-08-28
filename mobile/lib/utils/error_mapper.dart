import 'package:flutter/widgets.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

String mapError(BuildContext context, String? code, String defaultMessage) {
  final l10n = AppLocalizations.of(context)!;
  switch (code) {
    case 'invalid_pin':
      return l10n.error_invalid_pin;
    case 'unauthorized':
    case 'user_not_found':
      return l10n.error_user_not_registered;
    case 'not_found':
    case 'invalid_credentials':
      return l10n.error_invalid_credentials;
    case 'connection_error':
    case 'connection_timeout':
    case 'server_unreachable':
      return l10n.error_connection;
    case 'too_many_requests':
    case 'otp_rate_limited':
    case 'reset_password_rate_limited':
      return l10n.error_too_many_requests;
    case 'account_blocked':
    case 'account_disabled':
      return l10n.error_account_blocked;
    case 'phone_already_registered':
    case 'duplicate_phone':
      return l10n.error_phone_already_registered;
    case 'otp_expired':
      return l10n.error_otp_expired;
    case 'otp_invalid':
    case 'invalid_otp':
      return l10n.error_otp_invalid;
    case 'otp_max_attempts':
      return l10n.error_otp_max_attempts;
    case 'conflict':
    case 'forbidden':
      return l10n.error_generic(defaultMessage);
    case 'server_error':
      return l10n.error_server;
    case 'weak_pin':
      return l10n.error_password_too_short;
    case 'invalid_phone':
      return l10n.error_invalid_phone;
    case 'not_request_owner':
      return l10n.error_generic(defaultMessage);
    case 'invalid_status':
      return l10n.error_generic(defaultMessage);
    case 'bad_request':
    case 'validation_error':
      return l10n.error_generic(defaultMessage);
    default:
      final lowered = defaultMessage.toLowerCase();
      if (lowered.contains('connexion') ||
          lowered.contains('connection') ||
          lowered.contains('timeout')) {
        return l10n.error_connection;
      }
      return l10n.error_generic(defaultMessage);
  }
}

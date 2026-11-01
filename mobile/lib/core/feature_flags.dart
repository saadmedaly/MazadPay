// Build-time feature switches (set with --dart-define, compiled into the binary).

/// Bankily B-PAY automatic top-up. OFF unless a build passes
/// `--dart-define=BPAY_ENABLED=true`, so store builds cannot show B-PAY even
/// if the backend later reports it enabled. The B-PAY code stays in the
/// project for a later release.
const bool kBpayEnabled = bool.fromEnvironment('BPAY_ENABLED', defaultValue: false);

import 'package:dio/dio.dart';
import 'package:mezadpay/models/bankily_payment.dart';
import 'package:mezadpay/services/api_service.dart';

/// Bankily B-PAY automatic top-up. Every call goes to the MazadPay backend --
/// the app never talks to Bankily and never holds a Bankily credential. The
/// B-PAY passcode is sent once in the pay request body and never stored.
class BankilyApi {
  ApiService? _apiOverride;
  ApiService get _api => _apiOverride ??= ApiService();

  BankilyApi({ApiService? api}) : _apiOverride = api;

  /// The backend waits up to its own Bankily timeout and then runs a
  /// checkTransaction, so /pay can legitimately take well over the app's
  /// default 30 s receive timeout.
  static const Duration payReceiveTimeout = Duration(seconds: 90);

  Future<BankilyConfig> getConfig() async {
    final resp = await _api.get<Map<String, dynamic>>('/users/wallet/bankily/config');
    final data = resp?['data'];
    if (data is Map<String, dynamic>) return BankilyConfig.fromJson(data);
    return BankilyConfig.disabled;
  }

  /// [idempotencyKey] must stay the same for every retry of one submission:
  /// the backend then returns the original attempt instead of charging again.
  Future<BankilyPayment> pay({
    required String amount,
    required String clientPhone,
    required String passcode,
    required String idempotencyKey,
  }) async {
    final resp = await _api.post<Map<String, dynamic>>(
      '/users/wallet/bankily/pay',
      data: {
        'amount': amount,
        'client_phone': clientPhone,
        'passcode': passcode,
        'idempotency_key': idempotencyKey,
      },
      options: Options(receiveTimeout: payReceiveTimeout),
    );
    return _payment(resp);
  }

  Future<BankilyPayment> getPayment(String id) async {
    return _payment(await _api.get<Map<String, dynamic>>('/users/wallet/bankily/payments/$id'));
  }

  /// Asks the backend to run Bankily checkTransaction now.
  Future<BankilyPayment> check(String id) async {
    return _payment(await _api.post<Map<String, dynamic>>(
      '/users/wallet/bankily/payments/$id/check',
      options: Options(receiveTimeout: payReceiveTimeout),
    ));
  }

  Future<List<BankilyPayment>> list() async {
    final resp = await _api.get<Map<String, dynamic>>('/users/wallet/bankily/payments');
    final data = resp?['data'];
    if (data is! List) return const [];
    return data.whereType<Map<String, dynamic>>().map(BankilyPayment.fromJson).toList();
  }

  BankilyPayment _payment(Map<String, dynamic>? resp) {
    final data = resp?['data'];
    if (resp?['success'] == true && data is Map<String, dynamic>) {
      return BankilyPayment.fromJson(data);
    }
    throw const ApiException('Invalid response', code: 'invalid_response');
  }

  /// Network-level failures where the backend may still have processed the
  /// request: the app must retry with the same idempotency key, never start
  /// a fresh payment.
  static bool isUncertain(Object error) {
    if (error is ApiException) {
      return const {'connectionTimeout', 'sendTimeout', 'receiveTimeout', 'connectionError', 'unknown'}.contains(error.code) ||
          (error.httpStatus != null && error.httpStatus! >= 500);
    }
    return false;
  }
}

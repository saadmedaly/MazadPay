import 'package:mezadpay/models/payment_method.dart';
import 'package:mezadpay/services/api_service.dart';

class PaymentMethodsService {
  final ApiService _api = ApiService();

  Future<List<PaymentMethod>> getPaymentMethods() async {
    // The backend returns a JSON object, typically with a `data` field containing the list.
    final response = await _api.get<Map<String, dynamic>>('/admin/payment-methods');
    if (response == null) return [];
    // Support both direct list response and wrapped `{ data: [...] }` structure.
    final List<dynamic> list =
        response['data'] is List ? response['data'] as List<dynamic> : [];
    return list.map((e) => PaymentMethod.fromJson(e as Map<String, dynamic>)).toList();
  }
}

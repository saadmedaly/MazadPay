import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

class FaqApi {
  final ApiService _apiService = ApiService();

  Future<ApiResponse<List<dynamic>>> getFaqs() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/faq',
      );

      final List<dynamic> faqList = response?['data'] ?? response?['faqs'] ?? response?['items'] ?? response ?? [];
      return ApiResponse.success(faqList);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}

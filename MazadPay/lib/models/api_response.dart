/// Modèle de réponse API standardisé
/// Toutes les réponses de l'API backend suivent ce format
class ApiResponse<T> {
  final bool success;
  final T? data;
  final ApiError? error;
  final String? message;
  
  ApiResponse({
    required this.success,
    this.data,
    this.error,
    this.message,
  });
  
  factory ApiResponse.fromJson(Map<String, dynamic>? json) {
    if (json == null) {
      return ApiResponse<T>(
        success: false,
        error: ApiError(code: 'null_response', message: 'Réponse vide du serveur'),
        message: 'Réponse vide du serveur',
      );
    }
    
    // Handle web platform _JsonMap type safely
    final success = json['success'] as bool? ?? false;
    final data = json['data'];
    final errorData = json['error'];
    final message = json['message'] as String?;
    
    ApiError? error;
    if (errorData != null && errorData is Map) {
      error = ApiError.fromJson(errorData as Map<String, dynamic>);
    }
    
    return ApiResponse<T>(
      success: success,
      data: data,
      error: error,
      message: message,
    );
  }
  
  factory ApiResponse.success(T data, {String? message}) {
    return ApiResponse<T>(
      success: true,
      data: data,
      message: message,
    );
  }
  
  factory ApiResponse.error(String message, {String? code}) {
    return ApiResponse<T>(
      success: false,
      error: ApiError(code: code ?? 'error', message: message),
      message: message,
    );
  }
}

/// Modèle d'erreur API
class ApiError {
  final String code;
  final String message;
  final Map<String, dynamic>? details;
  
  ApiError({
    required this.code,
    required this.message,
    this.details,
  });
  
  factory ApiError.fromJson(Map<String, dynamic> json) {
    return ApiError(
      code: json['code'] ?? 'unknown',
      message: json['message'] ?? 'Unknown error',
      details: json['details'],
    );
  }
}

/// Modèle de réponse paginée
class PaginatedResponse<T> {
  final List<T> data;
  final int page;
  final int limit;
  final int total;
  final int totalPages;
  
  PaginatedResponse({
    required this.data,
    required this.page,
    required this.limit,
    required this.total,
    required this.totalPages,
  });
  
  factory PaginatedResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Map<String, dynamic>) fromJsonT,
  ) {
    final List<dynamic> dataList = json['data'] ?? [];
    return PaginatedResponse<T>(
      data: dataList.map((e) => fromJsonT(e as Map<String, dynamic>)).toList(),
      page: json['page'] ?? 1,
      limit: json['limit'] ?? 20,
      total: json['total'] ?? 0,
      totalPages: json['total_pages'] ?? 0,
    );
  }
}

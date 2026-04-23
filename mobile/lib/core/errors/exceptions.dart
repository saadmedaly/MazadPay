// lib/core/errors/exceptions.dart
// Exceptions personnalisées pour l'app MazadPay

abstract class AppException implements Exception {
  final String message;
  final int? statusCode;
  final String? code;
  
  const AppException(this.message, {this.statusCode, this.code});
  
  @override
  String toString() => '[$runtimeType] $message (Code: $statusCode)';
}

class NetworkException extends AppException {
  const NetworkException([String message = 'Network error']) 
    : super(message, code: 'NETWORK_ERROR');
}

class ServerException extends AppException {
  const ServerException([String message = 'Server error', int? statusCode]) 
    : super(message, statusCode: statusCode, code: 'SERVER_ERROR');
}

class TimeoutException extends AppException {
  const TimeoutException([String message = 'Request timeout']) 
    : super(message, code: 'TIMEOUT');
}

class UnauthorizedException extends AppException {
  const UnauthorizedException([String message = 'Unauthorized']) 
    : super(message, statusCode: 401, code: 'UNAUTHORIZED');
}

class NotFoundException extends AppException {
  const NotFoundException([String message = 'Resource not found']) 
    : super(message, statusCode: 404, code: 'NOT_FOUND');
}

class ValidationException extends AppException {
  final Map<String, dynamic>? errors;
  
  const ValidationException(String message, {this.errors}) 
    : super(message, statusCode: 400, code: 'VALIDATION_ERROR');
}

class ConflictException extends AppException {
  const ConflictException([String message = 'Conflict']) 
    : super(message, statusCode: 409, code: 'CONFLICT');
}

class BadRequestException extends AppException {
  const BadRequestException([String message = 'Bad request']) 
    : super(message, statusCode: 400, code: 'BAD_REQUEST');
}

class UnknownException extends AppException {
  const UnknownException([String message = 'Unknown error']) 
    : super(message, code: 'UNKNOWN');
}

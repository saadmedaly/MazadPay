import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:mezadpay/services/api_service.dart';

// dioExceptionTypeCode() is the pure mapping used by ApiService._handleError
// to populate ApiException.code, which CategoryApi.getCountries() and
// AuthApi's public methods forward into ApiResponse's ApiError.code. It
// splits every DioExceptionType into a distinct category string rather than
// collapsing timeouts and several other types into a single generic
// 'connection_error'/'unknown_error' value. These tests only cover the pure
// enum-to-string mapping -- no Dio network call or mock server needed.
void main() {
  group('dioExceptionTypeCode maps every DioExceptionType to a distinct, stable category', () {
    test('connectionTimeout', () {
      expect(dioExceptionTypeCode(DioExceptionType.connectionTimeout), 'connectionTimeout');
    });
    test('sendTimeout', () {
      expect(dioExceptionTypeCode(DioExceptionType.sendTimeout), 'sendTimeout');
    });
    test('receiveTimeout', () {
      expect(dioExceptionTypeCode(DioExceptionType.receiveTimeout), 'receiveTimeout');
    });
    test('connectionError', () {
      expect(dioExceptionTypeCode(DioExceptionType.connectionError), 'connectionError');
    });
    test('badCertificate', () {
      expect(dioExceptionTypeCode(DioExceptionType.badCertificate), 'badCertificate');
    });
    test('badResponse', () {
      expect(dioExceptionTypeCode(DioExceptionType.badResponse), 'badResponse');
    });
    test('cancel', () {
      expect(dioExceptionTypeCode(DioExceptionType.cancel), 'cancel');
    });
    test('unknown', () {
      expect(dioExceptionTypeCode(DioExceptionType.unknown), 'unknown');
    });

    test('every category is unique (no two DioExceptionTypes collapse together)', () {
      final codes = DioExceptionType.values.map(dioExceptionTypeCode).toSet();
      expect(codes.length, DioExceptionType.values.length);
    });
  });
}

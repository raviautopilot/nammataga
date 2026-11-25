import 'package:dio/dio.dart';
import '../constants/api_endpoints.dart';
import '../storage/secure_storage_service.dart';

/// Centralized Dio client with JWT interceptor and error formatting
class ApiClient {
  ApiClient._();

  static Dio? _dioInstance;

  static Dio get instance {
    _dioInstance ??= _createDio();
    return _dioInstance!;
  }

  static Dio _createDio() {
    final dio = Dio(
      BaseOptions(
        baseUrl: ApiEndpoints.baseUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 15),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ),
    );

    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await StorageService.getToken();
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          return handler.next(options);
        },
        onError: (DioException error, handler) async {
          final path = error.requestOptions.path;
          final isPasswordValidation = path.contains('/account-deletion') ||
              path.contains('/login') ||
              path.contains('/change-password');

          if (error.response?.statusCode == 401 && !isPasswordValidation) {
            // Auto-clear auth storage on 401 unauthorized
            await StorageService.clearAuth();
          }

          final formattedError = _formatNetworkError(error);
          return handler.next(formattedError);
        },
      ),
    );

    return dio;
  }

  /// Replaces raw Dio connection/timeout messages with user-friendly offline messages.
  static DioException _formatNetworkError(DioException error) {
    if (error.type == DioExceptionType.connectionTimeout ||
        error.type == DioExceptionType.sendTimeout ||
        error.type == DioExceptionType.receiveTimeout) {
      return error.copyWith(
        message:
            'Connection timed out. Please check your internet connection and try again.',
      );
    }

    final rawMsg = '${error.error ?? ''} ${error.message ?? ''}'.toLowerCase();
    final isOffline = error.type == DioExceptionType.connectionError ||
        rawMsg.contains('socketexception') ||
        rawMsg.contains('failed host lookup') ||
        rawMsg.contains('network is unreachable') ||
        rawMsg.contains('connection refused') ||
        rawMsg.contains('connection reset');

    if (isOffline) {
      return error.copyWith(
        message:
            'No internet connection. Please check your Wi-Fi or mobile data and try again.',
      );
    }

    return error;
  }
}

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
          if (error.response?.statusCode == 401) {
            // Auto-clear auth storage on 401 unauthorized
            await StorageService.clearAuth();
          }
          return handler.next(error);
        },
      ),
    );

    return dio;
  }
}

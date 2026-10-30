import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../domain/grievance_model.dart';
import '../domain/grievance_priority_model.dart';

/// Repository handling all network requests for the Grievance feature.
class GrievanceRepository {
  final Dio _dio;

  GrievanceRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  /// Fetches the list of valid grievance categories from the backend.
  /// Endpoint: GET /api/categories
  Future<List<String>> getCategories() async {
    try {
      final response = await _dio.get(ApiEndpoints.grievanceCategories);
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .map((item) => item.toString().trim())
            .where((item) => item.isNotEmpty)
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to fetch grievance categories'));
    } catch (e) {
      throw Exception('Failed to fetch grievance categories: $e');
    }
  }

  /// Fetches the list of available grievance priorities with label and value.
  /// Endpoint: GET /api/priorities
  Future<List<GrievancePriorityModel>> getPriorities() async {
    try {
      final response = await _dio.get(ApiEndpoints.grievancePriorities);
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List).map((item) {
          if (item is Map) {
            return GrievancePriorityModel.fromJson(Map<String, dynamic>.from(item));
          }
          final str = item.toString().trim();
          return GrievancePriorityModel(label: str, value: str.toLowerCase());
        }).toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to fetch grievance priorities'));
    } catch (e) {
      throw Exception('Failed to fetch grievance priorities: $e');
    }
  }

  /// Submits a new grievance petition to the backend.
  /// Endpoint: POST /api/grievances
  Future<GrievanceModel> submitGrievance({
    required String subject,
    required String category,
    required String priority,
    required String description,
    String? contactPhone,
    required String preferredResponse,
  }) async {
    try {
      final payload = <String, dynamic>{
        'subject': subject.trim(),
        'category': category.trim(),
        'priority': priority.trim(),
        'description': description.trim(),
        'contactPhone': contactPhone?.trim() ?? '',
        'preferredResponse': preferredResponse.trim(),
      };

      final response = await _dio.post(
        ApiEndpoints.grievances,
        data: payload,
      );

      if ((response.statusCode == 200 || response.statusCode == 201) &&
          response.data is Map) {
        return GrievanceModel.fromJson(
          Map<String, dynamic>.from(response.data as Map),
        );
      }

      throw Exception('Unexpected server response: ${response.statusCode}');
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to submit grievance'));
    } catch (e) {
      if (e is Exception) rethrow;
      throw Exception('Failed to submit grievance: $e');
    }
  }

  /// Fetches submitted grievances for the authenticated member.
  /// Endpoint: GET /api/grievances
  Future<List<GrievanceModel>> getMemberGrievances() async {
    try {
      final response = await _dio.get(ApiEndpoints.grievances);
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .whereType<Map>()
            .map((item) => GrievanceModel.fromJson(Map<String, dynamic>.from(item)))
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to fetch member grievances'));
    } catch (e) {
      throw Exception('Failed to fetch member grievances: $e');
    }
  }

  /// Extracts readable error messages from DioException responses.
  String _extractErrorMessage(DioException e, String fallback) {
    if (e.response?.data is Map) {
      final data = e.response!.data as Map;
      final msg = data['error'] ?? data['message'] ?? data['msg'];
      if (msg != null && msg.toString().trim().isNotEmpty) {
        return msg.toString();
      }
    } else if (e.response?.data is String &&
        (e.response!.data as String).trim().isNotEmpty) {
      return e.response!.data as String;
    }
    if (e.message != null && e.message!.isNotEmpty) {
      return e.message!;
    }
    return fallback;
  }
}

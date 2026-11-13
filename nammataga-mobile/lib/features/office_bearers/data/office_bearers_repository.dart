import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../domain/district_office_bearer.dart';
import '../domain/state_executive_member.dart';

/// Repository interacting with Office Bearers endpoints
class OfficeBearersRepository {
  final Dio _dio;

  OfficeBearersRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  /// Fetches State Executive committee members preserving server order
  Future<List<StateExecutiveMember>> getStateExecutive() async {
    try {
      final response = await _dio.get(ApiEndpoints.stateExecutive);
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .map((item) =>
                StateExecutiveMember.fromJson(item as Map<String, dynamic>))
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load state executive committee',
      );
    }
  }

  /// Fetches list of all districts
  Future<List<String>> getDistricts() async {
    try {
      final response = await _dio.get(ApiEndpoints.bearerDistricts);
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .map((item) => item.toString().trim())
            .where((item) => item.isNotEmpty)
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load districts',
      );
    }
  }

  /// Fetches district office bearers for a specific district
  Future<List<DistrictOfficeBearer>> getDistrictBearers(String district) async {
    try {
      final response = await _dio.get(ApiEndpoints.districtBearers(district));
      if (response.statusCode == 200) {
        if (response.data is List) {
          return (response.data as List)
              .map((item) =>
                  DistrictOfficeBearer.fromJson(item as Map<String, dynamic>))
              .toList();
        } else if (response.data is Map) {
          // In case API returns { "DistrictName": [...] }
          final map = response.data as Map;
          final list = map[district] ?? map.values.firstOrNull;
          if (list is List) {
            return list
                .map((item) =>
                    DistrictOfficeBearer.fromJson(item as Map<String, dynamic>))
                .toList();
          }
        }
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load district office bearers for $district',
      );
    }
  }
}

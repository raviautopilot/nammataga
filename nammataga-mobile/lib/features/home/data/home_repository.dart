import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../domain/association_info.dart';

class HomeRepository {
  final Dio _dio;

  HomeRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  Map<String, dynamic>? _toMap(dynamic data) {
    if (data is Map<String, dynamic>) return data;
    if (data is Map) return Map<String, dynamic>.from(data);
    if (data is String) {
      try {
        final decoded = jsonDecode(data);
        if (decoded is Map<String, dynamic>) return decoded;
        if (decoded is Map) return Map<String, dynamic>.from(decoded);
      } catch (_) {}
    }
    return null;
  }

  List<dynamic>? _toList(dynamic data) {
    if (data is List) return data;
    if (data is String) {
      try {
        final decoded = jsonDecode(data);
        if (decoded is List) return decoded;
      } catch (_) {}
    }
    return null;
  }

  /// Fetches core association information from GET /api/public/about
  Future<AboutData> getAboutInfo() async {
    try {
      final response = await _dio.get(ApiEndpoints.about);
      final map = _toMap(response.data);
      if (map != null) {
        return AboutData.fromJson(map);
      }
    } catch (e) {
      debugPrint('[HomeRepository] Error fetching about info from backend: $e');
      rethrow;
    }
    throw Exception('Failed to load about data from backend API');
  }

  /// Fetches association statistics counters from GET /api/public/about/stats
  Future<List<StatData>> getStats() async {
    try {
      final response = await _dio.get(ApiEndpoints.aboutStats);
      final list = _toList(response.data);
      if (list != null) {
        return list
            .map((e) => StatData.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    } catch (e) {
      debugPrint('[HomeRepository] Error fetching stats from backend: $e');
      rethrow;
    }
    throw Exception('Failed to load stats data from backend API');
  }

  /// Fetches association strategic objectives from GET /api/public/about/objectives
  Future<List<ObjectiveData>> getObjectives() async {
    try {
      final response = await _dio.get(ApiEndpoints.aboutObjectives);
      final list = _toList(response.data);
      if (list != null) {
        return list
            .map((e) => ObjectiveData.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    } catch (e) {
      debugPrint('[HomeRepository] Error fetching objectives from backend: $e');
      rethrow;
    }
    throw Exception('Failed to load objectives data from backend API');
  }

  /// Fetches headquarters and regional contact details from GET /api/public/about/contact
  Future<ContactData> getContactInfo() async {
    try {
      final response = await _dio.get(ApiEndpoints.aboutContact);
      final map = _toMap(response.data);
      if (map != null) {
        return ContactData.fromJson(map);
      }
    } catch (e) {
      debugPrint(
        '[HomeRepository] Error fetching contact info from backend: $e',
      );
      rethrow;
    }
    throw Exception('Failed to load contact info from backend API');
  }
}

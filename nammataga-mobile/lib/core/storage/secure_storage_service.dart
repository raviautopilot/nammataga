import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Unified storage service wrapping secure storage for sensitive tokens
/// and SharedPreferences for simple non-sensitive preference flags.
class StorageService {
  StorageService._();

  static const _secureStorage = FlutterSecureStorage();

  static const String keyMemberToken = 'member_token';
  static const String keyTokenExpiry = 'member_token_expiry';
  static const String keyMemberRole = 'member_role';
  static const String keyUserData = 'user_data';
  static const String keyIsGuest = 'is_guest_mode';
  static const String keyRememberEmail = 'remember_email';

  // Secure token methods
  static Future<void> saveToken(String token, {int? expiresInSeconds}) async {
    await _secureStorage.write(key: keyMemberToken, value: token);
    if (expiresInSeconds != null && expiresInSeconds > 0) {
      final expiryMs =
          DateTime.now().millisecondsSinceEpoch + (expiresInSeconds * 1000);
      await _secureStorage.write(
        key: keyTokenExpiry,
        value: expiryMs.toString(),
      );
    }
  }

  static Future<String?> getToken() async {
    return await _secureStorage.read(key: keyMemberToken);
  }

  static Future<bool> isTokenValid() async {
    final token = await getToken();
    if (token == null || token.isEmpty) return false;

    final expiryStr = await _secureStorage.read(key: keyTokenExpiry);
    if (expiryStr == null) return true;

    final expiryMs = int.tryParse(expiryStr);
    if (expiryMs == null) return true;

    // 5-minute buffer matching web app grace period
    const graceMs = 5 * 60 * 1000;
    return (DateTime.now().millisecondsSinceEpoch - graceMs) < expiryMs;
  }

  static Future<void> saveUserData(
    String jsonString, {
    String role = 'member',
  }) async {
    await _secureStorage.write(key: keyUserData, value: jsonString);
    await _secureStorage.write(key: keyMemberRole, value: role);
  }

  static Future<String?> getUserData() async {
    return await _secureStorage.read(key: keyUserData);
  }

  static Future<void> clearAuth() async {
    await _secureStorage.delete(key: keyMemberToken);
    await _secureStorage.delete(key: keyTokenExpiry);
    await _secureStorage.delete(key: keyMemberRole);
    await _secureStorage.delete(key: keyUserData);
  }

  // Non-secure preferences
  static Future<void> setGuestMode(bool isGuest) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(keyIsGuest, isGuest);
  }

  static Future<bool> isGuestMode() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(keyIsGuest) ?? false;
  }

  static Future<void> setRememberEmail(String email) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(keyRememberEmail, email);
  }

  static Future<String?> getRememberEmail() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(keyRememberEmail);
  }
}

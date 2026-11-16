import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../../core/widgets/taga_toast.dart';

/// Clean utility for dialing, messaging, emailing
class ContactActionsHelper {
  ContactActionsHelper._();

  /// Clean digits and +
  static String sanitizePhone(String phone) {
    return phone.replaceAll(RegExp(r'[^0-9+]'), '');
  }

  /// Whether valid phone
  static bool hasValidPhone(String? phone) {
    if (phone == null || phone.trim().isEmpty) return false;
    final clean = sanitizePhone(phone);
    return clean.length >= 7;
  }

  /// Whether valid email
  static bool hasValidEmail(String? email) {
    if (email == null || email.trim().isEmpty) return false;
    return RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$').hasMatch(email.trim());
  }

  /// Call a phone number
  static Future<bool> makePhoneCall(BuildContext context, String phone) async {
    final clean = sanitizePhone(phone);
    if (clean.isEmpty) return false;
    final uri = Uri.parse('tel:$clean');
    try {
      if (await canLaunchUrl(uri)) {
        return await launchUrl(uri);
      }
    } catch (_) {}
    if (context.mounted) {
      TagaToast.error(context, 'Could not open phone dialer for $phone');
    }
    return false;
  }

  /// Send email
  static Future<bool> sendEmail(BuildContext context, String email) async {
    final clean = email.trim();
    if (clean.isEmpty) return false;
    final uri = Uri.parse('mailto:$clean');
    try {
      if (await canLaunchUrl(uri)) {
        return await launchUrl(uri);
      }
    } catch (_) {}
    if (context.mounted) {
      TagaToast.error(context, 'Could not open email app for $email');
    }
    return false;
  }

  /// Open WhatsApp conversation
  static Future<bool> openWhatsApp(BuildContext context, String phone) async {
    var clean = sanitizePhone(phone);
    if (clean.isEmpty) return false;
    // Prepend 91 if 10-digit Indian number without country code
    if (clean.length == 10 && !clean.startsWith('+')) {
      clean = '91$clean';
    } else if (clean.startsWith('+')) {
      clean = clean.substring(1);
    }
    final uri = Uri.parse('https://wa.me/$clean');
    try {
      if (await canLaunchUrl(uri)) {
        return await launchUrl(uri, mode: LaunchMode.externalApplication);
      }
    } catch (_) {}
    if (context.mounted) {
      TagaToast.error(context, 'Could not launch WhatsApp for $phone');
    }
    return false;
  }
}

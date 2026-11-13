import '../../../../core/constants/api_endpoints.dart';

/// Model representing a State Executive member of TAGA
class StateExecutiveMember {
  final int sno;
  final String name;
  final String designation;
  final String phone;
  final String location;
  final String qualification;
  final int experience;
  final String description;
  final String? image;

  const StateExecutiveMember({
    required this.sno,
    required this.name,
    required this.designation,
    required this.phone,
    required this.location,
    required this.qualification,
    required this.experience,
    required this.description,
    this.image,
  });

  factory StateExecutiveMember.fromJson(Map<String, dynamic> json) {
    return StateExecutiveMember(
      sno: json['sno'] is int
          ? json['sno'] as int
          : int.tryParse(json['sno']?.toString() ?? '0') ?? 0,
      name: (json['name'] as String?)?.trim() ?? '',
      designation: (json['designation'] as String?)?.trim() ?? '',
      phone: (json['phone'] as String?)?.trim() ?? '',
      location: (json['location'] as String?)?.trim() ?? '',
      qualification: (json['qualification'] as String?)?.trim() ?? '',
      experience: json['experience'] is int
          ? json['experience'] as int
          : int.tryParse(json['experience']?.toString() ?? '0') ?? 0,
      description: (json['description'] as String?)?.trim() ?? '',
      image: (json['image'] as String?)?.trim(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'sno': sno,
      'name': name,
      'designation': designation,
      'phone': phone,
      'location': location,
      'qualification': qualification,
      'experience': experience,
      'description': description,
      if (image != null) 'image': image,
    };
  }

  /// Full absolute image URL safely resolved
  String? get resolvedImageUrl {
    if (image == null || image!.isEmpty) return null;
    final img = image!.trim();
    if (img.startsWith('http://') || img.startsWith('https://')) {
      return img;
    }
    final cleanBase = ApiEndpoints.baseUrl.replaceAll(RegExp(r'/api/?$'), '');
    var path = img.startsWith('/') ? img : '/$img';
    if (!path.startsWith('/api/') && path.startsWith('/images/')) {
      path = '/api$path';
    }
    return '$cleanBase$path';
  }

  String get initials {
    if (name.isEmpty) return 'T';
    final parts = name.split(RegExp(r'\s+')).where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return 'T';
    if (parts.length == 1) return parts[0][0].toUpperCase();
    return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
  }
}

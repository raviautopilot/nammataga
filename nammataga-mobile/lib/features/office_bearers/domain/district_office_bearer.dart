/// Model representing a District Office Bearer of TAGA
class DistrictOfficeBearer {
  final String name;
  final String title;
  final String department;
  final String contact;

  const DistrictOfficeBearer({
    required this.name,
    required this.title,
    required this.department,
    required this.contact,
  });

  factory DistrictOfficeBearer.fromJson(Map<String, dynamic> json) {
    return DistrictOfficeBearer(
      name: (json['name'] as String?)?.trim() ?? '',
      title: (json['title'] as String?)?.trim() ?? '',
      department: (json['department'] as String?)?.trim() ?? '',
      contact: (json['contact'] as String?)?.trim() ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'title': title,
      'department': department,
      'contact': contact,
    };
  }

  String get initials {
    if (name.isEmpty) return 'D';
    final parts = name.split(RegExp(r'\s+')).where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return 'D';
    if (parts.length == 1) return parts[0][0].toUpperCase();
    return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
  }
}

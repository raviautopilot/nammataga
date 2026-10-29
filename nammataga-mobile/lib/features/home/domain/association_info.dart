library;

/// Domain models representing public association information,
/// statistics, objectives, and contact details from backend API endpoints.

class AboutData {
  final int id;
  final String name;
  final String acronym;
  final int establishedYear;
  final String tagline;
  final String heroImageUrl;
  final String mission;
  final String vision;
  final String description;

  const AboutData({
    required this.id,
    required this.name,
    required this.acronym,
    required this.establishedYear,
    required this.tagline,
    required this.heroImageUrl,
    required this.mission,
    required this.vision,
    required this.description,
  });

  factory AboutData.fromJson(Map<String, dynamic> json) {
    return AboutData(
      id: json['id'] as int? ?? 0,
      name: (json['name'] as String? ?? '').trim(),
      acronym: (json['acronym'] as String? ?? '').trim(),
      establishedYear: json['established_year'] as int? ?? 0,
      tagline: (json['tagline'] as String? ?? '').trim(),
      heroImageUrl: (json['hero_image_url'] as String? ?? '').trim(),
      mission: (json['mission'] as String? ?? '').trim(),
      vision: (json['vision'] as String? ?? '').trim(),
      description: (json['description'] as String? ?? '').trim(),
    );
  }
}

class StatData {
  final int id;
  final String label;
  final int value;
  final String iconName;
  final String colorClass;

  const StatData({
    required this.id,
    required this.label,
    required this.value,
    this.iconName = '',
    this.colorClass = '',
  });

  factory StatData.fromJson(Map<String, dynamic> json) {
    return StatData(
      id: json['id'] as int? ?? 0,
      label: (json['label'] as String? ?? '').trim(),
      value: json['value'] as int? ?? 0,
      iconName: (json['icon_name'] as String? ?? '').trim(),
      colorClass: (json['color_class'] as String? ?? '').trim(),
    );
  }
}

class ObjectiveData {
  final int id;
  final String title;
  final String description;
  final int orderIndex;

  const ObjectiveData({
    required this.id,
    required this.title,
    required this.description,
    required this.orderIndex,
  });

  factory ObjectiveData.fromJson(Map<String, dynamic> json) {
    var rawDesc = (json['description'] as String? ?? '').trim();
    var rawTitle = (json['title'] as String? ?? '').trim();

    // If title is empty in the backend, split "Title: Description" prefix
    if (rawTitle.isEmpty && rawDesc.contains(':')) {
      final parts = rawDesc.split(':');
      rawTitle = parts[0].trim();
      rawDesc = parts.sublist(1).join(':').trim();
    }

    return ObjectiveData(
      id: json['id'] as int? ?? 0,
      title: rawTitle,
      description: rawDesc,
      orderIndex: json['order_index'] as int? ?? 0,
    );
  }
}

class ContactData {
  final String primaryPhone;
  final String primaryEmail;
  final String secondaryEmail;
  final String website;
  final String headquartersName;
  final String headquartersAddress;
  final String officeHours;
  final List<RegionalOffice> regionalOffices;

  const ContactData({
    required this.primaryPhone,
    required this.primaryEmail,
    required this.secondaryEmail,
    required this.website,
    required this.headquartersName,
    required this.headquartersAddress,
    required this.officeHours,
    required this.regionalOffices,
  });

  factory ContactData.fromJson(Map<String, dynamic> json) {
    final hq = json['headquarters'] as Map<String, dynamic>? ?? {};
    final officesRaw = json['regional_offices'] as List<dynamic>? ?? [];

    return ContactData(
      primaryPhone: (json['primary_phone'] as String? ?? '').trim(),
      primaryEmail: (json['primary_email'] as String? ?? '').trim(),
      secondaryEmail: (json['secondary_email'] as String? ?? '').trim(),
      website: (json['website'] as String? ?? '').trim(),
      headquartersName: (hq['name'] as String? ?? '').trim(),
      headquartersAddress: (hq['address'] as String? ?? '').trim(),
      officeHours: (json['office_hours'] as String? ?? '').trim(),
      regionalOffices: officesRaw
          .map((e) => RegionalOffice.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

class RegionalOffice {
  final int id;
  final String city;
  final String phone;
  final String address;

  const RegionalOffice({
    required this.id,
    required this.city,
    required this.phone,
    required this.address,
  });

  factory RegionalOffice.fromJson(Map<String, dynamic> json) {
    return RegionalOffice(
      id: json['id'] as int? ?? 0,
      city: (json['city'] as String? ?? '').trim(),
      phone: (json['phone'] as String? ?? '').trim(),
      address: (json['address'] as String? ?? '').trim(),
    );
  }
}

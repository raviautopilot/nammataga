import '../../../core/constants/api_endpoints.dart';

/// Domain model representing a NammaTAGA resource document.
class ResourceDocument {
  final String title;
  final String year;
  final String? subcategory;
  final String? url;

  const ResourceDocument({
    required this.title,
    required this.year,
    this.subcategory,
    this.url,
  });

  /// Factory constructor to parse document JSON payloads from backend.
  factory ResourceDocument.fromJson(Map<String, dynamic> json) {
    return ResourceDocument(
      title: json['title']?.toString() ?? '',
      year: json['year']?.toString() ?? '',
      subcategory: json['subcategory']?.toString(),
      url: json['url']?.toString(),
    );
  }

  /// Converts this document instance to a JSON-encodable map.
  Map<String, dynamic> toJson() {
    return {
      'title': title,
      'year': year,
      if (subcategory != null) 'subcategory': subcategory,
      if (url != null) 'url': url,
    };
  }

  /// Resolves relative document paths to absolute download / stream URLs.
  String get fullUrl =>
      url != null && url!.isNotEmpty ? ApiEndpoints.resolveDocUrl(url!) : '';

  /// Whether this document is a PDF file based on its URL file extension.
  bool get isPdf =>
      url != null &&
      url!.toLowerCase().split('?').first.trim().endsWith('.pdf');

  /// Creates a copy of this document with optional updated fields.
  ResourceDocument copyWith({
    String? title,
    String? year,
    String? subcategory,
    String? url,
  }) {
    return ResourceDocument(
      title: title ?? this.title,
      year: year ?? this.year,
      subcategory: subcategory ?? this.subcategory,
      url: url ?? this.url,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ResourceDocument &&
          runtimeType == other.runtimeType &&
          title == other.title &&
          year == other.year &&
          subcategory == other.subcategory &&
          url == other.url;

  @override
  int get hashCode => Object.hash(title, year, subcategory, url);

  @override
  String toString() =>
      'ResourceDocument(title: $title, year: $year, subcategory: $subcategory, url: $url)';
}

/// Sorts documents matching website logic (`taga-web/src/api/resources.ts`):
/// - Newest year first (descending), using first 4 characters of year parsed to int,
///   falling back to current year if missing.
/// - Same year: alphabetical ascending by title (case-insensitive).
List<ResourceDocument> sortDocuments(List<ResourceDocument> docs) {
  final currentYear = DateTime.now().year.toString();
  final currentYearNum = int.tryParse(currentYear) ?? DateTime.now().year;

  final sorted = List<ResourceDocument>.from(docs);
  sorted.sort((a, b) {
    // Get years, fallback to current year if missing
    final trimmedA = a.year.trim().isEmpty ? currentYear : a.year.trim();
    final trimmedB = b.year.trim().isEmpty ? currentYear : b.year.trim();

    final yearA = trimmedA.length >= 4 ? trimmedA.substring(0, 4) : trimmedA;
    final yearB = trimmedB.length >= 4 ? trimmedB.substring(0, 4) : trimmedB;

    // Convert to numbers for comparison
    final yearNumA = int.tryParse(yearA) ?? currentYearNum;
    final yearNumB = int.tryParse(yearB) ?? currentYearNum;

    // Sort by year descending (newest first)
    if (yearNumB != yearNumA) {
      return yearNumB.compareTo(yearNumA);
    }

    // Same year: sort alphabetically, case-insensitive
    final titleA = a.title.toLowerCase().trim();
    final titleB = b.title.toLowerCase().trim();

    return titleA.compareTo(titleB);
  });

  return sorted;
}

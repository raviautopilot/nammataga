/// Domain model representing a useful external link for NammaTAGA members.
class ExternalLink {
  final String title;
  final String url;

  const ExternalLink({
    required this.title,
    required this.url,
  });

  /// Factory constructor to parse link JSON payloads from backend.
  factory ExternalLink.fromJson(Map<String, dynamic> json) {
    return ExternalLink(
      title: json['title']?.toString() ?? '',
      url: json['url']?.toString() ?? '',
    );
  }

  /// Converts this link instance to a JSON-encodable map.
  Map<String, dynamic> toJson() {
    return {
      'title': title,
      'url': url,
    };
  }

  /// Creates a copy of this external link with optional updated fields.
  ExternalLink copyWith({
    String? title,
    String? url,
  }) {
    return ExternalLink(
      title: title ?? this.title,
      url: url ?? this.url,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ExternalLink &&
          runtimeType == other.runtimeType &&
          title == other.title &&
          url == other.url;

  @override
  int get hashCode => Object.hash(title, url);

  @override
  String toString() => 'ExternalLink(title: $title, url: $url)';
}

/// Sorts external links alphabetically ascending by title (case-insensitive).
List<ExternalLink> sortExternalLinks(List<ExternalLink> links) {
  final sorted = List<ExternalLink>.from(links);
  sorted.sort((a, b) {
    final titleA = a.title.toLowerCase().trim();
    final titleB = b.title.toLowerCase().trim();
    return titleA.compareTo(titleB);
  });
  return sorted;
}

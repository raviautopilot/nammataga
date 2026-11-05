import 'package:flutter/foundation.dart';

/// Domain model representing a NammaTAGA resource category.
class ResourceCategory {
  final String id;
  final String name;
  final List<String>? subcategories;

  const ResourceCategory({
    required this.id,
    required this.name,
    this.subcategories,
  });

  /// Factory constructor to parse category JSON payloads from backend.
  factory ResourceCategory.fromJson(Map<String, dynamic> json) {
    List<String>? subcats;
    if (json['subcategories'] is List) {
      subcats = (json['subcategories'] as List<dynamic>)
          .map((e) => e.toString())
          .toList();
    }

    return ResourceCategory(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      subcategories: subcats,
    );
  }

  /// Converts this instance to a JSON-encodable map.
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      if (subcategories != null) 'subcategories': subcategories,
    };
  }

  /// Creates a copy of this category with optional updated fields.
  ResourceCategory copyWith({
    String? id,
    String? name,
    List<String>? subcategories,
  }) {
    return ResourceCategory(
      id: id ?? this.id,
      name: name ?? this.name,
      subcategories: subcategories ?? this.subcategories,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ResourceCategory &&
          runtimeType == other.runtimeType &&
          id == other.id &&
          name == other.name &&
          listEquals(subcategories, other.subcategories);

  @override
  int get hashCode => Object.hash(
        id,
        name,
        subcategories == null ? null : Object.hashAll(subcategories!),
      );

  @override
  String toString() =>
      'ResourceCategory(id: $id, name: $name, subcategories: $subcategories)';
}

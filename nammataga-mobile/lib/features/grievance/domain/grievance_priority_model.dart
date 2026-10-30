/// Domain model representing a Grievance priority option.
class GrievancePriorityModel {
  final String label;
  final String value;

  const GrievancePriorityModel({
    required this.label,
    required this.value,
  });

  /// Creates a [GrievancePriorityModel] from a backend JSON map.
  /// Handles { "label": "Low", "value": "low" } format.
  factory GrievancePriorityModel.fromJson(Map<String, dynamic> json) {
    return GrievancePriorityModel(
      label: json['label']?.toString() ?? '',
      value: json['value']?.toString() ?? '',
    );
  }

  /// Serializes back to a JSON-compatible map.
  Map<String, dynamic> toJson() {
    return {
      'label': label,
      'value': value,
    };
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is GrievancePriorityModel &&
          runtimeType == other.runtimeType &&
          value.toLowerCase() == other.value.toLowerCase();

  @override
  int get hashCode => value.toLowerCase().hashCode;

  @override
  String toString() => 'GrievancePriorityModel(label: $label, value: $value)';
}

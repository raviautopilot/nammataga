library;

/// Domain model representing a member's complete profile
/// strictly matching the Go backend `model.Member` and React `Member` interface.
class MemberProfile {
  final String id;
  final String tagaId;
  final String username;
  final String name;
  final String initial;
  final String gender;
  final String dateOfBirth;
  final String fatherName;
  final String motherName;
  final String mobileNumber;
  final String emailId;
  final String designation;
  final String recruitmentBatch;
  final String seniorityNumber;
  final String workingDistrict;
  final String nativeDistrict;
  final String residentialAddress;
  final String permanentAddress;
  final String educationalQualification;
  final String tbfNumber;
  final String cpsGpfNumber;
  final String paymentStatus;
  final bool isPaid;
  final bool subscriptionActive;

  const MemberProfile({
    required this.id,
    required this.tagaId,
    required this.username,
    required this.name,
    required this.initial,
    required this.gender,
    required this.dateOfBirth,
    required this.fatherName,
    required this.motherName,
    required this.mobileNumber,
    required this.emailId,
    required this.designation,
    required this.recruitmentBatch,
    required this.seniorityNumber,
    required this.workingDistrict,
    required this.nativeDistrict,
    required this.residentialAddress,
    required this.permanentAddress,
    required this.educationalQualification,
    required this.tbfNumber,
    required this.cpsGpfNumber,
    required this.paymentStatus,
    required this.isPaid,
    required this.subscriptionActive,
  });

  /// Derives 1-2 initials for avatar monogram (e.g., "Ramesh Kumar" -> "RK", or "Ramesh" + initial "S" -> "RS")
  String get avatarInitials {
    final cleanInitial = initial.trim().toUpperCase();
    final cleanName = name.trim();

    if (cleanName.isEmpty) return 'T';

    final parts = cleanName
        .split(RegExp(r'\s+'))
        .where((p) => p.isNotEmpty)
        .toList();

    if (cleanInitial.isNotEmpty) {
      final firstLetter = cleanName.substring(0, 1).toUpperCase();
      final lastInitial = cleanInitial.replaceAll('.', '').substring(0, 1);
      return '$firstLetter$lastInitial';
    }

    if (parts.length >= 2) {
      final first = parts[0].substring(0, 1).toUpperCase();
      final second = parts[1].substring(0, 1).toUpperCase();
      return '$first$second';
    }

    return cleanName.length >= 2
        ? cleanName.substring(0, 2).toUpperCase()
        : cleanName.substring(0, 1).toUpperCase();
  }

  /// Full display name combining name and initial
  String get fullNameWithInitial {
    final cleanName = name.trim();
    final cleanInit = initial.trim();
    if (cleanInit.isEmpty) return cleanName;
    return '$cleanName $cleanInit';
  }

  /// Membership type title
  String get membershipBadgeTitle {
    return (isPaid || subscriptionActive)
        ? 'Subscriber Member'
        : 'General Member';
  }

  /// Formatted date of birth (e.g., "15 Jan 1985")
  String get formattedDob {
    if (dateOfBirth.isEmpty) return '—';
    try {
      DateTime? parsed;
      if (dateOfBirth.contains('-')) {
        final parts = dateOfBirth.split('-');
        if (parts.length == 3) {
          if (parts[0].length == 4) {
            // YYYY-MM-DD
            parsed = DateTime(
              int.parse(parts[0]),
              int.parse(parts[1]),
              int.parse(parts[2]),
            );
          } else {
            // DD-MM-YYYY
            parsed = DateTime(
              int.parse(parts[2]),
              int.parse(parts[1]),
              int.parse(parts[0]),
            );
          }
        }
      } else if (dateOfBirth.contains('/')) {
        final parts = dateOfBirth.split('/');
        if (parts.length == 3) {
          if (parts[2].length == 4) {
            // DD/MM/YYYY
            parsed = DateTime(
              int.parse(parts[2]),
              int.parse(parts[1]),
              int.parse(parts[0]),
            );
          } else {
            // YYYY/MM/DD
            parsed = DateTime(
              int.parse(parts[0]),
              int.parse(parts[1]),
              int.parse(parts[2]),
            );
          }
        }
      }

      if (parsed != null) {
        const months = [
          'Jan',
          'Feb',
          'Mar',
          'Apr',
          'May',
          'Jun',
          'Jul',
          'Aug',
          'Sep',
          'Oct',
          'Nov',
          'Dec',
        ];
        final day = parsed.day.toString().padLeft(2, '0');
        final month = months[parsed.month - 1];
        final year = parsed.year;
        return '$day $month $year';
      }
    } catch (_) {}
    return dateOfBirth;
  }

  /// Calculates age in years from DOB if valid
  String? get calculatedAge {
    if (dateOfBirth.isEmpty) return null;
    try {
      DateTime? parsed;
      if (dateOfBirth.contains('-')) {
        final parts = dateOfBirth.split('-');
        if (parts.length == 3) {
          if (parts[0].length == 4) {
            parsed = DateTime(
              int.parse(parts[0]),
              int.parse(parts[1]),
              int.parse(parts[2]),
            );
          } else {
            parsed = DateTime(
              int.parse(parts[2]),
              int.parse(parts[1]),
              int.parse(parts[0]),
            );
          }
        }
      }
      if (parsed != null) {
        final now = DateTime.now();
        var age = now.year - parsed.year;
        if (now.month < parsed.month ||
            (now.month == parsed.month && now.day < parsed.day)) {
          age--;
        }
        if (age > 0 && age < 120) {
          return '($age Years)';
        }
      }
    } catch (_) {}
    return null;
  }

  factory MemberProfile.fromJson(Map<String, dynamic> json) {
    final isPaidVal =
        json['isPaid'] == true ||
        json['subscription_active'] == true ||
        (json['payment_status'] as String? ?? '').toLowerCase() == 'paid';

    return MemberProfile(
      id: (json['id'] as String? ?? '').trim(),
      tagaId:
          (json['tagaId'] ?? json['taga_id'] ?? json['sno']?.toString() ?? '')
              .toString()
              .trim(),
      username: (json['username'] as String? ?? '').trim(),
      name: (json['name'] as String? ?? '').trim(),
      initial: (json['initial'] as String? ?? '').trim(),
      gender: (json['gender'] as String? ?? '').trim(),
      dateOfBirth: (json['date_of_birth'] ?? json['dateOfBirth'] ?? '')
          .toString()
          .trim(),
      fatherName:
          (json['father_name'] ?? json['fatherName'] ?? json['father'] ?? '')
              .toString()
              .trim(),
      motherName:
          (json['mother_name'] ?? json['motherName'] ?? json['mother'] ?? '')
              .toString()
              .trim(),
      mobileNumber: (json['mobile_number'] ?? json['mobileNumber'] ?? '')
          .toString()
          .trim(),
      emailId: (json['emailId'] ?? json['email'] ?? '').toString().trim(),
      designation: (json['designation'] as String? ?? '').trim(),
      recruitmentBatch:
          (json['recruitment_batch'] ?? json['recruitmentBatch'] ?? '')
              .toString()
              .trim(),
      seniorityNumber:
          (json['seniority_number'] ?? json['seniorityNumber'] ?? '')
              .toString()
              .trim(),
      workingDistrict:
          (json['working_district'] ?? json['workingDistrict'] ?? '')
              .toString()
              .trim(),
      nativeDistrict: (json['native_district'] ?? json['nativeDistrict'] ?? '')
          .toString()
          .trim(),
      residentialAddress:
          (json['residential_address'] ?? json['residentialAddress'] ?? '')
              .toString()
              .trim(),
      permanentAddress:
          (json['permanent_address'] ?? json['permanentAddress'] ?? '')
              .toString()
              .trim(),
      educationalQualification:
          (json['educational_qualification'] ??
                  json['educationalQualification'] ??
                  '')
              .toString()
              .trim(),
      tbfNumber: (json['tbf_number'] ?? json['tbfNumber'] ?? '')
          .toString()
          .trim(),
      cpsGpfNumber: (json['cps_gpf_number'] ?? json['cpsGpfNumber'] ?? '')
          .toString()
          .trim(),
      paymentStatus:
          (json['payment_status'] as String? ?? (isPaidVal ? 'Paid' : 'Unpaid'))
              .trim(),
      isPaid: isPaidVal,
      subscriptionActive: isPaidVal,
    );
  }

  MemberProfile copyWith({
    String? id,
    String? tagaId,
    String? username,
    String? name,
    String? initial,
    String? gender,
    String? dateOfBirth,
    String? fatherName,
    String? motherName,
    String? mobileNumber,
    String? emailId,
    String? designation,
    String? recruitmentBatch,
    String? seniorityNumber,
    String? workingDistrict,
    String? nativeDistrict,
    String? residentialAddress,
    String? permanentAddress,
    String? educationalQualification,
    String? tbfNumber,
    String? cpsGpfNumber,
    String? paymentStatus,
    bool? isPaid,
    bool? subscriptionActive,
  }) {
    return MemberProfile(
      id: id ?? this.id,
      tagaId: tagaId ?? this.tagaId,
      username: username ?? this.username,
      name: name ?? this.name,
      initial: initial ?? this.initial,
      gender: gender ?? this.gender,
      dateOfBirth: dateOfBirth ?? this.dateOfBirth,
      fatherName: fatherName ?? this.fatherName,
      motherName: motherName ?? this.motherName,
      mobileNumber: mobileNumber ?? this.mobileNumber,
      emailId: emailId ?? this.emailId,
      designation: designation ?? this.designation,
      recruitmentBatch: recruitmentBatch ?? this.recruitmentBatch,
      seniorityNumber: seniorityNumber ?? this.seniorityNumber,
      workingDistrict: workingDistrict ?? this.workingDistrict,
      nativeDistrict: nativeDistrict ?? this.nativeDistrict,
      residentialAddress: residentialAddress ?? this.residentialAddress,
      permanentAddress: permanentAddress ?? this.permanentAddress,
      educationalQualification:
          educationalQualification ?? this.educationalQualification,
      tbfNumber: tbfNumber ?? this.tbfNumber,
      cpsGpfNumber: cpsGpfNumber ?? this.cpsGpfNumber,
      paymentStatus: paymentStatus ?? this.paymentStatus,
      isPaid: isPaid ?? this.isPaid,
      subscriptionActive: subscriptionActive ?? this.subscriptionActive,
    );
  }
}

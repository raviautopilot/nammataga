library;

/// Payload model for submitting profile changes to `POST /api/member/edit-request`
class EditRequestPayload {
  final String email;
  final String mobileNumber;
  final String mailId;
  final String designation;
  final String workingDistrict;
  final String residentialAddress;
  final String permanentAddress;
  final String remarks;

  const EditRequestPayload({
    required this.email,
    required this.mobileNumber,
    required this.mailId,
    required this.designation,
    required this.workingDistrict,
    required this.residentialAddress,
    required this.permanentAddress,
    required this.remarks,
  });

  Map<String, dynamic> toJson() {
    return {
      'email': email.trim(),
      'mobileNumber': mobileNumber.trim(),
      'mailId': mailId.trim(),
      'designation': designation.trim(),
      'workingDistrict': workingDistrict.trim(),
      'residentialAddress': residentialAddress.trim(),
      'permanentAddress': permanentAddress.trim(),
      'remarks': remarks.trim(),
    };
  }
}

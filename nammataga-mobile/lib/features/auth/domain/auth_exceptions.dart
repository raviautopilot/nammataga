/// Thrown when a user with a pending deletion request attempts to sign in.
class DeletionPendingException implements Exception {
  final String message;

  const DeletionPendingException([
    this.message =
        'Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending.',
  ]);

  @override
  String toString() => message;
}

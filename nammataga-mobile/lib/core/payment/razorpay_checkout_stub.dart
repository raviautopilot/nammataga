void openWebRazorpay({
  required Map<String, dynamic> options,
  required void Function(String paymentId, String orderId, String signature)
  onSuccess,
  required void Function(String error) onError,
  required void Function() onDismiss,
}) {
  throw UnsupportedError('openWebRazorpay is only supported on Web platform');
}

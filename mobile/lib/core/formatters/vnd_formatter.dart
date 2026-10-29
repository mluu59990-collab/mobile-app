String formatVnd(int amount) {
  final digits = amount.abs().toString();
  final grouped = digits.replaceAllMapped(
    RegExp(r'(?=(\d{3})+(?!\d))'),
    (_) => '.',
  );
  return '${amount < 0 ? '-' : ''}$grouped ₫';
}

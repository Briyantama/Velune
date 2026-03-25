class TransactionSummary {
  final String from;
  final String to;
  final String currency;
  final int incomeMinor;
  final int expenseMinor;
  final int netMinor;

  const TransactionSummary({
    required this.from,
    required this.to,
    required this.currency,
    required this.incomeMinor,
    required this.expenseMinor,
    required this.netMinor,
  });

  factory TransactionSummary.fromJson(Map<String, dynamic> json) {
    return TransactionSummary(
      from: json['from'] as String,
      to: json['to'] as String,
      currency: json['currency'] as String,
      incomeMinor: (json['incomeMinor'] as num).toInt(),
      expenseMinor: (json['expenseMinor'] as num).toInt(),
      netMinor: (json['netMinor'] as num).toInt(),
    );
  }
}


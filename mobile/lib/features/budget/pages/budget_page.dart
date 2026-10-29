import 'package:flutter/material.dart';

class BudgetPage extends StatelessWidget {
  const BudgetPage({super.key});
  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Ngân sách')),
    body: const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.account_balance_wallet_outlined, size: 48),
          SizedBox(height: 12),
          Text(
            'Sắp có',
            style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700),
          ),
          SizedBox(height: 6),
          Text('Tính năng ngân sách sẽ được bổ sung sau.'),
        ],
      ),
    ),
  );
}

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/formatters/local_datetime_formatter.dart';
import '../../../core/formatters/vnd_formatter.dart';
import '../../../core/providers/app_providers.dart';
import '../../expense/data/expense.dart';
import '../../expense/data/monthly_report.dart';
import '../../expense/providers/expense_providers.dart';

const _pink = Color(0xFFEE6F8E);
const _ink = Color(0xFF25222A);

class HomePage extends ConsumerStatefulWidget {
  const HomePage({super.key});
  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  int _tab = 0;

  Future<void> _logout() async {
    try {
      await ref.read(authRepositoryProvider).logout();
    } catch (_) {}
    if (mounted) context.go('/login');
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: const Color(0xFFFFF9FA),
    body: SafeArea(
      child: IndexedStack(
        index: _tab,
        children: [
          const _DashboardTab(),
          const _HistoryTab(),
          const _ReportTab(),
          _ProfileTab(
            onSettings: () => context.push('/settings'),
            onLogout: _logout,
          ),
        ],
      ),
    ),
    floatingActionButton: FloatingActionButton(
      onPressed: () => context.push('/expense/new'),
      backgroundColor: _pink,
      foregroundColor: Colors.white,
      child: const Icon(Icons.add),
    ),
    floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
    bottomNavigationBar: BottomAppBar(
      color: Colors.white,
      shape: const CircularNotchedRectangle(),
      notchMargin: 7,
      child: SizedBox(
        height: 62,
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _NavItem(
              icon: Icons.home_outlined,
              label: 'Trang chủ',
              active: _tab == 0,
              onTap: () => setState(() => _tab = 0),
            ),
            _NavItem(
              icon: Icons.history,
              label: 'Lịch sử',
              active: _tab == 1,
              onTap: () => setState(() => _tab = 1),
            ),
            const SizedBox(width: 52),
            _NavItem(
              icon: Icons.bar_chart_rounded,
              label: 'Thống kê',
              active: _tab == 2,
              onTap: () => setState(() => _tab = 2),
            ),
            _NavItem(
              icon: Icons.person_outline,
              label: 'Cá nhân',
              active: _tab == 3,
              onTap: () => setState(() => _tab = 3),
            ),
          ],
        ),
      ),
    ),
  );
}

class _NavItem extends StatelessWidget {
  const _NavItem({
    required this.icon,
    required this.label,
    required this.active,
    required this.onTap,
  });
  final IconData icon;
  final String label;
  final bool active;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Expanded(
    child: InkWell(
      onTap: onTap,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: active ? _pink : Colors.grey, size: 21),
          const SizedBox(height: 3),
          Text(
            label,
            style: TextStyle(fontSize: 10, color: active ? _pink : Colors.grey),
          ),
        ],
      ),
    ),
  );
}

class _DashboardTab extends ConsumerWidget {
  const _DashboardTab();
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final expenses = ref.watch(expensesForSelectedMonthProvider);
    final report = ref.watch(monthlyReportProvider);
    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(expensesForSelectedMonthProvider);
        ref.invalidate(monthlyReportProvider);
      },
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(18),
        children: [
          const Text(
            'Tổng quan',
            style: TextStyle(
              fontSize: 23,
              fontWeight: FontWeight.w700,
              color: _ink,
            ),
          ),
          const SizedBox(height: 16),
          report.when(
            loading: () => const _Panel(child: LinearProgressIndicator()),
            error: (e, _) => _RetryPanel(
              error: e,
              onRetry: () {
                ref.invalidate(monthlyReportProvider);
                ref.invalidate(expensesForSelectedMonthProvider);
              },
            ),
            data: (data) => _SummaryCard(
              report: data,
              expenses: expenses.value ?? const [],
            ),
          ),
          const SizedBox(height: 20),
          Row(
            children: [
              const Expanded(
                child: Text(
                  'Giao dịch gần đây',
                  style: TextStyle(fontSize: 17, fontWeight: FontWeight.w700),
                ),
              ),
              TextButton(onPressed: () {}, child: const Text('')),
            ],
          ),
          expenses.when(
            loading: () => const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              ),
            ),
            error: (e, _) => _RetryPanel(
              error: e,
              onRetry: () => ref.invalidate(expensesForSelectedMonthProvider),
            ),
            data: (items) => items.isEmpty
                ? const _EmptyState(message: 'Tháng này chưa có chi tiêu.')
                : Column(
                    children: items
                        .take(5)
                        .map((item) => _ExpenseTile(expense: item))
                        .toList(),
                  ),
          ),
          const SizedBox(height: 18),
          _Panel(
            child: ListTile(
              leading: const Icon(
                Icons.account_balance_wallet_outlined,
                color: _pink,
              ),
              title: const Text('Ngân sách'),
              subtitle: const Text('Sắp có'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/budget'),
            ),
          ),
        ],
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard({required this.report, required this.expenses});
  final MonthlyReport report;
  final List<Expense> expenses;
  @override
  Widget build(BuildContext context) {
    final today = DateTime.now();
    final todayTotal = expenses
        .where((e) {
          final d = e.spentAt.toLocal();
          return d.year == today.year &&
              d.month == today.month &&
              d.day == today.day;
        })
        .fold<int>(0, (sum, e) => sum + e.amount);
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFFF2829E), Color(0xFFE96889)],
        ),
        borderRadius: BorderRadius.circular(22),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Tổng chi tiêu tháng này',
            style: TextStyle(color: Colors.white),
          ),
          const SizedBox(height: 7),
          Text(
            formatVnd(report.totalAmount),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 27,
              fontWeight: FontWeight.w700,
            ),
          ),
          const Divider(color: Colors.white54, height: 26),
          Row(
            children: [
              Expanded(
                child: _Metric(label: 'Hôm nay', value: formatVnd(todayTotal)),
              ),
              const Expanded(
                child: _Metric(label: 'Ngân sách', value: 'Sắp có'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _Metric extends StatelessWidget {
  const _Metric({required this.label, required this.value});
  final String label, value;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(label, style: const TextStyle(color: Colors.white70, fontSize: 11)),
      const SizedBox(height: 4),
      Text(
        value,
        style: const TextStyle(
          color: Colors.white,
          fontWeight: FontWeight.w700,
        ),
      ),
    ],
  );
}

class _HistoryTab extends ConsumerStatefulWidget {
  const _HistoryTab();
  @override
  ConsumerState<_HistoryTab> createState() => _HistoryTabState();
}

class _HistoryTabState extends ConsumerState<_HistoryTab> {
  String _filter = 'all';
  String _query = '';
  @override
  Widget build(BuildContext context) {
    final month = ref.watch(selectedExpenseMonthProvider);
    final async = ref.watch(expensesForSelectedMonthProvider);
    final filters = const [
      ('all', 'Tất cả'),
      ('food', 'Ăn uống'),
      ('transport', 'Di chuyển'),
      ('shopping', 'Mua sắm'),
    ];
    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(expensesForSelectedMonthProvider),
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(18),
        children: [
          const Text(
            'Lịch sử chi tiêu',
            style: TextStyle(
              fontSize: 23,
              fontWeight: FontWeight.w700,
              color: _ink,
            ),
          ),
          _MonthPicker(
            month: month,
            onShift: (d) =>
                ref.read(selectedExpenseMonthProvider.notifier).shift(d),
          ),
          TextField(
            decoration: const InputDecoration(
              prefixIcon: Icon(Icons.search),
              hintText: 'Tìm kiếm giao dịch...',
            ),
            onChanged: (value) => setState(() => _query = value.toLowerCase()),
          ),
          const SizedBox(height: 10),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: filters
                  .map(
                    (item) => Padding(
                      padding: const EdgeInsets.only(right: 6),
                      child: ChoiceChip(
                        label: Text(item.$2),
                        selected: _filter == item.$1,
                        onSelected: (_) => setState(() => _filter = item.$1),
                      ),
                    ),
                  )
                  .toList(),
            ),
          ),
          const SizedBox(height: 12),
          async.when(
            loading: () => const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              ),
            ),
            error: (e, _) => _RetryPanel(
              error: e,
              onRetry: () => ref.invalidate(expensesForSelectedMonthProvider),
            ),
            data: (items) {
              final shown = items
                  .where(
                    (e) =>
                        (_filter == 'all' || e.category == _filter) &&
                        ('${e.note ?? ''} ${categoryFor(e.category).label}'
                            .toLowerCase()
                            .contains(_query)),
                  )
                  .toList();
              return shown.isEmpty
                  ? const _EmptyState(
                      message: 'Không có giao dịch phù hợp trong tháng này.',
                    )
                  : Column(
                      children: shown
                          .map((item) => _ExpenseTile(expense: item))
                          .toList(),
                    );
            },
          ),
        ],
      ),
    );
  }
}

class _ReportTab extends ConsumerWidget {
  const _ReportTab();
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final month = ref.watch(selectedExpenseMonthProvider);
    final report = ref.watch(monthlyReportProvider);
    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(monthlyReportProvider),
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(18),
        children: [
          const Text(
            'Thống kê',
            style: TextStyle(
              fontSize: 23,
              fontWeight: FontWeight.w700,
              color: _ink,
            ),
          ),
          _MonthPicker(
            month: month,
            onShift: (d) =>
                ref.read(selectedExpenseMonthProvider.notifier).shift(d),
          ),
          report.when(
            loading: () => const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              ),
            ),
            error: (e, _) => _RetryPanel(
              error: e,
              onRetry: () => ref.invalidate(monthlyReportProvider),
            ),
            data: (data) => data.totalAmount == 0
                ? const _EmptyState(
                    message: 'Tháng này chưa có dữ liệu thống kê.',
                  )
                : Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _Panel(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('Tổng chi tiêu'),
                            const SizedBox(height: 5),
                            Text(
                              formatVnd(data.totalAmount),
                              style: const TextStyle(
                                fontSize: 23,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: 14),
                            _DailyBars(days: data.byDay),
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),
                      const Text(
                        'Tổng theo danh mục',
                        style: TextStyle(
                          fontSize: 17,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 10),
                      _Panel(
                        child: Column(
                          children: data.byCategory.map((item) {
                            final cat = categoryFor(item.category);
                            return ListTile(
                              dense: true,
                              contentPadding: EdgeInsets.zero,
                              leading: Icon(cat.icon, color: _pink),
                              title: Text(cat.label),
                              trailing: Text(formatVnd(item.totalAmount)),
                            );
                          }).toList(),
                        ),
                      ),
                    ],
                  ),
          ),
        ],
      ),
    );
  }
}

class _DailyBars extends StatelessWidget {
  const _DailyBars({required this.days});
  final List<DayTotal> days;
  @override
  Widget build(BuildContext context) {
    final max = days.fold<int>(
      0,
      (v, d) => d.totalAmount > v ? d.totalAmount : v,
    );
    return SizedBox(
      height: 125,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: days.map((day) {
          final label = day.date;
          final barHeight = max == 0 ? 2.0 : 8 + 86 * day.totalAmount / max;
          return Expanded(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 2),
              child: Tooltip(
                message: '$label: ${formatVnd(day.totalAmount)}',
                child: Container(
                  height: barHeight,
                  decoration: const BoxDecoration(
                    color: _pink,
                    borderRadius: BorderRadius.vertical(
                      top: Radius.circular(4),
                    ),
                  ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}

class _MonthPicker extends StatelessWidget {
  const _MonthPicker({required this.month, required this.onShift});
  final String month;
  final ValueChanged<int> onShift;
  @override
  Widget build(BuildContext context) => Row(
    children: [
      IconButton(
        onPressed: () => onShift(-1),
        icon: const Icon(Icons.chevron_left),
      ),
      Expanded(
        child: Center(
          child: Text(
            month,
            style: const TextStyle(fontWeight: FontWeight.w700),
          ),
        ),
      ),
      IconButton(
        onPressed: () => onShift(1),
        icon: const Icon(Icons.chevron_right),
      ),
    ],
  );
}

class _ExpenseTile extends ConsumerWidget {
  const _ExpenseTile({required this.expense});
  final Expense expense;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final category = categoryFor(expense.category);
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: CircleAvatar(
        backgroundColor: const Color(0xFFFFEEF2),
        child: Icon(category.icon, color: _pink),
      ),
      title: Text(
        (expense.note?.isNotEmpty ?? false) ? expense.note! : category.label,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(
        '${category.label} · ${formatLocalDateTime(expense.spentAt)}',
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            formatVnd(expense.amount),
            style: const TextStyle(fontWeight: FontWeight.w700),
          ),
          PopupMenuButton<String>(
            onSelected: (value) async {
              if (value == 'edit') {
                await context.push('/expense/${expense.id}/edit');
                return;
              }
              if (value == 'delete') {
                final confirmed = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('Xóa chi tiêu?'),
                    content: const Text('Bạn có chắc muốn xóa giao dịch này?'),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(context, false),
                        child: const Text('Hủy'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.pop(context, true),
                        child: const Text('Xóa'),
                      ),
                    ],
                  ),
                );
                if (confirmed == true && context.mounted) {
                  try {
                    await ref.read(expenseActionsProvider).delete(expense.id);
                  } catch (e) {
                    if (context.mounted) {
                      ScaffoldMessenger.of(context)
                          .showSnackBar(SnackBar(content: Text(e.toString())));
                    }
                  }
                }
              }
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: 'edit', child: Text('Sửa')),
              PopupMenuItem(value: 'delete', child: Text('Xóa')),
            ],
          ),
        ],
      ),
    );
  }
}

class _ProfileTab extends StatelessWidget {
  const _ProfileTab({required this.onSettings, required this.onLogout});
  final VoidCallback onSettings;
  final Future<void> Function() onLogout;
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.all(18),
    children: [
      const Text(
        'Cá nhân',
        style: TextStyle(fontSize: 23, fontWeight: FontWeight.w700),
      ),
      const SizedBox(height: 20),
      _Panel(
        child: Column(
          children: [
            ListTile(
              leading: const Icon(Icons.account_balance_wallet_outlined),
              title: const Text('Ngân sách của tôi'),
              subtitle: const Text('Sắp có'),
              onTap: () => context.push('/budget'),
            ),
            const ListTile(
              leading: Icon(Icons.category_outlined),
              title: Text('Danh mục chi tiêu'),
              subtitle: Text('Sắp có'),
            ),
            const ListTile(
              leading: Icon(Icons.credit_card_outlined),
              title: Text('Phương thức thanh toán'),
              subtitle: Text('Sắp có'),
            ),
            const ListTile(
              leading: Icon(Icons.file_upload_outlined),
              title: Text('Xuất dữ liệu'),
              subtitle: Text('Sắp có'),
            ),
            ListTile(
              leading: const Icon(Icons.settings_outlined),
              title: const Text('Cài đặt'),
              onTap: onSettings,
            ),
            ListTile(
              leading: const Icon(Icons.logout),
              title: const Text('Đăng xuất'),
              onTap: () => onLogout(),
            ),
          ],
        ),
      ),
    ],
  );
}

class _RetryPanel extends StatelessWidget {
  const _RetryPanel({required this.error, required this.onRetry});
  final Object error;
  final VoidCallback onRetry;
  @override
  Widget build(BuildContext context) => _Panel(
    child: Column(
      children: [
        Text(error.toString(), textAlign: TextAlign.center),
        TextButton(onPressed: onRetry, child: const Text('Thử lại')),
      ],
    ),
  );
}

class _EmptyState extends StatelessWidget {
  const _EmptyState({required this.message});
  final String message;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(30),
    child: Center(
      child: Text(
        message,
        textAlign: TextAlign.center,
        style: const TextStyle(color: Colors.black54),
      ),
    ),
  );
}

class _Panel extends StatelessWidget {
  const _Panel({required this.child});
  final Widget child;
  @override
  Widget build(BuildContext context) => Container(
    decoration: BoxDecoration(borderRadius: BorderRadius.circular(17)),
    child: Material(
      color: Colors.white,
      borderRadius: BorderRadius.circular(17),
      child: Padding(padding: const EdgeInsets.all(15), child: child),
    ),
  );
}

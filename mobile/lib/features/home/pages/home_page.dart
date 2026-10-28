import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

const _pink = Color(0xFFEE6F8E);
const _ink = Color(0xFF25222A);

class HomePage extends ConsumerStatefulWidget {
  const HomePage({super.key});

  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final pages = [
      const _DashboardTab(),
      const _HistoryTab(),
      const _ReportTab(),
      _ProfileTab(onSettings: () => context.push('/settings')),
    ];
    return Scaffold(
      backgroundColor: const Color(0xFFFFF9FA),
      body: SafeArea(
        child: IndexedStack(index: _tab, children: pages),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/expense/new'),
        backgroundColor: _pink,
        foregroundColor: Colors.white,
        elevation: 4,
        shape: const CircleBorder(),
        child: const Icon(Icons.camera_alt_outlined, size: 25),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: BottomAppBar(
        color: Colors.white,
        surfaceTintColor: Colors.white,
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
          Icon(icon, color: active ? _pink : const Color(0xFF77727A), size: 21),
          const SizedBox(height: 3),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              color: active ? _pink : const Color(0xFF77727A),
              fontWeight: active ? FontWeight.w600 : FontWeight.normal,
            ),
          ),
        ],
      ),
    ),
  );
}

class _DashboardTab extends StatelessWidget {
  const _DashboardTab();

  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.fromLTRB(18, 18, 18, 26),
    children: [
      Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Xin chào,',
                  style: TextStyle(color: Colors.grey.shade700, fontSize: 13),
                ),
                const SizedBox(height: 3),
                const Text(
                  'Nguyễn Thảo 👋',
                  style: TextStyle(
                    fontSize: 21,
                    fontWeight: FontWeight.w700,
                    color: _ink,
                  ),
                ),
              ],
            ),
          ),
          _CircleAction(icon: Icons.notifications_none_rounded, onTap: () {}),
        ],
      ),
      const SizedBox(height: 19),
      Container(
        padding: const EdgeInsets.all(19),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [Color(0xFFF2829E), Color(0xFFE96889)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(22),
          boxShadow: const [
            BoxShadow(
              color: Color(0x22E96889),
              blurRadius: 18,
              offset: Offset(0, 8),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Expanded(
                  child: Text(
                    'Tổng chi tiêu tháng này',
                    style: TextStyle(color: Colors.white, fontSize: 13),
                  ),
                ),
                _Tag(label: '↑ 12%'),
              ],
            ),
            const SizedBox(height: 5),
            const Text(
              '3.450.000 đ',
              style: TextStyle(
                color: Colors.white,
                fontSize: 27,
                fontWeight: FontWeight.w700,
                letterSpacing: .3,
              ),
            ),
            const SizedBox(height: 19),
            const Divider(color: Color(0x55FFFFFF), height: 1),
            const SizedBox(height: 12),
            const Row(
              children: [
                Expanded(
                  child: _CardMetric(label: 'Hôm nay', value: '120.000 đ'),
                ),
                Expanded(
                  child: _CardMetric(
                    label: 'Ngân sách còn',
                    value: '1.550.000 đ',
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
      const SizedBox(height: 20),
      Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: const [
          _QuickAction(
            icon: Icons.camera_alt_outlined,
            label: 'Chụp ảnh',
            color: Color(0xFFF58A9E),
          ),
          _QuickAction(
            icon: Icons.add,
            label: 'Thủ công',
            color: Color(0xFF9575CD),
          ),
          _QuickAction(
            icon: Icons.receipt_long_outlined,
            label: 'Hóa đơn',
            color: Color(0xFF638CE6),
          ),
          _QuickAction(
            icon: Icons.track_changes_outlined,
            label: 'Ngân sách',
            color: Color(0xFFFFA64D),
          ),
        ],
      ),
      const SizedBox(height: 24),
      _SectionTitle(
        title: 'Giao dịch gần đây',
        action: 'Xem tất cả',
        onTap: () {},
      ),
      const SizedBox(height: 8),
      const _TransactionTile(
        icon: Icons.local_cafe_rounded,
        color: Color(0xFFE67865),
        title: 'Cà phê Highlands',
        subtitle: 'Hôm nay, 08:32',
        amount: '-45.000 đ',
      ),
      const _TransactionTile(
        icon: Icons.restaurant_rounded,
        color: Color(0xFFED7890),
        title: 'Ăn trưa',
        subtitle: 'Hôm nay, 12:15',
        amount: '-75.000 đ',
      ),
      const _TransactionTile(
        icon: Icons.shopping_bag_outlined,
        color: Color(0xFFFFA64D),
        title: 'Mua sắm Shopee',
        subtitle: 'Hôm qua, 20:10',
        amount: '-250.000 đ',
      ),
    ],
  );
}

class _HistoryTab extends StatefulWidget {
  const _HistoryTab();
  @override
  State<_HistoryTab> createState() => _HistoryTabState();
}

class _HistoryTabState extends State<_HistoryTab> {
  String _filter = 'Tất cả';
  final _filters = const ['Tất cả', 'Ăn uống', 'Di chuyển', 'Mua sắm'];
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.fromLTRB(18, 20, 18, 24),
    children: [
      Row(
        children: [
          const Expanded(
            child: Text(
              'Lịch sử chi tiêu',
              style: TextStyle(
                fontSize: 23,
                fontWeight: FontWeight.w700,
                color: _ink,
              ),
            ),
          ),
          _CircleAction(icon: Icons.calendar_month_outlined, onTap: () {}),
          const SizedBox(width: 8),
          _CircleAction(icon: Icons.tune_rounded, onTap: () {}),
        ],
      ),
      const SizedBox(height: 16),
      TextField(
        decoration: InputDecoration(
          hintText: 'Tìm kiếm giao dịch...',
          prefixIcon: const Icon(Icons.search),
          filled: true,
          fillColor: Colors.white,
          contentPadding: EdgeInsets.zero,
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(15),
            borderSide: BorderSide.none,
          ),
        ),
      ),
      const SizedBox(height: 13),
      SizedBox(
        height: 37,
        child: ListView(
          scrollDirection: Axis.horizontal,
          children: _filters
              .map(
                (item) => Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: ChoiceChip(
                    label: Text(item),
                    selected: _filter == item,
                    onSelected: (_) => setState(() => _filter = item),
                    selectedColor: _pink,
                    backgroundColor: Colors.white,
                    labelStyle: TextStyle(
                      color: _filter == item ? Colors.white : _ink,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                    side: BorderSide.none,
                  ),
                ),
              )
              .toList(),
        ),
      ),
      const SizedBox(height: 20),
      const _DateHeader(title: 'Hôm nay - 12/10/2024', total: '-120.000 đ'),
      const _TransactionTile(
        icon: Icons.local_cafe_rounded,
        color: Color(0xFFE67865),
        title: 'Cà phê Highlands',
        subtitle: '08:32',
        amount: '-45.000 đ',
      ),
      const _TransactionTile(
        icon: Icons.restaurant_rounded,
        color: Color(0xFFED7890),
        title: 'Ăn trưa',
        subtitle: '12:15',
        amount: '-75.000 đ',
      ),
      const SizedBox(height: 15),
      const _DateHeader(title: 'Hôm qua - 11/10/2024', total: '-305.000 đ'),
      const _TransactionTile(
        icon: Icons.shopping_bag_outlined,
        color: Color(0xFFFFA64D),
        title: 'Mua sắm Shopee',
        subtitle: '20:10',
        amount: '-250.000 đ',
      ),
      const _TransactionTile(
        icon: Icons.directions_car_filled_outlined,
        color: Color(0xFF55AA83),
        title: 'Grab',
        subtitle: '18:20',
        amount: '-55.000 đ',
      ),
      const SizedBox(height: 15),
      const _DateHeader(title: '10/10/2024', total: '-180.000 đ'),
      const _TransactionTile(
        icon: Icons.restaurant,
        color: Color(0xFF9292D5),
        title: 'Ăn tối',
        subtitle: '19:05',
        amount: '-80.000 đ',
      ),
      const _TransactionTile(
        icon: Icons.shopping_cart_outlined,
        color: Color(0xFF5D9AD8),
        title: 'Siêu thị',
        subtitle: '15:30',
        amount: '-100.000 đ',
      ),
    ],
  );
}

class _ReportTab extends StatelessWidget {
  const _ReportTab();
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.fromLTRB(18, 20, 18, 24),
    children: [
      const Text(
        'Thống kê',
        style: TextStyle(
          fontSize: 23,
          fontWeight: FontWeight.w700,
          color: _ink,
        ),
      ),
      const SizedBox(height: 16),
      Row(
        children: const [
          _PeriodChip(label: 'Tháng', active: true),
          _PeriodChip(label: 'Quý'),
          _PeriodChip(label: 'Năm'),
          _PeriodChip(label: 'Tùy chỉnh'),
        ],
      ),
      const SizedBox(height: 18),
      const Row(
        children: [
          Icon(Icons.chevron_left),
          Expanded(
            child: Center(
              child: Text(
                'Tháng 10, 2024',
                style: TextStyle(fontWeight: FontWeight.w600),
              ),
            ),
          ),
          Icon(Icons.chevron_right),
        ],
      ),
      const SizedBox(height: 14),
      _Panel(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Expanded(
                  child: Text(
                    'Tổng chi tiêu',
                    style: TextStyle(color: Color(0xFF706B73)),
                  ),
                ),
                Text(
                  '↑ 12%',
                  style: TextStyle(color: _pink, fontWeight: FontWeight.w600),
                ),
              ],
            ),
            const SizedBox(height: 3),
            const Text(
              '3.450.000 đ',
              style: TextStyle(fontSize: 22, fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 12),
            const _BarChart(),
            const SizedBox(height: 6),
            const Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('1'),
                Text('5'),
                Text('10'),
                Text('15'),
                Text('20'),
                Text('25'),
                Text('31'),
              ],
            ),
          ],
        ),
      ),
      const SizedBox(height: 20),
      const Text(
        'Phân bổ chi tiêu',
        style: TextStyle(fontSize: 17, fontWeight: FontWeight.w700),
      ),
      const SizedBox(height: 10),
      _Panel(
        child: Column(
          children: const [
            _CategoryRow(
              icon: Icons.restaurant,
              label: 'Ăn uống',
              amount: '1.380.000 đ',
              percent: '40%',
              color: Color(0xFFF27491),
            ),
            _CategoryRow(
              icon: Icons.shopping_bag,
              label: 'Mua sắm',
              amount: '862.500 đ',
              percent: '25%',
              color: Color(0xFFEF9BAE),
            ),
            _CategoryRow(
              icon: Icons.directions_car,
              label: 'Di chuyển',
              amount: '517.500 đ',
              percent: '15%',
              color: Color(0xFF669BE2),
            ),
            _CategoryRow(
              icon: Icons.receipt,
              label: 'Hóa đơn',
              amount: '345.000 đ',
              percent: '10%',
              color: Color(0xFFF5A262),
            ),
            _CategoryRow(
              icon: Icons.movie,
              label: 'Giải trí',
              amount: '172.500 đ',
              percent: '5%',
              color: Color(0xFF8F80D4),
            ),
            _CategoryRow(
              icon: Icons.more_horiz,
              label: 'Khác',
              amount: '172.500 đ',
              percent: '5%',
              color: Color(0xFFB8B4BA),
            ),
          ],
        ),
      ),
    ],
  );
}

class _ProfileTab extends StatelessWidget {
  const _ProfileTab({required this.onSettings});
  final VoidCallback onSettings;
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.fromLTRB(18, 24, 18, 24),
    children: [
      const Text(
        'Cá nhân',
        style: TextStyle(
          fontSize: 23,
          fontWeight: FontWeight.w700,
          color: _ink,
        ),
      ),
      const SizedBox(height: 20),
      Row(
        children: [
          CircleAvatar(
            radius: 29,
            backgroundColor: const Color(0xFFFAD9E0),
            child: Icon(Icons.person, color: _pink, size: 33),
          ),
          const SizedBox(width: 13),
          const Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Nguyễn Thảo',
                style: TextStyle(fontWeight: FontWeight.w700, fontSize: 17),
              ),
              SizedBox(height: 4),
              Text(
                'thaonguyen@gmail.com',
                style: TextStyle(color: Colors.black54, fontSize: 12),
              ),
            ],
          ),
          const Spacer(),
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.edit_outlined, color: _pink),
          ),
        ],
      ),
      const SizedBox(height: 18),
      _Panel(
        child: Column(
          children: [
            _ProfileItem(
              icon: Icons.account_balance_wallet_outlined,
              label: 'Ngân sách của tôi',
              onTap: () {},
            ),
            _ProfileItem(
              icon: Icons.category_outlined,
              label: 'Danh mục chi tiêu',
              onTap: () {},
            ),
            _ProfileItem(
              icon: Icons.credit_card_outlined,
              label: 'Phương thức thanh toán',
              onTap: () {},
            ),
            _ProfileItem(
              icon: Icons.file_upload_outlined,
              label: 'Xuất dữ liệu',
              onTap: () {},
            ),
            _ProfileItem(
              icon: Icons.settings_outlined,
              label: 'Cài đặt',
              onTap: onSettings,
            ),
            _ProfileItem(
              icon: Icons.help_outline,
              label: 'Trợ giúp & Phản hồi',
              onTap: () {},
            ),
            _ProfileItem(
              icon: Icons.info_outline,
              label: 'Về ứng dụng',
              onTap: () {},
            ),
          ],
        ),
      ),
    ],
  );
}

class _CircleAction extends StatelessWidget {
  const _CircleAction({required this.icon, required this.onTap});
  final IconData icon;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Material(
    color: const Color(0xFFFFEEF2),
    shape: const CircleBorder(),
    child: InkWell(
      customBorder: const CircleBorder(),
      onTap: onTap,
      child: SizedBox(
        width: 42,
        height: 42,
        child: Icon(icon, color: _ink, size: 21),
      ),
    ),
  );
}

class _Tag extends StatelessWidget {
  const _Tag({required this.label});
  final String label;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
    decoration: BoxDecoration(
      color: const Color(0x44FFFFFF),
      borderRadius: BorderRadius.circular(20),
    ),
    child: Text(
      label,
      style: const TextStyle(
        color: Colors.white,
        fontSize: 11,
        fontWeight: FontWeight.w600,
      ),
    ),
  );
}

class _CardMetric extends StatelessWidget {
  const _CardMetric({required this.label, required this.value});
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(
        label,
        style: const TextStyle(color: Color(0xEEFFFFFF), fontSize: 11),
      ),
      const SizedBox(height: 4),
      Text(
        value,
        style: const TextStyle(
          color: Colors.white,
          fontSize: 13,
          fontWeight: FontWeight.w700,
        ),
      ),
    ],
  );
}

class _QuickAction extends StatelessWidget {
  const _QuickAction({
    required this.icon,
    required this.label,
    required this.color,
  });
  final IconData icon;
  final String label;
  final Color color;
  @override
  Widget build(BuildContext context) => Column(
    children: [
      Container(
        width: 54,
        height: 54,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(17),
          boxShadow: const [
            BoxShadow(
              color: Color(0x0D000000),
              blurRadius: 8,
              offset: Offset(0, 3),
            ),
          ],
        ),
        child: Icon(icon, color: color, size: 26),
      ),
      const SizedBox(height: 6),
      Text(label, style: const TextStyle(fontSize: 10, color: _ink)),
    ],
  );
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle({
    required this.title,
    required this.action,
    required this.onTap,
  });
  final String title;
  final String action;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Row(
    children: [
      Expanded(
        child: Text(
          title,
          style: const TextStyle(fontSize: 17, fontWeight: FontWeight.w700),
        ),
      ),
      TextButton(
        onPressed: onTap,
        child: Text(action, style: const TextStyle(color: _pink, fontSize: 12)),
      ),
    ],
  );
}

class _TransactionTile extends StatelessWidget {
  const _TransactionTile({
    required this.icon,
    required this.color,
    required this.title,
    required this.subtitle,
    required this.amount,
  });
  final IconData icon;
  final Color color;
  final String title, subtitle, amount;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 7),
    child: Row(
      children: [
        Container(
          width: 43,
          height: 43,
          decoration: BoxDecoration(
            color: color.withValues(alpha: .13),
            shape: BoxShape.circle,
          ),
          child: Icon(icon, color: color, size: 21),
        ),
        const SizedBox(width: 11),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: const TextStyle(
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                subtitle,
                style: const TextStyle(color: Colors.black54, fontSize: 11),
              ),
            ],
          ),
        ),
        Text(
          amount,
          style: const TextStyle(
            fontWeight: FontWeight.w700,
            fontSize: 12,
            color: _ink,
          ),
        ),
      ],
    ),
  );
}

class _DateHeader extends StatelessWidget {
  const _DateHeader({required this.title, required this.total});
  final String title, total;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 4),
    child: Row(
      children: [
        Expanded(
          child: Text(
            title,
            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w700),
          ),
        ),
        Text(
          total,
          style: const TextStyle(
            color: _pink,
            fontSize: 12,
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    ),
  );
}

class _PeriodChip extends StatelessWidget {
  const _PeriodChip({required this.label, this.active = false});
  final String label;
  final bool active;
  @override
  Widget build(BuildContext context) => Expanded(
    child: Padding(
      padding: const EdgeInsets.only(right: 6),
      child: Container(
        height: 34,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: active ? _pink : const Color(0xFFFFE9EE),
          borderRadius: BorderRadius.circular(18),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: active ? Colors.white : _ink,
            fontSize: 10,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    ),
  );
}

class _Panel extends StatelessWidget {
  const _Panel({required this.child});
  final Widget child;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(15),
    decoration: BoxDecoration(
      color: Colors.white,
      borderRadius: BorderRadius.circular(17),
      boxShadow: const [
        BoxShadow(
          color: Color(0x08000000),
          blurRadius: 10,
          offset: Offset(0, 3),
        ),
      ],
    ),
    child: child,
  );
}

class _BarChart extends StatelessWidget {
  const _BarChart();

  @override
  Widget build(BuildContext context) {
    const heights = [
      22.0,
      30.0,
      19.0,
      38.0,
      43.0,
      28.0,
      51.0,
      34.0,
      45.0,
      58.0,
      30.0,
      42.0,
      24.0,
      55.0,
      33.0,
      47.0,
      27.0,
      39.0,
      19.0,
    ];
    return SizedBox(
      height: 84,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: heights
            .asMap()
            .entries
            .map(
              (entry) => Container(
                width: 8,
                height: entry.value,
                decoration: BoxDecoration(
                  color: entry.key == 9 ? _pink : const Color(0xFFF5B3C2),
                  borderRadius: const BorderRadius.vertical(
                    top: Radius.circular(5),
                  ),
                ),
              ),
            )
            .toList(),
      ),
    );
  }
}

class _CategoryRow extends StatelessWidget {
  const _CategoryRow({
    required this.icon,
    required this.label,
    required this.amount,
    required this.percent,
    required this.color,
  });
  final IconData icon;
  final String label, amount, percent;
  final Color color;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: Row(
      children: [
        CircleAvatar(
          radius: 17,
          backgroundColor: color.withValues(alpha: .13),
          child: Icon(icon, color: color, size: 17),
        ),
        const SizedBox(width: 9),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
              Text(
                amount,
                style: const TextStyle(fontSize: 10, color: Colors.black54),
              ),
            ],
          ),
        ),
        Text(
          percent,
          style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600),
        ),
      ],
    ),
  );
}

class _ProfileItem extends StatelessWidget {
  const _ProfileItem({
    required this.icon,
    required this.label,
    required this.onTap,
  });
  final IconData icon;
  final String label;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => ListTile(
    dense: true,
    contentPadding: EdgeInsets.zero,
    leading: Icon(icon, size: 20, color: _ink),
    title: Text(label, style: const TextStyle(fontSize: 13)),
    trailing: const Icon(Icons.chevron_right, size: 18, color: Colors.black45),
    onTap: onTap,
  );
}

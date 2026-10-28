import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

const _pink = Color(0xFFEE6F8E);

class AddExpensePage extends StatefulWidget {
  const AddExpensePage({super.key});
  @override
  State<AddExpensePage> createState() => _AddExpensePageState();
}

class _AddExpensePageState extends State<AddExpensePage> {
  String _category = 'Ăn uống';
  bool _showSaved = false;
  final _amount = TextEditingController();
  final _name = TextEditingController();
  final _note = TextEditingController();

  @override
  void dispose() {
    _amount.dispose();
    _name.dispose();
    _note.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: const Color(0xFFFFF9FA),
    appBar: AppBar(
      backgroundColor: const Color(0xFFFFF9FA),
      leading: IconButton(
        onPressed: () => context.pop(),
        icon: const Icon(Icons.close),
      ),
      title: const Text(
        'Thêm chi tiêu',
        style: TextStyle(fontSize: 17, fontWeight: FontWeight.w700),
      ),
      centerTitle: true,
      actions: [
        TextButton(
          onPressed: _save,
          child: const Text(
            'Lưu',
            style: TextStyle(color: _pink, fontWeight: FontWeight.w700),
          ),
        ),
      ],
    ),
    body: SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(18, 8, 18, 28),
        children: [
          InkWell(
            onTap: () {},
            borderRadius: BorderRadius.circular(18),
            child: Container(
              height: 92,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                border: Border.all(color: const Color(0xFFFFDCE4)),
              ),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: const [
                  Icon(Icons.camera_alt_outlined, color: _pink, size: 27),
                  SizedBox(height: 6),
                  Text(
                    'Thêm ảnh món đồ (tùy chọn)',
                    style: TextStyle(fontSize: 12, color: Colors.black54),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 14),
          TextField(
            controller: _amount,
            keyboardType: TextInputType.number,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 26, fontWeight: FontWeight.w700),
            decoration: InputDecoration(
              hintText: '0 đ',
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide.none,
              ),
            ),
          ),
          const SizedBox(height: 18),
          const Text(
            'Danh mục',
            style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 10),
          Wrap(
            spacing: 9,
            runSpacing: 9,
            children: [
              _CategoryChip(
                label: 'Ăn uống',
                icon: Icons.restaurant,
                selected: _category == 'Ăn uống',
                onTap: () => setState(() => _category = 'Ăn uống'),
              ),
              _CategoryChip(
                label: 'Đồ uống',
                icon: Icons.local_cafe,
                selected: _category == 'Đồ uống',
                onTap: () => setState(() => _category = 'Đồ uống'),
              ),
              _CategoryChip(
                label: 'Mua sắm',
                icon: Icons.shopping_bag,
                selected: _category == 'Mua sắm',
                onTap: () => setState(() => _category = 'Mua sắm'),
              ),
              _CategoryChip(
                label: 'Di chuyển',
                icon: Icons.directions_car,
                selected: _category == 'Di chuyển',
                onTap: () => setState(() => _category = 'Di chuyển'),
              ),
              _CategoryChip(
                label: 'Hóa đơn',
                icon: Icons.receipt_long,
                selected: _category == 'Hóa đơn',
                onTap: () => setState(() => _category = 'Hóa đơn'),
              ),
              _CategoryChip(
                label: 'Giải trí',
                icon: Icons.movie,
                selected: _category == 'Giải trí',
                onTap: () => setState(() => _category = 'Giải trí'),
              ),
              _CategoryChip(
                label: 'Sức khỏe',
                icon: Icons.favorite,
                selected: _category == 'Sức khỏe',
                onTap: () => setState(() => _category = 'Sức khỏe'),
              ),
              _CategoryChip(
                label: 'Khác',
                icon: Icons.more_horiz,
                selected: _category == 'Khác',
                onTap: () => setState(() => _category = 'Khác'),
              ),
            ],
          ),
          const SizedBox(height: 20),
          _FormLabel(
            label: 'Tên giao dịch',
            child: TextField(
              controller: _name,
              decoration: _input('Ví dụ: Cà phê Highlands'),
            ),
          ),
          const SizedBox(height: 14),
          _FormLabel(
            label: 'Ngày',
            child: TextField(
              readOnly: true,
              controller: TextEditingController(text: '12/10/2024 14:30'),
              decoration: _input('Ngày giao dịch').copyWith(
                prefixIcon: const Icon(Icons.calendar_today_outlined, size: 18),
              ),
            ),
          ),
          const SizedBox(height: 14),
          _FormLabel(
            label: 'Ghi chú (tùy chọn)',
            child: TextField(
              controller: _note,
              decoration: _input('Thêm ghi chú...'),
            ),
          ),
          const SizedBox(height: 20),
          FilledButton(
            onPressed: _save,
            style: FilledButton.styleFrom(
              backgroundColor: _pink,
              minimumSize: const Size.fromHeight(50),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
            ),
            child: const Text(
              'Lưu chi tiêu',
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
          ),
          if (_showSaved)
            const Padding(
              padding: EdgeInsets.only(top: 12),
              child: Text(
                'Đã lưu bản xem trước (chưa kết nối API).',
                textAlign: TextAlign.center,
                style: TextStyle(color: _pink, fontSize: 12),
              ),
            ),
        ],
      ),
    ),
  );

  InputDecoration _input(String hint) => InputDecoration(
    hintText: hint,
    filled: true,
    fillColor: Colors.white,
    contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
    border: OutlineInputBorder(
      borderRadius: BorderRadius.circular(14),
      borderSide: BorderSide.none,
    ),
  );

  void _save() {
    setState(() => _showSaved = true);
  }
}

class _CategoryChip extends StatelessWidget {
  const _CategoryChip({
    required this.label,
    required this.icon,
    required this.selected,
    required this.onTap,
  });
  final String label;
  final IconData icon;
  final bool selected;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => InkWell(
    onTap: onTap,
    borderRadius: BorderRadius.circular(14),
    child: Container(
      width: 77,
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 5),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFE3EA) : Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: selected ? _pink : const Color(0xFFF3E8EB)),
      ),
      child: Column(
        children: [
          Icon(
            icon,
            size: 20,
            color: selected ? _pink : const Color(0xFF79737A),
          ),
          const SizedBox(height: 5),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              color: selected ? _pink : Colors.black87,
              fontWeight: selected ? FontWeight.w700 : FontWeight.normal,
            ),
          ),
        ],
      ),
    ),
  );
}

class _FormLabel extends StatelessWidget {
  const _FormLabel({required this.label, required this.child});
  final String label;
  final Widget child;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(
        label,
        style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600),
      ),
      const SizedBox(height: 7),
      child,
    ],
  );
}

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});
  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: const Color(0xFFFFF9FA),
    appBar: AppBar(
      backgroundColor: const Color(0xFFFFF9FA),
      leading: IconButton(
        onPressed: () => context.pop(),
        icon: const Icon(Icons.chevron_left),
      ),
      title: const Text(
        'Cài đặt',
        style: TextStyle(fontSize: 17, fontWeight: FontWeight.w700),
      ),
      centerTitle: true,
    ),
    body: ListView(
      padding: const EdgeInsets.all(18),
      children: const [
        _SettingsGroup(
          title: 'Giao diện',
          items: [
            _SettingRow(
              icon: Icons.dark_mode_outlined,
              label: 'Chế độ tối',
              trailing: Switch(value: false, onChanged: null),
            ),
            _SettingRow(
              icon: Icons.palette_outlined,
              label: 'Màu chủ đạo',
              detail: 'Hồng pastel',
            ),
          ],
        ),
        _SettingsGroup(
          title: 'Thông báo',
          items: [
            _SettingRow(
              icon: Icons.notifications_none,
              label: 'Nhắc nhở ngân sách',
              trailing: Switch(value: true, onChanged: null),
            ),
            _SettingRow(
              icon: Icons.receipt_long_outlined,
              label: 'Thông báo giao dịch',
              trailing: Switch(value: true, onChanged: null),
            ),
            _SettingRow(
              icon: Icons.summarize_outlined,
              label: 'Tóm tắt chi tiêu hàng tuần',
              trailing: Switch(value: false, onChanged: null),
            ),
          ],
        ),
        _SettingsGroup(
          title: 'Bảo mật',
          items: [
            _SettingRow(
              icon: Icons.face_retouching_natural,
              label: 'Đăng nhập bằng Face ID',
              trailing: Switch(value: true, onChanged: null),
            ),
            _SettingRow(icon: Icons.password_outlined, label: 'Đổi mật khẩu'),
            _SettingRow(
              icon: Icons.language,
              label: 'Ngôn ngữ',
              detail: 'Tiếng Việt',
            ),
          ],
        ),
      ],
    ),
  );
}

class _SettingsGroup extends StatelessWidget {
  const _SettingsGroup({required this.title, required this.items});
  final String title;
  final List<Widget> items;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 19),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 2, bottom: 8),
          child: Text(
            title,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700),
          ),
        ),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 4),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(17),
          ),
          child: Column(children: items),
        ),
      ],
    ),
  );
}

class _SettingRow extends StatelessWidget {
  const _SettingRow({
    required this.icon,
    required this.label,
    this.detail,
    this.trailing,
  });
  final IconData icon;
  final String label;
  final String? detail;
  final Widget? trailing;
  @override
  Widget build(BuildContext context) => ListTile(
    contentPadding: EdgeInsets.zero,
    leading: Icon(icon, size: 20),
    title: Text(label, style: const TextStyle(fontSize: 12)),
    trailing:
        trailing ??
        (detail != null
            ? Text(
                detail!,
                style: const TextStyle(fontSize: 11, color: Colors.black54),
              )
            : const Icon(Icons.chevron_right, size: 18, color: Colors.black45)),
  );
}

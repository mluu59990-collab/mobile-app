import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/providers/app_providers.dart';

const _pink = Color(0xFFEE6F8E);

class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});

  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _confirmPassword = TextEditingController();
  bool _hidePassword = true;
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _confirmPassword.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: const Color(0xFFFFF7F8),
    appBar: AppBar(
      backgroundColor: const Color(0xFFFFF7F8),
      leading: IconButton(
        onPressed: () => context.go('/login'),
        icon: const Icon(Icons.chevron_left),
      ),
    ),
    body: SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(24, 18, 24, 24),
        children: [
          Text(
            'Tạo tài khoản',
            style: Theme.of(context).textTheme.headlineSmall
                ?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 5),
          const Text(
            'Bắt đầu quản lý chi tiêu của bạn',
            style: TextStyle(color: Color(0xFF706B73)),
          ),
          const SizedBox(height: 27),
          TextField(
            controller: _email,
            keyboardType: TextInputType.emailAddress,
            autofillHints: const [AutofillHints.email],
            textInputAction: TextInputAction.next,
            decoration: _decoration('Email', Icons.mail_outline),
          ),
          const SizedBox(height: 10),
          TextField(
            controller: _password,
            obscureText: _hidePassword,
            autofillHints: const [AutofillHints.newPassword],
            textInputAction: TextInputAction.next,
            decoration:
                _decoration(
                  'Mật khẩu (tối thiểu 8 ký tự)',
                  Icons.lock_outline,
                ).copyWith(
                  suffixIcon: IconButton(
                    onPressed: () =>
                        setState(() => _hidePassword = !_hidePassword),
                    icon: Icon(
                      _hidePassword
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined,
                      size: 20,
                    ),
                  ),
                ),
          ),
          const SizedBox(height: 10),
          TextField(
            controller: _confirmPassword,
            obscureText: _hidePassword,
            onSubmitted: (_) => _register(),
            decoration: _decoration('Nhập lại mật khẩu', Icons.lock_outline),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(
              _error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ],
          const SizedBox(height: 20),
          SizedBox(
            height: 49,
            child: FilledButton(
              onPressed: _submitting ? null : _register,
              style: FilledButton.styleFrom(
                backgroundColor: _pink,
                shape: const StadiumBorder(),
              ),
              child: _submitting
                  ? const SizedBox.square(
                      dimension: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Text(
                      'Đăng ký',
                      style: TextStyle(fontWeight: FontWeight.w700),
                    ),
            ),
          ),
          const SizedBox(height: 20),
          Center(
            child: Wrap(
              children: [
                const Text(
                  'Đã có tài khoản? ',
                  style: TextStyle(fontSize: 12, color: Colors.black54),
                ),
                GestureDetector(
                  onTap: () => context.go('/login'),
                  child: const Text(
                    'Đăng nhập',
                    style: TextStyle(
                      color: _pink,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    ),
  );

  InputDecoration _decoration(String hint, IconData icon) => InputDecoration(
    hintText: hint,
    prefixIcon: Icon(icon, size: 19),
    filled: true,
    fillColor: Colors.white,
    contentPadding: const EdgeInsets.symmetric(vertical: 16),
    border: OutlineInputBorder(
      borderRadius: BorderRadius.circular(14),
      borderSide: const BorderSide(color: Color(0xFFF1E6E9)),
    ),
    enabledBorder: OutlineInputBorder(
      borderRadius: BorderRadius.circular(14),
      borderSide: const BorderSide(color: Color(0xFFF1E6E9)),
    ),
  );

  Future<void> _register() async {
    if (_password.text != _confirmPassword.text) {
      setState(() => _error = 'Mật khẩu nhập lại chưa khớp.');
      return;
    }
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await ref
          .read(authRepositoryProvider)
          .register(email: _email.text, password: _password.text);
      if (mounted) context.go('/home');
    } on ApiException catch (error) {
      if (mounted) setState(() => _error = error.message);
    } catch (_) {
      if (mounted) setState(() => _error = 'Không thể tạo tài khoản lúc này.');
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }
}

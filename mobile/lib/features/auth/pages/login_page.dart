import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/providers/app_providers.dart';

const _pink = Color(0xFFEE6F8E);

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  bool _remember = true;
  bool _hidePassword = true;
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: const Color(0xFFFFF7F8),
    appBar: AppBar(
      backgroundColor: const Color(0xFFFFF7F8),
      automaticallyImplyLeading: false,
    ),
    body: SafeArea(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(24, 18, 24, 24),
        children: [
          Text(
            'Đăng nhập',
            style: Theme.of(context).textTheme.headlineSmall
                ?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 5),
          const Text(
            'Chào mừng bạn trở lại 👋',
            style: TextStyle(color: Color(0xFF706B73)),
          ),
          const SizedBox(height: 29),
          TextField(
            controller: _email,
            keyboardType: TextInputType.emailAddress,
            autofillHints: const [AutofillHints.username, AutofillHints.email],
            textInputAction: TextInputAction.next,
            decoration: _decoration('Email', Icons.mail_outline),
          ),
          const SizedBox(height: 10),
          TextField(
            controller: _password,
            obscureText: _hidePassword,
            autofillHints: const [AutofillHints.password],
            onSubmitted: (_) => _login(),
            decoration: _decoration('Mật khẩu', Icons.lock_outline).copyWith(
              suffixIcon: IconButton(
                onPressed: () => setState(() => _hidePassword = !_hidePassword),
                icon: Icon(
                  _hidePassword
                      ? Icons.visibility_off_outlined
                      : Icons.visibility_outlined,
                  size: 20,
                ),
              ),
            ),
          ),
          const SizedBox(height: 5),
          Row(
            children: [
              SizedBox(
                width: 30,
                child: Checkbox(
                  value: _remember,
                  activeColor: _pink,
                  onChanged: (value) =>
                      setState(() => _remember = value ?? false),
                ),
              ),
              const Text('Ghi nhớ đăng nhập', style: TextStyle(fontSize: 11)),
              const Spacer(),
              TextButton(
                onPressed: _notReady,
                child: const Text(
                  'Quên mật khẩu?',
                  style: TextStyle(color: _pink, fontSize: 11),
                ),
              ),
            ],
          ),
          if (_error case final error?) ...[
            const SizedBox(height: 8),
            Text(
              error,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ],
          const SizedBox(height: 15),
          SizedBox(
            height: 49,
            child: FilledButton(
              onPressed: _submitting ? null : _login,
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
                      'Đăng nhập',
                      style: TextStyle(fontWeight: FontWeight.w700),
                    ),
            ),
          ),
          const SizedBox(height: 22),
          const Row(
            children: [
              Expanded(child: Divider()),
              Padding(
                padding: EdgeInsets.symmetric(horizontal: 12),
                child: Text(
                  'Hoặc đăng nhập với',
                  style: TextStyle(fontSize: 11, color: Colors.black54),
                ),
              ),
              Expanded(child: Divider()),
            ],
          ),
          const SizedBox(height: 14),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _SocialButton(
                icon: Icons.apple,
                label: 'Apple',
                onTap: _notReady,
              ),
              const SizedBox(width: 12),
              _SocialButton(
                icon: Icons.g_mobiledata_rounded,
                label: 'Google',
                onTap: _notReady,
              ),
            ],
          ),
          const SizedBox(height: 25),
          Center(
            child: Wrap(
              alignment: WrapAlignment.center,
              children: [
                const Text(
                  'Chưa có tài khoản? ',
                  style: TextStyle(fontSize: 11, color: Colors.black54),
                ),
                GestureDetector(
                  onTap: () => context.push('/register'),
                  child: const Text(
                    'Đăng ký',
                    style: TextStyle(
                      fontSize: 11,
                      color: _pink,
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

  Future<void> _login() async {
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await ref
          .read(authRepositoryProvider)
          .login(email: _email.text, password: _password.text);
      if (mounted) context.go('/home');
    } on ApiException catch (error) {
      if (mounted) setState(() => _error = error.message);
    } catch (_) {
      if (mounted) setState(() => _error = 'Không thể đăng nhập lúc này.');
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _notReady() => ScaffoldMessenger.of(context).showSnackBar(
    const SnackBar(content: Text('Tính năng này chưa được kết nối.')),
  );
}

class _SocialButton extends StatelessWidget {
  const _SocialButton({
    required this.icon,
    required this.label,
    required this.onTap,
  });
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => OutlinedButton.icon(
    onPressed: onTap,
    icon: Icon(
      icon,
      color: label == 'Apple' ? Colors.black : const Color(0xFF4285F4),
    ),
    label: Text(label, style: const TextStyle(color: Colors.black87)),
    style: OutlinedButton.styleFrom(
      backgroundColor: Colors.white,
      side: const BorderSide(color: Color(0xFFF0E5E8)),
      minimumSize: const Size(125, 48),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
    ),
  );
}

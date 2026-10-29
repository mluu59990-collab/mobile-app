import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/formatters/local_datetime_formatter.dart';
import '../data/expense.dart';
import '../providers/expense_providers.dart';

const _pink = Color(0xFFEE6F8E);

class AddExpensePage extends ConsumerStatefulWidget {
  const AddExpensePage({super.key, this.expenseId});
  final String? expenseId;

  @override
  ConsumerState<AddExpensePage> createState() => _AddExpensePageState();
}

class _AddExpensePageState extends ConsumerState<AddExpensePage> {
  final _amount = TextEditingController();
  final _note = TextEditingController();
  String _category = 'food';
  DateTime _spentAt = DateTime.now();
  bool _saving = false;
  bool _initialized = false;
  String? _error;

  @override
  void dispose() {
    _amount.dispose();
    _note.dispose();
    super.dispose();
  }

  void _fill(Expense expense) {
    if (_initialized) return;
    _initialized = true;
    _amount.text = expense.amount.toString();
    _note.text = expense.note ?? '';
    _category = expense.category ?? 'other';
    _spentAt = expense.spentAt.toLocal();
  }

  @override
  Widget build(BuildContext context) {
    final edit = widget.expenseId != null;
    if (!edit) _initialized = true;
    final expenseAsync = edit
        ? ref.watch(expenseByIdProvider(widget.expenseId!))
        : null;
    expenseAsync?.whenData(_fill);
    return Scaffold(
      backgroundColor: const Color(0xFFFFF9FA),
      appBar: AppBar(
        title: Text(edit ? 'Sửa chi tiêu' : 'Thêm chi tiêu'),
        centerTitle: true,
        actions: [
          TextButton(
            onPressed: _saving ? null : _save,
            child: const Text('Lưu'),
          ),
        ],
      ),
      body: edit && (expenseAsync?.isLoading ?? false)
          ? const Center(child: CircularProgressIndicator())
          : edit && (expenseAsync?.hasError ?? false)
          ? _loadError(
              expenseAsync?.error.toString() ?? 'Không tải được chi tiêu.',
            )
          : SafeArea(
              child: ListView(
                padding: const EdgeInsets.fromLTRB(18, 8, 18, 28),
                children: [
                  TextField(
                    controller: _amount,
                    keyboardType: TextInputType.number,
                    textAlign: TextAlign.center,
                    decoration: const InputDecoration(
                      labelText: 'Số tiền (VND)',
                      suffixText: '₫',
                    ),
                  ),
                  const SizedBox(height: 20),
                  const Text('Danh mục'),
                  Wrap(
                    spacing: 8,
                    children: expenseCategories
                        .map(
                          (category) => ChoiceChip(
                            label: Text(category.label),
                            avatar: Icon(category.icon, size: 18),
                            selected: _category == category.key,
                            onSelected: (_) =>
                                setState(() => _category = category.key),
                          ),
                        )
                        .toList(),
                  ),
                  TextField(
                    controller: _note,
                    decoration: const InputDecoration(
                      labelText: 'Tên giao dịch',
                      hintText: 'Ví dụ: Cà phê',
                    ),
                  ),
                  const SizedBox(height: 14),
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: const Text('Ngày và giờ'),
                    subtitle: Text(formatLocalDateTime(_spentAt)),
                    trailing: const Icon(Icons.calendar_today_outlined),
                    onTap: _pickDateTime,
                  ),
                  if (_error != null)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      child: Text(
                        _error!,
                        style: const TextStyle(color: Colors.red),
                      ),
                    ),
                  const SizedBox(height: 12),
                  FilledButton(
                    onPressed: _saving ? null : _save,
                    style: FilledButton.styleFrom(
                      backgroundColor: _pink,
                      minimumSize: const Size.fromHeight(50),
                    ),
                    child: Text(_saving ? 'Đang lưu…' : 'Lưu chi tiêu'),
                  ),
                ],
              ),
            ),
    );
  }

  Widget _loadError(String error) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(error),
        TextButton(
          onPressed: () =>
              ref.invalidate(expenseByIdProvider(widget.expenseId!)),
          child: const Text('Thử lại'),
        ),
      ],
    ),
  );

  Future<void> _pickDateTime() async {
    final date = await showDatePicker(
      context: context,
      initialDate: _spentAt,
      firstDate: DateTime(2000),
      lastDate: DateTime.now().add(const Duration(days: 1)),
    );
    if (date == null || !mounted) return;
    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(_spentAt),
    );
    if (time == null) return;
    setState(
      () => _spentAt = DateTime(
        date.year,
        date.month,
        date.day,
        time.hour,
        time.minute,
      ),
    );
  }

  Future<void> _save() async {
    final amount = int.tryParse(
      _amount.text.trim().replaceAll('.', '').replaceAll(',', ''),
    );
    if (amount == null || amount <= 0) {
      setState(() => _error = 'Vui lòng nhập số tiền VND hợp lệ.');
      return;
    }
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      final input = ExpenseInput(
        amount: amount,
        note: _note.text.trim().isEmpty ? null : _note.text.trim(),
        category: _category,
        spentAt: _spentAt.toUtc(),
      );
      final actions = ref.read(expenseActionsProvider);
      if (widget.expenseId == null) {
        await actions.create(input);
      } else {
        await actions.update(widget.expenseId!, input);
      }
      if (mounted) context.pop();
    } catch (error) {
      if (mounted) setState(() => _error = error.toString());
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }
}
